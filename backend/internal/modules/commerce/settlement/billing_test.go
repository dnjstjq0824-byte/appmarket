package settlement

import (
	"errors"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
)

func paidOrder() *domain.Order {
	now := time.Now().UTC()
	return &domain.Order{
		ID: "ord-inv-1", UserID: "user-1", AppID: "app-1", OrgID: "org-1",
		Currency: domain.CurrencyKRW, GrossAmount: 11000, TaxAmount: 1000,
		CommissionAmount: 1500, DevPayoutAmount: 8500,
		Status: domain.OrderPaid, PGPaymentID: "pg-1", PaidAt: &now,
	}
}

func TestInvoiceIssueIsIdempotentPerOrder(t *testing.T) {
	svc := NewInvoiceService()
	first := svc.IssueFromOrder(paidOrder())
	second := svc.IssueFromOrder(paidOrder())
	if first.ID != second.ID {
		t.Fatalf("expected same invoice on reissue, got %s vs %s", first.ID, second.ID)
	}
	if first.InvoiceNumber == "" || first.Status != domain.InvoiceIssued {
		t.Fatalf("unexpected invoice: %+v", first)
	}
	if got := svc.ListByUser("user-1"); len(got) != 1 {
		t.Fatalf("expected 1 invoice, got %d", len(got))
	}
	if svc.Get(first.ID) == nil || svc.Get("missing") != nil {
		t.Fatalf("invoice lookup mismatch")
	}
}

func TestInvoiceMarkRefunded(t *testing.T) {
	svc := NewInvoiceService()
	inv := svc.IssueFromOrder(paidOrder())
	svc.MarkRefunded("ord-unknown")
	svc.MarkRefunded(inv.OrderID)
	if got := svc.Get(inv.ID); got.Status != domain.InvoiceRefunded {
		t.Fatalf("expected REFUNDED, got %s", got.Status)
	}
}

func TestSubscriptionLifecycle(t *testing.T) {
	svc := NewSubscriptionService()
	sub, err := svc.CreateSubscription("user-1", "app-1", "org-1", domain.BillingYearly, 99000, domain.CurrencyKRW)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if sub.Status != domain.SubscriptionActive {
		t.Fatalf("expected ACTIVE, got %s", sub.Status)
	}
	if sub.CurrentPeriodEnd.Sub(sub.CurrentPeriodStart) < 364*24*time.Hour {
		t.Fatalf("yearly period too short: %v", sub.CurrentPeriodEnd.Sub(sub.CurrentPeriodStart))
	}
	if _, err := svc.CancelSubscription(sub.ID, "user-2"); !errors.Is(err, domain.ErrSubscriptionAccessDenied) {
		t.Fatalf("expected access denied for other user, got %v", err)
	}
	cancelled, err := svc.CancelSubscription(sub.ID, "user-1")
	if err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	if cancelled.Status != domain.SubscriptionCancelled || cancelled.CancelledAt == nil {
		t.Fatalf("unexpected cancelled state: %+v", cancelled)
	}
	if _, err := svc.CancelSubscription(sub.ID, "user-1"); !errors.Is(err, domain.ErrSubscriptionAlreadyEnded) {
		t.Fatalf("expected already-ended on second cancel, got %v", err)
	}
	if _, err := svc.CancelSubscription("missing", "user-1"); !errors.Is(err, domain.ErrSubscriptionNotFound) {
		t.Fatalf("expected not-found, got %v", err)
	}
}

func TestSubscriptionValidation(t *testing.T) {
	svc := NewSubscriptionService()
	if _, err := svc.CreateSubscription("u", "a", "o", "WEEKLY", 100, domain.CurrencyKRW); !errors.Is(err, domain.ErrInvalidBillingInterval) {
		t.Fatalf("expected invalid interval, got %v", err)
	}
	if _, err := svc.CreateSubscription("u", "a", "o", domain.BillingMonthly, 0, domain.CurrencyKRW); !errors.Is(err, domain.ErrInvalidSubscriptionAmount) {
		t.Fatalf("expected invalid amount, got %v", err)
	}
	monthly, err := svc.CreateSubscription("u", "a", "o", domain.BillingMonthly, 100, "")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if monthly.Currency != domain.CurrencyKRW {
		t.Fatalf("expected KRW default, got %s", monthly.Currency)
	}
}
