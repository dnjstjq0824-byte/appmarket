package rbac

import (
	"context"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
)

func TestDualControlListPendingFiltersDecided(t *testing.T) {
	ctx := context.Background()
	engine := NewDualControlEngine(nil)

	pending, err := engine.CreateRequest(ctx, "SETTLEMENT_PAYOUT_OVER_LIMIT", "admin-a", "payout-1", "{}", "monthly payout")
	if err != nil {
		t.Fatalf("create pending: %v", err)
	}
	other, err := engine.CreateRequest(ctx, "PERMANENT_APP_BAN", "admin-a", "app-1", "{}", "malware")
	if err != nil {
		t.Fatalf("create other: %v", err)
	}
	if _, err := engine.Approve(ctx, other.ID, "admin-b"); err != nil {
		t.Fatalf("approve: %v", err)
	}

	list, err := engine.ListPending(ctx)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(list) != 1 || list[0].ID != pending.ID {
		t.Fatalf("expected only pending request %s, got %+v", pending.ID, list)
	}
	if list[0].Status != domain.ApprovalPending {
		t.Fatalf("expected PENDING status, got %s", list[0].Status)
	}
}

func TestDualControlSelfApprovalForbidden(t *testing.T) {
	ctx := context.Background()
	engine := NewDualControlEngine(nil)

	req, err := engine.CreateRequest(ctx, "GRANT_ADMIN_ROLE", "admin-a", "user-9", "{}", "promote")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := engine.Approve(ctx, req.ID, "admin-a"); err != domain.ErrDualControlSelfApproval {
		t.Fatalf("expected ErrDualControlSelfApproval, got %v", err)
	}
	if _, err := engine.Reject(ctx, req.ID, "admin-a", "no"); err != domain.ErrDualControlSelfApproval {
		t.Fatalf("expected ErrDualControlSelfApproval on reject, got %v", err)
	}
}
