package notification

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is OPEN: external messaging service unavailable")
	ErrNotificationNotFound = errors.New("notification not found")
	ErrUnauthorizedAccess = errors.New("unauthorized access to notification")
)

type Category string

const (
	CategorySecurity  Category = "SECURITY"
	CategoryBilling   Category = "BILLING"
	CategoryReview    Category = "REVIEW"
	CategorySystem    Category = "SYSTEM"
	CategoryPromotion Category = "PROMOTION"
)

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Category  Category  `json:"category"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	ActionURL string    `json:"action_url,omitempty"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type DeliveryResult struct {
	Notification *Notification `json:"notification"`
	InAppSaved   bool          `json:"in_app_saved"`
	ExternalSent bool          `json:"external_sent"`
	ExternalError string       `json:"external_error,omitempty"`
	CircuitState string        `json:"circuit_state"`
}

type CircuitState string

const (
	StateClosed   CircuitState = "CLOSED"
	StateOpen     CircuitState = "OPEN"
	StateHalfOpen CircuitState = "HALF_OPEN"
)

type CircuitBreaker struct {
	mu               sync.Mutex
	state            CircuitState
	failureThreshold int
	consecutiveFails int
	openDuration     time.Duration
	lastStateChange  time.Time
}

func NewCircuitBreaker(failureThreshold int, openDuration time.Duration) *CircuitBreaker {
	if failureThreshold <= 0 {
		failureThreshold = 3
	}
	if openDuration <= 0 {
		openDuration = 10 * time.Second
	}
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		openDuration:     openDuration,
		lastStateChange:  time.Now(),
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	now := time.Now()

	switch cb.state {
	case StateOpen:
		if now.Sub(cb.lastStateChange) > cb.openDuration {
			cb.state = StateHalfOpen
			cb.lastStateChange = now
		} else {
			cb.mu.Unlock()
			return ErrCircuitBreakerOpen
		}
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.consecutiveFails++
		if cb.consecutiveFails >= cb.failureThreshold {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}
		return err
	}

	// Success
	cb.consecutiveFails = 0
	cb.state = StateClosed
	return nil
}

func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == StateOpen && time.Since(cb.lastStateChange) > cb.openDuration {
		return StateHalfOpen
	}
	return cb.state
}

type ExternalMessagingClient interface {
	SendEmail(ctx context.Context, recipient, subject, body string) error
	SendPush(ctx context.Context, deviceToken, title, body string) error
}

type NotificationRepository interface {
	Save(ctx context.Context, n *Notification) error
	FindByID(ctx context.Context, id string) (*Notification, error)
	FindByUserID(ctx context.Context, userID string, unreadOnly bool) ([]Notification, error)
	MarkAsRead(ctx context.Context, id, userID string) error
}

type MemoryNotificationRepository struct {
	mu    sync.RWMutex
	items map[string]*Notification
}

func NewMemoryNotificationRepository() *MemoryNotificationRepository {
	return &MemoryNotificationRepository{
		items: make(map[string]*Notification),
	}
}

func (r *MemoryNotificationRepository) Save(ctx context.Context, n *Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *n
	r.items[n.ID] = &cp
	return nil
}

func (r *MemoryNotificationRepository) FindByID(ctx context.Context, id string) (*Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, exists := r.items[id]
	if !exists {
		return nil, ErrNotificationNotFound
	}
	cp := *n
	return &cp, nil
}

func (r *MemoryNotificationRepository) FindByUserID(ctx context.Context, userID string, unreadOnly bool) ([]Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []Notification
	for _, n := range r.items {
		if n.UserID == userID {
			if unreadOnly && n.IsRead {
				continue
			}
			res = append(res, *n)
		}
	}
	return res, nil
}

func (r *MemoryNotificationRepository) MarkAsRead(ctx context.Context, id, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, exists := r.items[id]
	if !exists {
		return ErrNotificationNotFound
	}
	if n.UserID != userID {
		return ErrUnauthorizedAccess
	}
	n.IsRead = true
	return nil
}

type NotificationService struct {
	repo           NotificationRepository
	externalClient ExternalMessagingClient
	circuitBreaker *CircuitBreaker
	retryQueue     []Notification
	queueMu        sync.Mutex
}

func NewNotificationService(repo NotificationRepository, ext ExternalMessagingClient, cb *CircuitBreaker) *NotificationService {
	if cb == nil {
		cb = NewCircuitBreaker(3, 5*time.Second)
	}
	return &NotificationService{
		repo:           repo,
		externalClient: ext,
		circuitBreaker: cb,
		retryQueue:     make([]Notification, 0),
	}
}

func (s *NotificationService) Send(ctx context.Context, n Notification, recipientEmail, pushToken string) (*DeliveryResult, error) {
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}

	// 1. Guaranteed in-app persistence first
	if err := s.repo.Save(ctx, &n); err != nil {
		return nil, err
	}

	result := &DeliveryResult{
		Notification: &n,
		InAppSaved:   true,
		CircuitState: string(s.circuitBreaker.GetState()),
	}

	// 2. If no external channel requested, finish
	if recipientEmail == "" && pushToken == "" {
		result.ExternalSent = true
		return result, nil
	}

	// 3. Protected external delivery via circuit breaker
	err := s.circuitBreaker.Execute(func() error {
		if recipientEmail != "" && s.externalClient != nil {
			if emailErr := s.externalClient.SendEmail(ctx, recipientEmail, n.Title, n.Body); emailErr != nil {
				return emailErr
			}
		}
		if pushToken != "" && s.externalClient != nil {
			if pushErr := s.externalClient.SendPush(ctx, pushToken, n.Title, n.Body); pushErr != nil {
				return pushErr
			}
		}
		return nil
	})

	result.CircuitState = string(s.circuitBreaker.GetState())

	if err != nil {
		// External delivery failed, but in-app notification is NOT rolled back
		result.ExternalSent = false
		result.ExternalError = err.Error()

		// Enqueue for background retry
		s.queueMu.Lock()
		s.retryQueue = append(s.retryQueue, n)
		s.queueMu.Unlock()
	} else {
		result.ExternalSent = true
	}

	return result, nil
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userID string, unreadOnly bool) ([]Notification, error) {
	return s.repo.FindByUserID(ctx, userID, unreadOnly)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, id, userID string) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}

func (s *NotificationService) RetryPendingQueue(ctx context.Context, email, token string) int {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()

	var remaining []Notification
	successCount := 0

	for _, item := range s.retryQueue {
		err := s.circuitBreaker.Execute(func() error {
			if email != "" && s.externalClient != nil {
				if err := s.externalClient.SendEmail(ctx, email, item.Title, item.Body); err != nil {
					return err
				}
			}
			return nil
		})
		if err == nil {
			successCount++
		} else {
			remaining = append(remaining, item)
		}
	}
	s.retryQueue = remaining
	return successCount
}

