package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPGClient_WebhookHMACVerification(t *testing.T) {
	webhookSecret := "sandbox_pg_webhook_secret_key_32_bytes"
	cfg := PGConfig{
		Provider:      PGProviderToss,
		SecretKey:     "test_sk_sample",
		ClientKey:     "test_ck_sample",
		WebhookSecret: webhookSecret,
		IsLive:        false,
	}
	client := NewTossPaymentsClient(cfg)

	payload := []byte(`{"orderId":"ord-12345","amount":4900,"status":"DONE"}`)

	// Compute valid HMAC
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(payload)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	// 1. Valid Signature Test
	if !client.VerifyWebhookSignature(payload, validSignature) {
		t.Fatal("valid signature verification failed")
	}

	// 2. Tampered Signature Test
	tamperedSig := validSignature[:len(validSignature)-4] + "dead"
	if client.VerifyWebhookSignature(payload, tamperedSig) {
		t.Fatal("SECURITY FLAW: Tampered webhook signature was accepted!")
	}
	t.Log("Tampered webhook signature properly rejected.")

	// 3. Tampered Payload Test
	tamperedPayload := []byte(`{"orderId":"ord-12345","amount":999999,"status":"DONE"}`)
	if client.VerifyWebhookSignature(tamperedPayload, validSignature) {
		t.Fatal("SECURITY FLAW: Webhook with altered payload was accepted!")
	}
	t.Log("Tampered payload properly rejected.")
}

func TestPGClient_SandboxSessionGeneration(t *testing.T) {
	ctx := context.Background()
	cfg := PGConfig{
		Provider:      PGProviderToss,
		SecretKey:     "test_sk_sample",
		ClientKey:     "test_ck_sample",
		WebhookSecret: "test_wh_secret",
		IsLive:        false,
	}
	client := NewTossPaymentsClient(cfg)

	req := PGSessionRequest{
		OrderID:     "ord-test-001",
		AmountKRW:   9900,
		OrderName:   "Test App License",
		CustomerKey: "cust-001",
		SuccessURL:  "https://appmarket.local/success",
		FailURL:     "https://appmarket.local/fail",
	}

	res, err := client.CreateCheckoutSession(ctx, req)
	if err != nil {
		t.Fatalf("checkout session creation failed: %v", err)
	}

	if res.CheckoutURL == "" {
		t.Fatal("expected non-empty checkout URL")
	}
	if res.AmountKRW != 9900 {
		t.Fatalf("expected 9900, got %d", res.AmountKRW)
	}
}

func TestPortOneV2Client_SessionAndSandboxConfirm(t *testing.T) {
	ctx := context.Background()
	client := NewPortOneV2Client(PGConfig{
		Provider:  PGProviderPortOne,
		ClientKey: "store-test-1",
		IsLive:    false,
	})
	res, err := client.CreateCheckoutSession(ctx, PGSessionRequest{OrderID: "ord-1", AmountKRW: 5000})
	if err != nil {
		t.Fatalf("session creation failed: %v", err)
	}
	if res.Provider != PGProviderPortOne || res.PaymentKey == "" || res.ClientKey != "store-test-1" {
		t.Fatalf("unexpected portone session: %+v", res)
	}
	if res.CheckoutURL != "" {
		t.Fatalf("portone session must not set redirect URL, got %q", res.CheckoutURL)
	}
	if err := client.ConfirmPayment(ctx, res.PaymentKey, "ord-1", 5000); err != nil {
		t.Fatalf("sandbox confirm must succeed: %v", err)
	}
	if _, err := client.CreateCheckoutSession(ctx, PGSessionRequest{OrderID: "ord-1"}); err == nil {
		t.Fatalf("expected error for zero amount")
	}
}

func TestPortOneV2Client_LiveConfirmVerifiesLedger(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/login/api-secret":
			_, _ = w.Write([]byte(`{"accessToken":"tok-test"}`))
		case strings.HasPrefix(r.URL.Path, "/payments/pay-ok"):
			_, _ = w.Write([]byte(`{"status":"PAID","amount":{"total":5000},"orderName":"License"}`))
		case strings.HasPrefix(r.URL.Path, "/payments/pay-wrong-amount"):
			_, _ = w.Write([]byte(`{"status":"PAID","amount":{"total":1},"orderName":"License"}`))
		default:
			_, _ = w.Write([]byte(`{"status":"FAILED","amount":{"total":5000},"orderName":"License"}`))
		}
	}))
	defer server.Close()

	client := NewPortOneV2Client(PGConfig{
		Provider:  PGProviderPortOne,
		SecretKey: "live-secret",
		BaseURL:   server.URL,
		IsLive:    true,
	})
	if err := client.ConfirmPayment(ctx, "pay-ok", "ord-1", 5000); err != nil {
		t.Fatalf("expected live confirm success: %v", err)
	}
	if err := client.ConfirmPayment(ctx, "pay-wrong-amount", "ord-1", 5000); err == nil {
		t.Fatalf("expected amount-mismatch rejection")
	}
	if !errors.Is(client.ConfirmPayment(ctx, "pay-failed", "ord-1", 5000), ErrPGPaymentCancelled) {
		t.Fatalf("expected ErrPGPaymentCancelled for FAILED payment")
	}
}

func TestTossPaymentsClient_LiveConfirmPostsToConfirmEndpoint(t *testing.T) {
	ctx := context.Background()
	const secret = "live_sk_realistic_secret"
	var gotAuth, gotMethod, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotMethod, gotPath = r.Header.Get("Authorization"), r.Method, r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewTossPaymentsClient(PGConfig{
		Provider:  PGProviderToss,
		SecretKey: secret,
		BaseURL:   server.URL,
		IsLive:    true,
	})
	if err := client.ConfirmPayment(ctx, "pay-live-1", "ord-live-1", 9900); err != nil {
		t.Fatalf("expected live confirm success: %v", err)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(secret+":"))
	if gotAuth != wantAuth {
		t.Fatalf("expected Authorization %q, got %q", wantAuth, gotAuth)
	}
	if gotMethod != http.MethodPost || gotPath != "/payments/confirm" {
		t.Fatalf("expected POST /payments/confirm, got %s %s", gotMethod, gotPath)
	}

	reject := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer reject.Close()
	rejectClient := NewTossPaymentsClient(PGConfig{
		Provider:  PGProviderToss,
		SecretKey: secret,
		BaseURL:   reject.URL,
		IsLive:    true,
	})
	if err := rejectClient.ConfirmPayment(ctx, "pay-live-2", "ord-live-2", 9900); err == nil {
		t.Fatalf("expected rejection error for 400 confirm response")
	} else if !strings.Contains(err.Error(), "toss confirm rejected") {
		t.Fatalf("unexpected error for rejected confirm: %v", err)
	}
}
