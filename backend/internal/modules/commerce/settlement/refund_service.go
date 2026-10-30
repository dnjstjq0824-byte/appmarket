package settlement

import (
	"context"
	"fmt"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
)

type RefundService struct {
	ledgerService *application.LedgerService
	orderService  *application.OrderService
}

func NewRefundService(ledger *application.LedgerService, order *application.OrderService) *RefundService {
	return &RefundService{
		ledgerService: ledger,
		orderService:  order,
	}
}

// ProcessRefund executes customer refund, revokes entitlement, and journals reverse double-entry
func (s *RefundService) ProcessRefund(ctx context.Context, order *domain.Order, reason string) error {
	if order.Status == domain.OrderRefunded {
		return domain.ErrOrderAlreadyRefunded
	}

	now := time.Now().UTC()
	order.Status = domain.OrderRefunded
	order.RefundedAt = &now

	// 1. Revoke User Entitlement immediately
	entitlement := s.orderService.GetEntitlement(order.UserID, order.AppID)
	if entitlement != nil {
		entitlement.Status = domain.EntitlementRevoked
		entitlement.RevokedAt = &now
	}

	// 2. Reverse Double-Entry Ledger Entry
	// Debit Commission, Tax, Dev Escrow; Credit PG Clearing
	txID := fmt.Sprintf("tx-ref-%s", order.ID)
	entries := []domain.LedgerEntry{
		{Account: "REV:PLATFORM_COMMISSION", Type: domain.EntryDebit, Amount: order.CommissionAmount},
		{Account: "LIAB:TAX_PAYABLE", Type: domain.EntryDebit, Amount: order.TaxAmount},
		{Account: "LIAB:DEV_ESCROW", Type: domain.EntryDebit, Amount: order.DevPayoutAmount},
		{Account: "ASSET:PG_CLEARING", Type: domain.EntryCredit, Amount: order.GrossAmount},
	}

	_, err := s.ledgerService.RecordTransaction(ctx, domain.LedgerTransaction{
		ID:          txID,
		ReferenceID: order.ID,
		Description: fmt.Sprintf("Refund for Order %s: %s", order.ID, reason),
		Entries:     entries,
		CreatedAt:   now,
	}, "refund-idem-"+order.ID)

	return err
}
