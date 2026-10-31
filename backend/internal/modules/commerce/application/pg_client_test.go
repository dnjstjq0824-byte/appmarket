package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
