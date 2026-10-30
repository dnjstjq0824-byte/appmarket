package settlement

import (
	"context"
	"fmt"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
)

type EscrowReleaseWorker struct {
	ledgerService *application.LedgerService
}

func NewEscrowReleaseWorker(ledger *application.LedgerService) *EscrowReleaseWorker {
	return &EscrowReleaseWorker{
		ledgerService: ledger,
	}
}

// ReleaseMaturedEscrow transfers funds from LIAB:DEV_ESCROW to LIAB:DEV_SETTLEABLE after exactly 14 days
func (w *EscrowReleaseWorker) ReleaseMaturedEscrow(ctx context.Context, order *domain.Order) error {
	if order.PaidAt == nil {
		return fmt.Errorf("order %s has not been paid", order.ID)
	}

	escrowMaturityWindow := 14 * 24 * time.Hour
	if time.Since(*order.PaidAt) < escrowMaturityWindow {
		return domain.ErrEscrowNotMatured
	}

	txID := fmt.Sprintf("tx-escrow-rel-%s", order.ID)
	entries := []domain.LedgerEntry{
		{Account: "LIAB:DEV_ESCROW", Type: domain.EntryDebit, Amount: order.DevPayoutAmount},
		{Account: "LIAB:DEV_SETTLEABLE", Type: domain.EntryCredit, Amount: order.DevPayoutAmount},
	}

	_, err := w.ledgerService.RecordTransaction(ctx, domain.LedgerTransaction{
		ID:          txID,
		ReferenceID: order.ID,
		Description: fmt.Sprintf("14-day escrow release for Order %s", order.ID),
		Entries:     entries,
		CreatedAt:   time.Now().UTC(),
	}, "escrow-rel-idem-"+order.ID)

	return err
}
