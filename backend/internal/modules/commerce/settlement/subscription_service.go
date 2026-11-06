package settlement

import (
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
)

type SubscriptionService struct {
	mu   sync.RWMutex
	subs map[string]*domain.Subscription
}

func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{subs: make(map[string]*domain.Subscription)}
}

// CreateSubscription starts a recurring billing relationship. The first period
// starts immediately; renewal charging is driven by PG billing keys.
func (s *SubscriptionService) CreateSubscription(userID, appID, orgID string, interval domain.BillingInterval, amountMinor int64, currency domain.Currency) (*domain.Subscription, error) {
	if interval != domain.BillingMonthly && interval != domain.BillingYearly {
		return nil, domain.ErrInvalidBillingInterval
	}
	if amountMinor <= 0 {
		return nil, domain.ErrInvalidSubscriptionAmount
	}
	if currency == "" {
		currency = domain.CurrencyKRW
	}
	now := time.Now().UTC()
	end := now.AddDate(0, 1, 0)
	if interval == domain.BillingYearly {
		end = now.AddDate(1, 0, 0)
	}
	sub := &domain.Subscription{
		ID:                 fmt.Sprintf("sub-%d", now.UnixNano()),
		UserID:             userID,
		AppID:              appID,
		OrgID:              orgID,
		BillingInterval:    interval,
		AmountMinor:        amountMinor,
		Currency:           currency,
		Status:             domain.SubscriptionActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   end,
		CreatedAt:          now,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs[sub.ID] = sub
	return sub, nil
}

// CancelSubscription ends billing at the current period end. Only the owning
// user may cancel.
func (s *SubscriptionService) CancelSubscription(subID, userID string) (*domain.Subscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub, ok := s.subs[subID]
	if !ok {
		return nil, domain.ErrSubscriptionNotFound
	}
	if sub.UserID != userID {
		return nil, domain.ErrSubscriptionAccessDenied
	}
	if sub.Status != domain.SubscriptionActive {
		return nil, domain.ErrSubscriptionAlreadyEnded
	}
	now := time.Now().UTC()
	sub.Status = domain.SubscriptionCancelled
	sub.CancelledAt = &now
	return sub, nil
}

func (s *SubscriptionService) Get(subID string) *domain.Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.subs[subID]
}

func (s *SubscriptionService) ListByUser(userID string) []domain.Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Subscription
	for _, sub := range s.subs {
		if sub.UserID == userID {
			out = append(out, *sub)
		}
	}
	return out
}
