package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
)

type OrderService struct {
	ledgerService *LedgerService
	calculator    *domain.PricingCalculator
	webhookSecret []byte

	mu           sync.RWMutex
	orders       map[string]*domain.Order
	entitlements map[string]*domain.UserEntitlement // key: userID + ":" + appID
	paidAppList  map[string]bool                    // appID -> isPaid
}

func NewOrderService(ledger *LedgerService, webhookSecret string) *OrderService {
	return &OrderService{
		ledgerService: ledger,
		calculator:    domain.NewPricingCalculator(),
		webhookSecret: []byte(webhookSecret),
		orders:        make(map[string]*domain.Order),
		entitlements:  make(map[string]*domain.UserEntitlement),
		paidAppList:   make(map[string]bool),
	}
}

func (s *OrderService) RegisterPaidApp(appID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paidAppList[appID] = true
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, appID, orgID string, grossAmount int64, currency domain.Currency, idempotencyKey string) (*domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	breakdown := s.calculator.CalculateBreakdown(grossAmount)

	orderID := fmt.Sprintf("ord-%d", time.Now().UnixNano())
	order := &domain.Order{
		ID:               orderID,
		UserID:           userID,
		AppID:            appID,
		OrgID:            orgID,
		Currency:         currency,
		GrossAmount:      breakdown.GrossAmount,
		TaxAmount:        breakdown.TaxAmount,
		CommissionAmount: breakdown.CommissionAmount,
		DevPayoutAmount:  breakdown.DevPayoutAmount,
		Status:           domain.OrderPending,
		IdempotencyKey:   idempotencyKey,
		CreatedAt:        time.Now().UTC(),
	}

	s.orders[orderID] = order
	return order, nil
}

// ProcessPaymentWebhook verifies signature and posts idempotent payment & double-entry ledger
func (s *OrderService) ProcessPaymentWebhook(ctx context.Context, orderID, pgPaymentID string, rawPayload []byte, signatureHex string) (*domain.Order, error) {
	// 1. Verify PG HMAC-SHA256 signature
	if !s.verifySignature(rawPayload, signatureHex) {
		return nil, domain.ErrInvalidWebhookSignature
	}

	s.mu.Lock()
	order, exists := s.orders[orderID]
	if !exists {
		s.mu.Unlock()
		return nil, errors.New("order not found")
	}

	// 2. Idempotency Check: if already paid, return idempotently without duplicate journalizing
	if order.Status == domain.OrderPaid {
		s.mu.Unlock()
		return order, nil
	}

	now := time.Now().UTC()
	order.Status = domain.OrderPaid
	order.PGPaymentID = pgPaymentID
	order.PaidAt = &now

	// 3. Issue User Entitlement
	entitlementKey := order.UserID + ":" + order.AppID
	entitlement := &domain.UserEntitlement{
		ID:        fmt.Sprintf("ent-%d", now.UnixNano()),
		UserID:    order.UserID,
		AppID:     order.AppID,
		OrderID:   order.ID,
		Status:    domain.EntitlementActive,
		GrantedAt: now,
	}
	s.entitlements[entitlementKey] = entitlement
	s.mu.Unlock()

	// 4. Double-Entry Ledger Journal: Debit Clearing, Credit Commission, Tax, Dev Escrow
	txID := fmt.Sprintf("tx-pay-%s", order.ID)
	entries := []domain.LedgerEntry{
		{Account: "ASSET:PG_CLEARING", Type: domain.EntryDebit, Amount: order.GrossAmount},
		{Account: "REV:PLATFORM_COMMISSION", Type: domain.EntryCredit, Amount: order.CommissionAmount},
		{Account: "LIAB:TAX_PAYABLE", Type: domain.EntryCredit, Amount: order.TaxAmount},
		{Account: "LIAB:DEV_ESCROW", Type: domain.EntryCredit, Amount: order.DevPayoutAmount},
	}

	_, err := s.ledgerService.RecordTransaction(ctx, domain.LedgerTransaction{
		ID:          txID,
		ReferenceID: order.ID,
		Description: fmt.Sprintf("Payment for Order %s", order.ID),
		Entries:     entries,
		CreatedAt:   now,
	}, "webhook-idem-"+order.ID)

	if err != nil {
		return nil, err
	}

	return order, nil
}

// VerifyUserEntitlement checks if user has entitlement for paid app
func (s *OrderService) VerifyUserEntitlement(ctx context.Context, userID, appID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	isPaid := s.paidAppList[appID]
	if !isPaid {
		return true, nil // Free app requires no entitlement
	}

	entitlement, exists := s.entitlements[userID+":"+appID]
	if !exists || entitlement.Status != domain.EntitlementActive {
		return false, domain.ErrEntitlementRequired
	}

	return true, nil
}

func (s *OrderService) GetEntitlement(userID, appID string) *domain.UserEntitlement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entitlements[userID+":"+appID]
}

func (s *OrderService) verifySignature(payload []byte, signatureHex string) bool {
	if len(s.webhookSecret) == 0 {
		return true
	}
	mac := hmac.New(sha256.New, s.webhookSecret)
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signatureHex))
}
