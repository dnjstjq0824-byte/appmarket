package settlement

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
)

type PayoutStatus string

const (
	PayoutPendingApproval PayoutStatus = "PENDING_APPROVAL"
	PayoutPaidCompleted   PayoutStatus = "PAID_COMPLETED"
	PayoutRejected        PayoutStatus = "REJECTED"
)

type SettlementPayout struct {
	ID           string       `json:"id"`
	OrgID        string       `json:"org_id"`
	Amount       int64        `json:"amount"`
	Currency     domain.Currency `json:"currency"`
	Status       PayoutStatus `json:"status"`
	CreatedBy    string       `json:"created_by"`
	ApprovedBy   string       `json:"approved_by,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	ApprovedAt   *time.Time   `json:"approved_at,omitempty"`
}

type MonthlyPayoutService struct {
	ledgerService *application.LedgerService
	mu            sync.RWMutex
	payouts       map[string]*SettlementPayout
}

func NewMonthlyPayoutService(ledger *application.LedgerService) *MonthlyPayoutService {
	return &MonthlyPayoutService{
		ledgerService: ledger,
		payouts:       make(map[string]*SettlementPayout),
	}
}

func (s *MonthlyPayoutService) CreatePayoutRequest(ctx context.Context, orgID string, amount int64, currency domain.Currency, creatorID string) (*SettlementPayout, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payoutID := fmt.Sprintf("payout-%d", time.Now().UnixNano())
	payout := &SettlementPayout{
		ID:        payoutID,
		OrgID:     orgID,
		Amount:    amount,
		Currency:  currency,
		Status:    PayoutPendingApproval,
		CreatedBy: creatorID,
		CreatedAt: time.Now().UTC(),
	}

	s.payouts[payoutID] = payout
	return payout, nil
}

// ApprovePayout enforces Dual-Control: Approver must be distinct from Creator
func (s *MonthlyPayoutService) ApprovePayout(ctx context.Context, payoutID string, approverID string) (*SettlementPayout, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payout, exists := s.payouts[payoutID]
	if !exists {
		return nil, fmt.Errorf("payout %s not found", payoutID)
	}

	// Dual-Control Enforcement: Self-approval is strictly forbidden
	if approverID == payout.CreatedBy {
		return nil, domain.ErrSelfApprovalForbidden
	}

	now := time.Now().UTC()
	payout.Status = PayoutPaidCompleted
	payout.ApprovedBy = approverID
	payout.ApprovedAt = &now

	// Record ledger transaction: Debit LIAB:DEV_SETTLEABLE, Credit ASSET:BANK_PAYOUT
	txID := fmt.Sprintf("tx-payout-%s", payout.ID)
	entries := []domain.LedgerEntry{
		{Account: "LIAB:DEV_SETTLEABLE", Type: domain.EntryDebit, Amount: payout.Amount},
		{Account: "ASSET:BANK_PAYOUT", Type: domain.EntryCredit, Amount: payout.Amount},
	}

	_, err := s.ledgerService.RecordTransaction(ctx, domain.LedgerTransaction{
		ID:          txID,
		ReferenceID: payout.ID,
		Description: fmt.Sprintf("Monthly settlement bank payout for Org %s", payout.OrgID),
		Entries:     entries,
		CreatedAt:   now,
	}, "payout-idem-"+payout.ID)

	if err != nil {
		return nil, err
	}

	return payout, nil
}
