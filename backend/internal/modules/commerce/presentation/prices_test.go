package presentation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/presentation"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/settlement"
)

func pricedCommerceRouter(t *testing.T) (*application.PricingService, *application.PromoService, http.Handler) {
	t.Helper()
	ledger := application.NewLedgerService()
	orderService := application.NewOrderService(ledger, "price-test-secret")
	refundService := settlement.NewRefundService(ledger, orderService)
	pricing := application.NewPricingService()
	promos := application.NewPromoService()
	mockIdentity := func(r *http.Request) (string, error) { return "user-price-01", nil }
	pgClient := application.NewTossPaymentsClient(application.PGConfig{Provider: application.PGProviderToss})
	handler := presentation.NewCommerceHandler(orderService, refundService, nil, nil, mockIdentity, pgClient, "https://appmarket.test", promos)
	handler.WithPricing(pricing)
	return pricing, promos, handler.Routes()
}

func TestCommercePublicPrices(t *testing.T) {
	pricing, _, router := pricedCommerceRouter(t)
	ctx := context.Background()
	if _, err := pricing.SetTier(ctx, "app-paid-1", "KR", domain.CurrencyKRW, 4900, true); err != nil {
		t.Fatalf("set tier: %v", err)
	}
	if _, err := pricing.SetTier(ctx, "app-paid-1", "US", domain.CurrencyUSD, 399, true); err != nil {
		t.Fatalf("set tier: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/apps/app-paid-1/prices", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("prices: %d %s", w.Code, w.Body.String())
	}
	var body struct {
		Items []domain.PriceTier `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 2 {
		t.Fatalf("expected 2 tiers, got %+v", body.Items)
	}
}

func TestCommerceOrderRejectsTamperedAmount(t *testing.T) {
	pricing, _, router := pricedCommerceRouter(t)
	ctx := context.Background()
	if _, err := pricing.SetTier(ctx, "app-paid-2", "KR", domain.CurrencyKRW, 4900, true); err != nil {
		t.Fatalf("set tier: %v", err)
	}

	post := func(amount int64) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(map[string]any{
			"app_id": "app-paid-2", "gross_amount": amount, "currency": "KRW",
		})
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	if w := post(100); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for tampered amount, got %d %s", w.Code, w.Body.String())
	}
	if w := post(4900); w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for list price, got %d %s", w.Code, w.Body.String())
	}
}

func TestCommerceOrderWithPromoCode(t *testing.T) {
	pricing, promos, router := pricedCommerceRouter(t)
	ctx := context.Background()
	if _, err := pricing.SetTier(ctx, "app-promo-1", "KR", domain.CurrencyKRW, 15000, true); err != nil {
		t.Fatalf("set tier: %v", err)
	}
	starts := time.Now().UTC().Add(-time.Hour)
	ends := time.Now().UTC().Add(24 * time.Hour)
	_, codes, err := promos.CreatePromotion(ctx, application.CreatePromoInput{
		AppID: "app-promo-1", CampaignName: "Half", DiscountType: domain.DiscountPercentage,
		DiscountValue: 50, StartsAt: starts, EndsAt: ends, MaxUses: 10, CodeCount: 1, CodePrefix: "HALF",
	})
	if err != nil {
		t.Fatalf("create promo: %v", err)
	}

	validateBody, _ := json.Marshal(map[string]any{"code": codes[0].Code, "appId": "app-promo-1"})
	req := httptest.NewRequest(http.MethodPost, "/promos/validate", bytes.NewReader(validateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("validate: %d %s", w.Code, w.Body.String())
	}

	post := func(amount int64, promo, key string) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(map[string]any{
			"app_id": "app-promo-1", "gross_amount": amount, "currency": "KRW",
			"promo_code": promo, "idempotency_key": key,
		})
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	if w := post(15000, codes[0].Code, "k-wrong-amount"); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for undiscounted amount with promo, got %d %s", w.Code, w.Body.String())
	}
	w = post(7500, codes[0].Code, "k-half-1")
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for discounted amount, got %d %s", w.Code, w.Body.String())
	}
	if w := post(7500, codes[0].Code, "k-half-2"); w.Code != http.StatusConflict {
		t.Fatalf("expected 409 for reused code, got %d %s", w.Code, w.Body.String())
	}
	// Same idempotency key replays the original order instead of failing.
	if w := post(7500, codes[0].Code, "k-half-1"); w.Code != http.StatusCreated {
		t.Fatalf("expected 201 replay, got %d %s", w.Code, w.Body.String())
	}
}

func TestCommerceFreeRedeemCompletesWithoutPG(t *testing.T) {
	pricing, promos, router := pricedCommerceRouter(t)
	ctx := context.Background()
	if _, err := pricing.SetTier(ctx, "app-promo-2", "KR", domain.CurrencyKRW, 9900, true); err != nil {
		t.Fatalf("set tier: %v", err)
	}
	starts := time.Now().UTC().Add(-time.Hour)
	ends := time.Now().UTC().Add(24 * time.Hour)
	_, codes, err := promos.CreatePromotion(ctx, application.CreatePromoInput{
		AppID: "app-promo-2", CampaignName: "Free", DiscountType: domain.DiscountFreeRedeem,
		DiscountValue: 0, StartsAt: starts, EndsAt: ends, MaxUses: 10, CodeCount: 1, CodePrefix: "FREE",
	})
	if err != nil {
		t.Fatalf("create promo: %v", err)
	}
	raw, _ := json.Marshal(map[string]any{
		"app_id": "app-promo-2", "gross_amount": 0, "currency": "KRW",
		"promo_code": codes[0].Code, "idempotency_key": "k-free-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for free redeem, got %d %s", w.Code, w.Body.String())
	}
	var order domain.Order
	if err := json.Unmarshal(w.Body.Bytes(), &order); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if order.Status != domain.OrderPaid {
		t.Fatalf("expected free order PAID, got %s", order.Status)
	}
}
