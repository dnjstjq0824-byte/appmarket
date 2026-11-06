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

// PGSessionResult returned to mobile/web client.
// Toss fills CheckoutURL (redirect the buyer there). PortOne fills PaymentKey
// plus ClientKey (complete the payment with the PortOne browser SDK using the
// store id in ClientKey); CheckoutURL is empty for PortOne.
type PGSessionResult struct {
	Provider    PGProvider `json:"provider"`
	PaymentKey  string     `json:"payment_key,omitempty"`
	ClientKey   string     `json:"client_key,omitempty"`
	CheckoutURL string     `json:"checkout_url,omitempty"`
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

// PortOneV2Client implements PGClient for PortOne (ex-Iamport) V2 REST API.
// Live mode authenticates with the API secret, verifies payments against the
// PortOne ledger, and HMAC-verifies webhooks. Without credentials (or with
// IsLive=false) every call succeeds idempotently in sandbox mode.
type PortOneV2Client struct {
	cfg PGConfig
}

func NewPortOneV2Client(cfg PGConfig) *PortOneV2Client {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.portone.io"
	}
	return &PortOneV2Client{cfg: cfg}
}

func (c *PortOneV2Client) sandbox() bool {
	return !c.cfg.IsLive || c.cfg.SecretKey == ""
}

func (c *PortOneV2Client) CreateCheckoutSession(_ context.Context, req PGSessionRequest) (*PGSessionResult, error) {
	if req.AmountKRW <= 0 {
		return nil, ErrPGSessionFailed
	}
	return &PGSessionResult{
		Provider:   PGProviderPortOne,
		PaymentKey: fmt.Sprintf("pay-%s-%d", req.OrderID, time.Now().Unix()),
		ClientKey:  c.cfg.ClientKey,
		OrderID:    req.OrderID,
		AmountKRW:  req.AmountKRW,
		ExpiresAt:  time.Now().Add(30 * time.Minute),
	}, nil
}

func (c *PortOneV2Client) portOneToken(ctx context.Context) (string, error) {
	body, _ := json.Marshal(map[string]string{"apiSecret": c.cfg.SecretKey})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/login/api-secret", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("portone token request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("portone token rejected with status: %d", resp.StatusCode)
	}
	var tokenResp struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("portone token decode failed: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", errors.New("portone token response missing accessToken")
	}
	return tokenResp.AccessToken, nil
}

func (c *PortOneV2Client) ConfirmPayment(ctx context.Context, paymentKey, orderID string, amount int64) error {
	if c.sandbox() {
		return nil
	}
	token, err := c.portOneToken(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/payments/%s", c.cfg.BaseURL, paymentKey), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("portone payment lookup failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("portone payment lookup rejected with status: %d", resp.StatusCode)
	}
	var payment struct {
		Status string `json:"status"`
		Amount struct {
			Total int64 `json:"total"`
		} `json:"amount"`
		OrderName string `json:"orderName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
		return fmt.Errorf("portone payment decode failed: %w", err)
	}
	if payment.Status != "PAID" {
		return ErrPGPaymentCancelled
	}
	if payment.Amount.Total != amount {
		return fmt.Errorf("%w: ledger amount %d != order amount %d", ErrPGSessionFailed, payment.Amount.Total, amount)
	}
	return nil
}

func (c *PortOneV2Client) VerifyWebhookSignature(payload []byte, signatureHeader string) bool {
	if c.cfg.WebhookSecret == "" {
		return true // Sandbox permissive
	}
	mac := hmac.New(sha256.New, []byte(c.cfg.WebhookSecret))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedSig), []byte(signatureHeader))
}

