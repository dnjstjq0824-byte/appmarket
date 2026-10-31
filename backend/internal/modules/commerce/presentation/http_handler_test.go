package presentation_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/presentation"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/settlement"
)

func generateSig(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestCommerceHandler_FullLifecycle(t *testing.T) {
	webhookSecret := "live-test-pg-secret-key-32b"
	ledger := application.NewLedgerService()
	orderService := application.NewOrderService(ledger, webhookSecret)
	refundService := settlement.NewRefundService(ledger, orderService)

	currentUserID := "user-tester-01"
	mockIdentity := func(r *http.Request) (string, error) {
		if r.Header.Get("X-Test-Unauth") == "true" {
			return "", nil
		}
		return currentUserID, nil
	}

	handler := presentation.NewCommerceHandler(orderService, refundService, mockIdentity)
	router := handler.Routes()

	// 1. Unauthenticated request -> 401
	unauthReq := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader([]byte(`{"gross_amount":10000}`)))
	unauthReq.Header.Set("Content-Type", "application/json")
	unauthReq.Header.Set("X-Test-Unauth", "true")
	wUnauth := httptest.NewRecorder()
	router.ServeHTTP(wUnauth, unauthReq)
	if wUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d: %s", wUnauth.Code, wUnauth.Body.String())
	}

	// 2. Create Order -> 201 Created
	orderReqBody, _ := json.Marshal(map[string]any{
		"app_id":          "app-pro-editor",
		"org_id":          "org-dev-studio",
		"gross_amount":    20000,
		"currency":        "KRW",
		"idempotency_key": "order-idem-12345",
	})
	reqOrder := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(orderReqBody))
	reqOrder.Header.Set("Content-Type", "application/json")
	wOrder := httptest.NewRecorder()
	router.ServeHTTP(wOrder, reqOrder)

	if wOrder.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", wOrder.Code, wOrder.Body.String())
	}

	var createdOrder domain.Order
	if err := json.NewDecoder(wOrder.Body).Decode(&createdOrder); err != nil {
		t.Fatalf("failed to decode created order: %v", err)
	}
	if createdOrder.GrossAmount != 20000 || createdOrder.Status != domain.OrderPending {
		t.Fatalf("unexpected order values: %+v", createdOrder)
	}

	// 3. Create Checkout Session -> 200 OK
	reqCheckout := httptest.NewRequest(http.MethodPost, "/orders/"+createdOrder.ID+"/checkout", nil)
	wCheckout := httptest.NewRecorder()
	router.ServeHTTP(wCheckout, reqCheckout)
	if wCheckout.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for checkout session, got %d: %s", wCheckout.Code, wCheckout.Body.String())
	}

	// 4. PG Webhook with invalid signature -> 401
	webhookBody, _ := json.Marshal(map[string]any{
		"order_id":      createdOrder.ID,
		"pg_payment_id": "pg-tx-998877",
		"status":        "PAID",
	})
	reqBadWebhook := httptest.NewRequest(http.MethodPost, "/webhooks/pg", bytes.NewReader(webhookBody))
	reqBadWebhook.Header.Set("Content-Type", "application/json")
	reqBadWebhook.Header.Set("X-PG-Signature", "invalid-fake-signature")
	wBadWebhook := httptest.NewRecorder()
	router.ServeHTTP(wBadWebhook, reqBadWebhook)
	if wBadWebhook.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for bad signature, got %d: %s", wBadWebhook.Code, wBadWebhook.Body.String())
	}

	// 5. PG Webhook with valid HMAC-SHA256 signature -> 200 OK
	validSig := generateSig(webhookBody, webhookSecret)
	reqGoodWebhook := httptest.NewRequest(http.MethodPost, "/webhooks/pg", bytes.NewReader(webhookBody))
	reqGoodWebhook.Header.Set("Content-Type", "application/json")
	reqGoodWebhook.Header.Set("X-PG-Signature", validSig)
	wGoodWebhook := httptest.NewRecorder()
	router.ServeHTTP(wGoodWebhook, reqGoodWebhook)
	if wGoodWebhook.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for valid webhook, got %d: %s", wGoodWebhook.Code, wGoodWebhook.Body.String())
	}

	// 6. Query Entitlements -> User has 1 active entitlement
	reqEnt := httptest.NewRequest(http.MethodGet, "/entitlements", nil)
	wEnt := httptest.NewRecorder()
	router.ServeHTTP(wEnt, reqEnt)
	if wEnt.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for entitlements, got %d: %s", wEnt.Code, wEnt.Body.String())
	}

	var entitlements []domain.UserEntitlement
	if err := json.NewDecoder(wEnt.Body).Decode(&entitlements); err != nil {
		t.Fatalf("failed to decode entitlements: %v", err)
	}
	if len(entitlements) != 1 || entitlements[0].AppID != "app-pro-editor" {
		t.Fatalf("expected 1 active entitlement for app-pro-editor, got %+v", entitlements)
	}

	// 7. Request Refund -> 200 OK, Entitlement Revoked
	reqRefund := httptest.NewRequest(http.MethodPost, "/orders/"+createdOrder.ID+"/refund", bytes.NewReader([]byte(`{"reason":"Accidental purchase"}`)))
	reqRefund.Header.Set("Content-Type", "application/json")
	wRefund := httptest.NewRecorder()
	router.ServeHTTP(wRefund, reqRefund)
	if wRefund.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for refund, got %d: %s", wRefund.Code, wRefund.Body.String())
	}

	// Verify entitlement is revoked
	entAfter := orderService.GetEntitlement(currentUserID, "app-pro-editor")
	if entAfter.Status != domain.EntitlementRevoked {
		t.Fatalf("expected entitlement revoked after refund, got %s", entAfter.Status)
	}
}

