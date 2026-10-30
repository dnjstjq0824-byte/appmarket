package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/domain"
)

type AppealService struct {
	mu      sync.RWMutex
	appeals map[string]*domain.DeveloperAppeal
}

func NewAppealService() *AppealService {
	return &AppealService{
		appeals: make(map[string]*domain.DeveloperAppeal),
	}
}

func (s *AppealService) SubmitAppeal(ctx context.Context, orgID, releaseID, reason, originalReviewerID string) (*domain.DeveloperAppeal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	appealID := fmt.Sprintf("apl-%d", time.Now().UnixNano())
	appeal := &domain.DeveloperAppeal{
		ID:                 appealID,
		OrgID:              orgID,
		ReleaseID:          releaseID,
		Reason:             reason,
		OriginalReviewerID: originalReviewerID,
		Status:             domain.AppealPending,
		CreatedAt:          time.Now().UTC(),
	}

	s.appeals[appealID] = appeal
	return appeal, nil
}

// DecideAppeal enforces Dual-Control: Original reviewer cannot judge their own appeal (TC-FUNC-013)
func (s *AppealService) DecideAppeal(ctx context.Context, appealID, reviewerID string, decision domain.AppealStatus, notes string) (*domain.DeveloperAppeal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	appeal, exists := s.appeals[appealID]
	if !exists {
		return nil, domain.ErrAppealNotFound
	}

	// Dual-Control Enforcement: Original reviewer cannot review appeal of their own rejection
	if reviewerID == appeal.OriginalReviewerID {
		return nil, domain.ErrAppealSameReviewerForbidden
	}

	now := time.Now().UTC()
	appeal.AssignedReviewerID = reviewerID
	appeal.Status = decision
	appeal.DecisionNotes = notes
	appeal.DecidedAt = &now

	return appeal, nil
}
