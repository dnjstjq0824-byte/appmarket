package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/ads/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/sanitize"
)

// CampaignService manages advertiser campaigns through review into the live
// GSP auction. New campaigns start PENDING_REVIEW; only ACTIVE campaigns are
// added to the auction engine.
type CampaignService struct {
	mu       sync.RWMutex
	campaigns map[string]*domain.AdCampaign
	engine   *AuctionEngine
}

func NewCampaignService(engine *AuctionEngine) *CampaignService {
	if engine == nil {
		engine = NewAuctionEngine(50)
	}
	return &CampaignService{campaigns: make(map[string]*domain.AdCampaign), engine: engine}
}

type CreateCampaignInput struct {
	OrgID          string
	AppID          string
	AppName        string
	TargetKeywords []string
	MaxBidCPC      int64
	DailyBudget    int64
}

func (s *CampaignService) CreateCampaign(_ context.Context, in CreateCampaignInput) (*domain.AdCampaign, error) {
	if in.OrgID == "" || in.AppID == "" || len(in.TargetKeywords) == 0 {
		return nil, fmt.Errorf("%w: org, app and keywords are required", domain.ErrInvalidCampaign)
	}
	if in.MaxBidCPC <= 0 || in.DailyBudget <= 0 {
		return nil, fmt.Errorf("%w: bid and budget must be positive", domain.ErrInvalidCampaign)
	}
	now := time.Now().UTC()
	c := &domain.AdCampaign{
		ID:             fmt.Sprintf("camp-%d", now.UnixNano()),
		OrgID:          in.OrgID,
		AppID:          in.AppID,
		AppName:        sanitize.Strict(in.AppName),
		TargetKeywords: sanitizeKeywords(in.TargetKeywords),
		MaxBidCPC:      in.MaxBidCPC,
		QualityScore:   5.0,
		DailyBudget:    in.DailyBudget,
		Status:         domain.CampaignPendingReview,
		CreatedAt:      now,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.campaigns[c.ID] = c
	cp := *c
	return &cp, nil
}

func sanitizeKeywords(in []string) []string {
	out := make([]string, 0, len(in))
	for _, k := range in {
		if v := sanitize.Strict(k); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func (s *CampaignService) GetCampaign(_ context.Context, id string) (*domain.AdCampaign, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.campaigns[id]
	if !ok {
		return nil, domain.ErrAdCampaignNotFound
	}
	cp := *c
	return &cp, nil
}

func (s *CampaignService) ListCampaignsByOrg(_ context.Context, orgID string) []domain.AdCampaign {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.AdCampaign
	for _, c := range s.campaigns {
		if c.OrgID == orgID {
			out = append(out, *c)
		}
	}
	return out
}

func (s *CampaignService) ListCampaignsByStatus(_ context.Context, status domain.CampaignStatus) []domain.AdCampaign {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.AdCampaign
	for _, c := range s.campaigns {
		if status == "" || c.Status == status {
			out = append(out, *c)
		}
	}
	return out
}

// ReviewCampaign approves (activates into the auction) or rejects a pending
// campaign. Only the owning org's campaigns are visible to developers; the
// admin console reviews across orgs.
func (s *CampaignService) ReviewCampaign(_ context.Context, id string, approve bool) (*domain.AdCampaign, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.campaigns[id]
	if !ok {
		return nil, domain.ErrAdCampaignNotFound
	}
	if c.Status != domain.CampaignPendingReview {
		return nil, domain.ErrCampaignNotReviewable
	}
	if approve {
		c.Status = domain.CampaignActive
		s.engine.AddCampaign(*c)
	} else {
		c.Status = domain.CampaignRejected
	}
	cp := *c
	return &cp, nil
}

// SetPaused flips an active campaign between ACTIVE and PAUSED_MANUAL.
func (s *CampaignService) SetPaused(_ context.Context, orgID, id string, paused bool) (*domain.AdCampaign, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.campaigns[id]
	if !ok {
		return nil, domain.ErrAdCampaignNotFound
	}
	if c.OrgID != orgID {
		return nil, domain.ErrCampaignAccessDenied
	}
	if paused && c.Status == domain.CampaignActive {
		c.Status = domain.CampaignPausedManual
	} else if !paused && c.Status == domain.CampaignPausedManual {
		c.Status = domain.CampaignActive
		s.engine.AddCampaign(*c)
	}
	cp := *c
	return &cp, nil
}
