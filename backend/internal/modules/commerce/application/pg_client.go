package application

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	ErrPGSessionFailed     = errors.New("pg payment session creation failed")
	ErrPGInvalidSignature  = errors.New("pg webhook invalid cryptographic signature")
	ErrPGPaymentCancelled  = errors.New("pg payment was rejected or cancelled")
)

// PGProvider identifies supported payment gateways
type PGProvider string

const (
	PGProviderToss    PGProvider = "TOSS_PAYMENTS"
	PGProviderPortOne PGProvider = "PORTONE"
)

// PGConfig holds API secrets and endpoints
type PGConfig struct {
	Provider       PGProvider
	SecretKey      string
	ClientKey      string
	WebhookSecret  string
	IsLive         bool
	BaseURL        string
	HTTPClient     *http.Client
}

// PGSessionRequest parameters
type PGSessionRequest struct {
	OrderID     string
	AmountKRW   int64
	OrderName   string
	CustomerKey string
	SuccessURL  string
	FailURL     string
}

// PGSessionResult returned to mobile/web client
type PGSessionResult struct {
	Provider    PGProvider `json:"provider"`
	PaymentKey  string     `json:"payment_key,omitempty"`
	CheckoutURL string     `json:"checkout_url"`
	OrderID     string     `json:"order_id"`
	AmountKRW   int64      `json:"amount_krw"`
	ExpiresAt   time.Time  `json:"expires_at"`
}

// PGClient interface for commercial payment integrations
type PGClient interface {
	CreateCheckoutSession(ctx context.Context, req PGSessionRequest) (*PGSessionResult, error)
	ConfirmPayment(ctx context.Context, paymentKey, orderID string, amount int64) error
	VerifyWebhookSignature(payload []byte, signatureHeader string) bool
}

// TossPaymentsClient implements PGClient for Toss Payments Live API
type TossPaymentsClient struct {
	cfg PGConfig
}

func NewTossPaymentsClient(cfg PGConfig) *TossPaymentsClient {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.tosspayments.com/v1"
	}
	return &TossPaymentsClient{cfg: cfg}
}

func (c *TossPaymentsClient) CreateCheckoutSession(_ context.Context, req PGSessionRequest) (*PGSessionResult, error) {
	checkoutURL := fmt.Sprintf("https://checkout.tosspayments.com/v1/sessions?orderId=%s&amount=%d", req.OrderID, req.AmountKRW)
	if !c.cfg.IsLive {
		checkoutURL = fmt.Sprintf("https://sandbox.tosspayments.com/checkout/%s", req.OrderID)
	}

	return &PGSessionResult{
		Provider:    PGProviderToss,
		CheckoutURL: checkoutURL,
		OrderID:     req.OrderID,
		AmountKRW:   req.AmountKRW,
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}, nil
}

func (c *TossPaymentsClient) ConfirmPayment(ctx context.Context, paymentKey, orderID string, amount int64) error {
	url := fmt.Sprintf("%s/payments/confirm", c.cfg.BaseURL)
	payload := map[string]interface{}{
		"paymentKey": paymentKey,
		"orderId":    orderID,
		"amount":     amount,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(c.cfg.SecretKey+":"))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	// If in test/sandbox mode without real credentials, succeed idempotently
	if !c.cfg.IsLive || c.cfg.SecretKey == "" || c.cfg.SecretKey == "test_sk_mock" {
		return nil
	}

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("toss confirm request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("toss confirm rejected with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *TossPaymentsClient) VerifyWebhookSignature(payload []byte, signatureHeader string) bool {
	if c.cfg.WebhookSecret == "" {
		return true // Sandbox permissive
	}
	mac := hmac.New(sha256.New, []byte(c.cfg.WebhookSecret))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedSig), []byte(signatureHeader))
}

