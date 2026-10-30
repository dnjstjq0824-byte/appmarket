package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/ads/domain"
)

type ClickFraudFilter struct {
	mu           sync.Mutex
	recentClicks map[string]time.Time // key: deviceHash + ":" + campaignID -> lastChargedAt
	engine       *AuctionEngine
}

func NewClickFraudFilter(engine *AuctionEngine) *ClickFraudFilter {
	return &ClickFraudFilter{
		recentClicks: make(map[string]time.Time),
		engine:       engine,
	}
}

// ProcessAdClick verifies impression token, checks fraud filters, and deducts cost up to daily budget
func (f *ClickFraudFilter) ProcessAdClick(ctx context.Context, token domain.ImpressionToken, clickTime time.Time) (*domain.ClickResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	campaign := f.engine.GetCampaign(token.CampaignID)
	if campaign == nil {
		return nil, domain.ErrAdCampaignNotFound
	}

	clickID := fmt.Sprintf("clk-%d", clickTime.UnixNano())

	// 1. Impression Token Expiration Check (> 5 minutes)
	if clickTime.After(token.ExpiresAt) {
		return &domain.ClickResult{
			ClickID:       clickID,
			CampaignID:    token.CampaignID,
			ChargedAmount: 0,
			IsFraud:       true,
			FraudReason:   "IMPRESSION_TOKEN_EXPIRED",
		}, domain.ErrImpressionTokenExpired
	}

	// 2. Fast Bot Click Check (< 300ms from impression)
	if clickTime.Sub(token.IssuedAt) < 300*time.Millisecond {
		return &domain.ClickResult{
			ClickID:       clickID,
			CampaignID:    token.CampaignID,
			ChargedAmount: 0,
			IsFraud:       true,
			FraudReason:   "FAST_BOT_CLICK_UNDER_300MS",
		}, domain.ErrClickFraudFastBot
	}

	// 3. 1-Hour Device Duplicate Deduplication (50 rapid clicks from same device -> 1st charged, rest 49 free)
	deviceKey := token.DeviceHash + ":" + token.CampaignID
	if lastCharged, exists := f.recentClicks[deviceKey]; exists {
		if clickTime.Sub(lastCharged) < 1*time.Hour {
			return &domain.ClickResult{
				ClickID:       clickID,
				CampaignID:    token.CampaignID,
				ChargedAmount: 0,
				IsFraud:       true,
				FraudReason:   "DUPLICATE_DEVICE_CLICK_WITHIN_1HOUR",
			}, nil // Non-fatal duplicate: returns free click with 0 charge
		}
	}

	// 4. Daily Budget Cap Check
	remainingBudget := campaign.DailyBudget - campaign.SpentToday
	if remainingBudget <= 0 {
		campaign.Status = domain.CampaignPausedBudgetCap
		return &domain.ClickResult{
			ClickID:       clickID,
			CampaignID:    token.CampaignID,
			ChargedAmount: 0,
			IsFraud:       false,
			FraudReason:   "CAMPAIGN_BUDGET_EXHAUSTED",
		}, domain.ErrDailyBudgetCapReached
	}

	// Calculate charge amount (cannot exceed remaining budget)
	charge := token.ClearingCPC
	if charge > remainingBudget {
		charge = remainingBudget
	}

	campaign.SpentToday += charge
	if campaign.SpentToday >= campaign.DailyBudget {
		campaign.Status = domain.CampaignPausedBudgetCap
	}

	f.recentClicks[deviceKey] = clickTime

	return &domain.ClickResult{
		ClickID:       clickID,
		CampaignID:    token.CampaignID,
		ChargedAmount: charge,
		IsFraud:       false,
	}, nil
}

