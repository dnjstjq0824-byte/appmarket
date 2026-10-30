package application

import (
	"context"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/ads/domain"
)

// TC-COM-003: GSP Auction, 3-Tier Click Fraud Filter & Daily Budget Cap
func TestTCCOM003GSPAuctionAndClickFraudFilter(t *testing.T) {
	ctx := context.Background()
	engine := NewAuctionEngine(50) // ₩50 floor
	filter := NewClickFraudFilter(engine)

	// Setup 2 competing campaigns for keyword "hotel"
	// Campaign A: MaxBid = 500, Quality = 8.0 -> RankScore = 4000
	campA := domain.AdCampaign{
		ID:             "camp-a",
		OrgID:          "org-hotel-a",
		AppID:          "app-hotel-a",
		AppName:        "Hotel Booking Best",
		TargetKeywords: []string{"hotel", "travel"},
		MaxBidCPC:      500,
		QualityScore:   8.0,
		DailyBudget:    2000, // ₩2,000 daily budget
		SpentToday:     0,
		Status:         domain.CampaignActive,
	}

	// Campaign B: MaxBid = 400, Quality = 7.0 -> RankScore = 2800
	campB := domain.AdCampaign{
		ID:             "camp-b",
		OrgID:          "org-hotel-b",
		AppID:          "app-hotel-b",
		AppName:        "Stay Cheap Hotel",
		TargetKeywords: []string{"hotel"},
		MaxBidCPC:      400,
		QualityScore:   7.0,
		DailyBudget:    10000,
		SpentToday:     0,
		Status:         domain.CampaignActive,
	}

	engine.AddCampaign(campA)
	engine.AddCampaign(campB)

	t.Run("Sub-test 1: GSP Auction Winner and Clearing Price Calculation", func(t *testing.T) {
		winners, tokens := engine.RunGSPAuction("hotel", "dev-hash-test", 1)
		if len(winners) != 1 || len(tokens) != 1 {
			t.Fatalf("expected 1 winner, got %d", len(winners))
		}

		winner := winners[0]
		if winner.Campaign.ID != "camp-a" {
			t.Fatalf("expected camp-a to win with higher rank score, got %s", winner.Campaign.ID)
		}

		// Second price: ceil(2800 / 8.0) + 1 = 350 + 1 = ₩351
		if winner.ClearingCPC != 351 {
			t.Fatalf("expected GSP clearing price 351, got %d", winner.ClearingCPC)
		}
	})

	t.Run("Sub-test 2: 50 Continuous Clicks from same device charges exactly 1st click", func(t *testing.T) {
		_, tokens := engine.RunGSPAuction("hotel", "dev-device-clicker", 1)
		token := tokens[0]

		var totalCharged int64 = 0
		fraudCount := 0

		baseTime := token.IssuedAt.Add(500 * time.Millisecond) // After 500ms (valid human click)

		for i := 0; i < 50; i++ {
			// Clicks occur rapidly within 10 seconds of baseTime
			clickTime := baseTime.Add(time.Duration(i*200) * time.Millisecond)
			result, err := filter.ProcessAdClick(ctx, token, clickTime)
			if err != nil {
				t.Fatalf("click %d failed: %v", i, err)
			}

			totalCharged += result.ChargedAmount
			if result.IsFraud {
				fraudCount++
			}
		}

		// Exactly 1st click charged ₩351, 49 duplicate clicks charged ₩0
		if totalCharged != token.ClearingCPC {
			t.Fatalf("expected total charged to equal exactly 1 click (%d), got %d", token.ClearingCPC, totalCharged)
		}
		if fraudCount != 49 {
			t.Fatalf("expected 49 fraudulent duplicate clicks, got %d", fraudCount)
		}
	})

	t.Run("Sub-test 3: Bot Fast Click (<300ms) is Rejected", func(t *testing.T) {
		_, tokens := engine.RunGSPAuction("hotel", "dev-bot-device", 1)
		token := tokens[0]

		botTime := token.IssuedAt.Add(50 * time.Millisecond) // Only 50ms after impression
		_, err := filter.ProcessAdClick(ctx, token, botTime)

		if err != domain.ErrClickFraudFastBot {
			t.Fatalf("expected ErrClickFraudFastBot for 50ms click, got %v", err)
		}
	})

	t.Run("Sub-test 4: Daily Budget Cap automatically pauses campaign without budget overshoot", func(t *testing.T) {
		// Campaign A budget is 2,000. It already spent 351 in sub-test 2.
		camp := engine.GetCampaign("camp-a")

		// Simulate multiple unique devices clicking until budget is depleted
		for i := 0; i < 10; i++ {
			devHash := "unique-dev-" + string(rune('0'+i))
			_, tokens := engine.RunGSPAuction("hotel", devHash, 1)
			if len(tokens) == 0 {
				break // Campaign paused due to budget cap!
			}
			token := tokens[0]
			clickTime := token.IssuedAt.Add(1 * time.Second)

			_, _ = filter.ProcessAdClick(ctx, token, clickTime)
		}

		if camp.SpentToday > camp.DailyBudget {
			t.Fatalf("Daily budget exceeded! SpentToday (%d) > DailyBudget (%d)", camp.SpentToday, camp.DailyBudget)
		}
		if camp.Status != domain.CampaignPausedBudgetCap {
			t.Fatalf("expected CampaignPausedBudgetCap status, got %s", camp.Status)
		}
	})
}

