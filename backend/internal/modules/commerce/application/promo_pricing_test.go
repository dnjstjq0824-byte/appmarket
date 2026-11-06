package application

import (
	"context"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
)

func promoWindow() (time.Time, time.Time) {
	now := time.Now().UTC()
	return now.Add(-time.Hour), now.Add(24 * time.Hour)
}

func TestPromoCreateAndRedeem(t *testing.T) {
	ctx := context.Background()
	svc := NewPromoService()
	starts, ends := promoWindow()
	promo, codes, err := svc.CreatePromotion(ctx, CreatePromoInput{
		AppID: "app-1", CampaignName: "Launch Week", DiscountType: domain.DiscountPercentage,
		DiscountValue: 20, StartsAt: starts, EndsAt: ends, MaxUses: 100, CodeCount: 3, CodePrefix: "LAUNCH",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if len(codes) != 3 {
		t.Fatalf("expected 3 codes, got %d", len(codes))
	}
	for _, c := range codes {
		if len(c.Code) < 8 || c.Code[:7] != "LAUNCH-" {
			t.Fatalf("unexpected code format: %q", c.Code)
		}
	}
	redemption, err := svc.Redeem(ctx, codes[0].Code, "app-1", "user-1")
	if err != nil {
		t.Fatalf("redeem failed: %v", err)
	}
	if redemption.Promotion.ID != promo.ID || redemption.Promotion.DiscountValue != 20 {
		t.Fatalf("unexpected redemption: %+v", redemption)
	}
	if _, err := svc.Redeem(ctx, codes[0].Code, "app-1", "user-1"); err != domain.ErrPromoAlreadyUsed {
		t.Fatalf("expected already-used, got %v", err)
	}
	if _, err := svc.Redeem(ctx, codes[0].Code, "app-1", "user-2"); err != domain.ErrPromoExhausted {
		t.Fatalf("expected exhausted single-use code, got %v", err)
	}
	if _, err := svc.Redeem(ctx, codes[1].Code, "app-other", "user-1"); err != domain.ErrPromoNotFound {
		t.Fatalf("expected not-found for wrong app, got %v", err)
	}
	if _, err := svc.Redeem(ctx, "NOPE-123", "app-1", "user-1"); err != domain.ErrPromoNotFound {
		t.Fatalf("expected not-found, got %v", err)
	}
	if list := svc.ListByApp(ctx, "app-1"); len(list) != 1 {
		t.Fatalf("expected 1 promotion, got %d", len(list))
	}
}

func TestPromoExpiryAndValidation(t *testing.T) {
	ctx := context.Background()
	svc := NewPromoService()
	now := time.Now().UTC()
	_, codes, err := svc.CreatePromotion(ctx, CreatePromoInput{
		AppID: "app-1", CampaignName: "Past", DiscountType: domain.DiscountFixedMinor,
		DiscountValue: 500, StartsAt: now.Add(-48 * time.Hour), EndsAt: now.Add(-time.Hour),
		MaxUses: 10, CodeCount: 1,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := svc.Redeem(ctx, codes[0].Code, "app-1", "user-1"); err != domain.ErrPromoExpired {
		t.Fatalf("expected expired, got %v", err)
	}
	starts, ends := promoWindow()
	for _, tc := range []struct {
		name string
		in   CreatePromoInput
	}{
		{"bad percentage", CreatePromoInput{AppID: "a", CampaignName: "c", DiscountType: domain.DiscountPercentage, DiscountValue: 101, StartsAt: starts, EndsAt: ends, MaxUses: 1, CodeCount: 1}},
		{"bad window", CreatePromoInput{AppID: "a", CampaignName: "c", DiscountType: domain.DiscountPercentage, DiscountValue: 10, StartsAt: ends, EndsAt: starts, MaxUses: 1, CodeCount: 1}},
		{"bad count", CreatePromoInput{AppID: "a", CampaignName: "c", DiscountType: domain.DiscountPercentage, DiscountValue: 10, StartsAt: starts, EndsAt: ends, MaxUses: 1}},
		{"bad type", CreatePromoInput{AppID: "a", CampaignName: "c", DiscountType: "BOGO", DiscountValue: 10, StartsAt: starts, EndsAt: ends, MaxUses: 1, CodeCount: 1}},
	} {
		if _, _, err := svc.CreatePromotion(ctx, tc.in); err == nil {
			t.Fatalf("%s: expected validation error", tc.name)
		}
	}
}

func TestPromoValidateDoesNotConsume(t *testing.T) {
	ctx := context.Background()
	svc := NewPromoService()
	starts, ends := promoWindow()
	_, codes, err := svc.CreatePromotion(ctx, CreatePromoInput{
		AppID: "app-1", CampaignName: "Validate", DiscountType: domain.DiscountPercentage,
		DiscountValue: 50, StartsAt: starts, EndsAt: ends, MaxUses: 10, CodeCount: 1,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	for i := 0; i < 3; i++ {
		promo, err := svc.Validate(ctx, codes[0].Code, "app-1", "user-1")
		if err != nil {
			t.Fatalf("validate %d failed: %v", i, err)
		}
		if promo.DiscountValue != 50 {
			t.Fatalf("unexpected promo: %+v", promo)
		}
	}
	if _, err := svc.Redeem(ctx, codes[0].Code, "app-1", "user-1"); err != nil {
		t.Fatalf("redeem after validates failed: %v", err)
	}
	if _, err := svc.Validate(ctx, codes[0].Code, "app-1", "user-1"); err != domain.ErrPromoAlreadyUsed {
		t.Fatalf("expected already-used after redeem, got %v", err)
	}
}

func TestPromoRedeemIdempotentOnKey(t *testing.T) {
	ctx := context.Background()
	svc := NewPromoService()
	starts, ends := promoWindow()
	_, codes, err := svc.CreatePromotion(ctx, CreatePromoInput{
		AppID: "app-1", CampaignName: "Idem", DiscountType: domain.DiscountFixedMinor,
		DiscountValue: 500, StartsAt: starts, EndsAt: ends, MaxUses: 10, CodeCount: 1,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	first, err := svc.RedeemWithKey(ctx, codes[0].Code, "app-1", "user-1", "order-key-1")
	if err != nil {
		t.Fatalf("first redeem failed: %v", err)
	}
	second, err := svc.RedeemWithKey(ctx, codes[0].Code, "app-1", "user-1", "order-key-1")
	if err != nil {
		t.Fatalf("retry with same key failed: %v", err)
	}
	if second.Promotion.UsedCount != first.Promotion.UsedCount {
		t.Fatalf("retry consumed an extra use: %+v vs %+v", first, second)
	}
	if _, err := svc.RedeemWithKey(ctx, codes[0].Code, "app-1", "user-1", "order-key-2"); err != domain.ErrPromoAlreadyUsed {
		t.Fatalf("expected already-used for new key, got %v", err)
	}
}

func TestPricingTiers(t *testing.T) {
	ctx := context.Background()
	svc := NewPricingService()
	tier, err := svc.SetTier(ctx, "app-1", "kr", domain.CurrencyKRW, 15000, true)
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if tier.CountryCode != "KR" || tier.ListPriceMinor != 15000 {
		t.Fatalf("unexpected tier: %+v", tier)
	}
	def, err := svc.SetTier(ctx, "app-1", "us", "", 1299, false)
	if err != nil || def.Currency != domain.CurrencyKRW {
		t.Fatalf("expected KRW default, got %+v / %v", def, err)
	}
	if _, err := svc.SetTier(ctx, "app-1", "KOR", domain.CurrencyKRW, 1, true); err == nil {
		t.Fatalf("expected country validation error")
	}
	if _, err := svc.SetTier(ctx, "app-1", "JP", domain.CurrencyKRW, -1, true); err == nil {
		t.Fatalf("expected negative price error")
	}
	if tiers := svc.ListTiers(ctx, "app-1"); len(tiers) != 2 {
		t.Fatalf("expected 2 tiers, got %d", len(tiers))
	}
}
