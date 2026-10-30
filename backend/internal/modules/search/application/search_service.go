package application

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/search/domain"
)

type CircuitBreakerState string

const (
	BreakerClosed   CircuitBreakerState = "CLOSED"
	BreakerOpen     CircuitBreakerState = "OPEN"
	BreakerHalfOpen CircuitBreakerState = "HALF_OPEN"
)

type CircuitBreaker struct {
	state          CircuitBreakerState
	failureCount   int
	failureThreshold int
	resetTimeout   time.Duration
	lastFailure    time.Time
	mu             sync.Mutex
}

func NewCircuitBreaker(threshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            BreakerClosed,
		failureThreshold: threshold,
		resetTimeout:     resetTimeout,
	}
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == BreakerClosed {
		return true
	}

	if cb.state == BreakerOpen {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = BreakerHalfOpen
			return true
		}
		return false
	}

	// Half-open: allow canary trial
	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.state = BreakerClosed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount++
	cb.lastFailure = time.Now()
	if cb.failureCount >= cb.failureThreshold {
		cb.state = BreakerOpen
	}
}

func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

type SearchService struct {
	primaryClient  FallbackSearchEngine // Represents OpenSearch client
	fallbackClient FallbackSearchEngine // Represents PostgreSQL pg_trgm client
	breaker        *CircuitBreaker
	isPrimaryDown  bool
	mu             sync.RWMutex
}

func NewSearchService(primary FallbackSearchEngine, fallback FallbackSearchEngine) *SearchService {
	return &SearchService{
		primaryClient:  primary,
		fallbackClient: fallback,
		breaker:        NewCircuitBreaker(3, 10*time.Second),
	}
}

func (s *SearchService) SetPrimarySimulatedDown(down bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isPrimaryDown = down
}

func (s *SearchService) IndexDocument(ctx context.Context, doc domain.SearchDoc) error {
	// Dual write to primary and fallback store
	if s.primaryClient != nil {
		_ = s.primaryClient.IndexDocument(ctx, doc)
	}
	if s.fallbackClient != nil {
		return s.fallbackClient.IndexDocument(ctx, doc)
	}
	return nil
}

func (s *SearchService) Search(ctx context.Context, q domain.SearchQuery) (*domain.SearchResult, error) {
	s.mu.RLock()
	simulatedDown := s.isPrimaryDown
	s.mu.RUnlock()

	// If breaker allows and not simulated down, try primary OpenSearch
	if s.breaker.AllowRequest() && !simulatedDown && s.primaryClient != nil {
		res, err := s.primaryClient.Search(ctx, q)
		if err == nil {
			s.breaker.RecordSuccess()
			res.FallbackUsed = false
			return res, nil
		}
		// Primary failed: record failure into circuit breaker
		s.breaker.RecordFailure()
	}

	// Transparently execute fallback without breaking caller
	if s.fallbackClient != nil {
		fallbackRes, err := s.fallbackClient.Search(ctx, q)
		if err != nil {
			return nil, err
		}
		fallbackRes.FallbackUsed = true
		return fallbackRes, nil
	}

	return nil, domain.ErrSearchEngineUnavailable
}

func (s *SearchService) Suggest(ctx context.Context, prefix, locale string, limit int) (*domain.SuggestResult, error) {
	s.mu.RLock()
	simulatedDown := s.isPrimaryDown
	s.mu.RUnlock()

	if s.breaker.AllowRequest() && !simulatedDown && s.primaryClient != nil {
		res, err := s.primaryClient.Suggest(ctx, prefix, locale, limit)
		if err == nil {
			return res, nil
		}
	}

	if s.fallbackClient != nil {
		return s.fallbackClient.Suggest(ctx, prefix, locale, limit)
	}

	return nil, errors.New("suggest service unavailable")
}

