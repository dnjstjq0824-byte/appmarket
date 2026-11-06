package application_test

import (
	"context"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
)

func TestDevOrgIDORBlocked(t *testing.T) {
	orgs, members, _, _ := setupDev(t)
	ctx := context.Background()
	orgA, _ := orgs.CreateOrg(ctx, "orga", "Org A")
	orgB, _ := orgs.CreateOrg(ctx, "orgb", "Org B")
	if _, err := members.AddMember(ctx, orgA.ID, "user1", domain.RoleDevOwner); err != nil {
		t.Fatalf("add member failed: %v", err)
	}
	if err := members.CheckPermission(ctx, orgB.ID, "user1", "app:create"); err != domain.ErrForbiddenTenantAccess {
		t.Fatalf("expected forbidden tenant access, got %v", err)
	}
	if err := members.CheckPermission(ctx, orgA.ID, "user1", "app:create"); err != nil {
		t.Fatalf("expected ok for own org, got %v", err)
	}
}

func TestDevRolePermissionMatrix(t *testing.T) {
	cases := []struct {
		role domain.DevRole
		perm string
		ok   bool
	}{
		{domain.RoleListingEditor, "listing:edit", true},
		{domain.RoleListingEditor, "release:upload", false},
		{domain.RoleSupportAgent, "review:reply", true},
		{domain.RoleSupportAgent, "finance:view", false},
		{domain.RoleFinanceManager, "finance:view", true},
		{domain.RoleFinanceManager, "release:submit", false},
	}
	for _, c := range cases {
		got := domain.RoleHasPermission(c.role, c.perm)
		if got != c.ok {
			t.Fatalf("role %s perm %s: got %v want %v", c.role, c.perm, got, c.ok)
		}
	}
}

func TestDevAPIKeyLifecycle(t *testing.T) {
	orgs, _, keys, _ := setupDev(t)
	ctx := context.Background()
	org, _ := orgs.CreateOrg(ctx, "keyorg", "Key Org")
	raw, k, err := keys.IssueAPIKey(ctx, org.ID, nil, "ci-key", []string{"release:upload"}, 30)
	if err != nil {
		t.Fatalf("issue key failed: %v", err)
	}
	if raw == "" {
		t.Fatal("raw key empty")
	}
	got, err := keys.ValidateAPIKey(ctx, raw)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if got.ID != k.ID {
		t.Fatalf("key mismatch")
	}
	if !got.HasScope("release:upload") {
		t.Fatal("scope missing")
	}
	if err := keys.RevokeAPIKey(ctx, k.ID); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}
	if _, err := keys.ValidateAPIKey(ctx, raw); err == nil {
		t.Fatal("expected validation failure after revoke")
	}
}

func TestDevMemberRemoveLifecycle(t *testing.T) {
	orgsSvc, members, _, _ := setupDev(t)
	ctx := context.Background()
	org, _ := orgsSvc.CreateOrg(ctx, "rmorg", "Remove Org")
	if _, err := members.AddMember(ctx, org.ID, "owner1", domain.RoleDevOwner); err != nil {
		t.Fatalf("add owner failed: %v", err)
	}
	if _, err := members.AddMember(ctx, org.ID, "editor1", domain.RoleListingEditor); err != nil {
		t.Fatalf("add editor failed: %v", err)
	}
	if err := members.RemoveMember(ctx, org.ID, "ghost"); err != domain.ErrMemberNotFound {
		t.Fatalf("expected member not found, got %v", err)
	}
	if err := members.RemoveMember(ctx, org.ID, "owner1"); err != domain.ErrLastOwnerRemoval {
		t.Fatalf("expected last owner guard, got %v", err)
	}
	if _, err := members.AddMember(ctx, org.ID, "owner2", domain.RoleDevOwner); err != nil {
		t.Fatalf("add second owner failed: %v", err)
	}
	if err := members.RemoveMember(ctx, org.ID, "owner1"); err != nil {
		t.Fatalf("remove with co-owner failed: %v", err)
	}
	if err := members.RemoveMember(ctx, org.ID, "editor1"); err != nil {
		t.Fatalf("remove editor failed: %v", err)
	}
	list, err := members.ListMembers(ctx, org.ID)
	if err != nil || len(list) != 1 || list[0].UserID != "owner2" {
		t.Fatalf("unexpected members after removal: %+v err=%v", list, err)
	}
}

func TestOrgDecideVerification(t *testing.T) {
	orgs, _, _, _ := setupDev(t)
	ctx := context.Background()
	org, _ := orgs.CreateOrg(ctx, "kyborg", "KYB Org")
	if _, err := orgs.SubmitVerification(ctx, org.ID, "BUSINESS_REGISTRATION", "https://docs.local/br.pdf", "kyb.example"); err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	decided, err := orgs.DecideVerification(ctx, org.ID, true, "docs valid", "admin-1")
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if decided.VerificationStatus != domain.OrgVerified {
		t.Fatalf("expected VERIFIED org, got %s", decided.VerificationStatus)
	}
	org2, _ := orgs.CreateOrg(ctx, "kyborg2", "KYB Org 2")
	if _, err := orgs.SubmitVerification(ctx, org2.ID, "BUSINESS_REGISTRATION", "https://docs.local/br.pdf", "kyb2.example"); err != nil {
		t.Fatalf("submit2 failed: %v", err)
	}
	decided2, err := orgs.DecideVerification(ctx, org2.ID, false, "expired docs", "admin-1")
	if err != nil {
		t.Fatalf("reject failed: %v", err)
	}
	if decided2.VerificationStatus != domain.OrgRejected {
		t.Fatalf("expected REJECTED org, got %s", decided2.VerificationStatus)
	}
	if _, err := orgs.DecideVerification(ctx, "missing", true, "", "admin-1"); err != domain.ErrOrgNotFound {
		t.Fatalf("expected org not found, got %v", err)
	}
	org3, _ := orgs.CreateOrg(ctx, "kyborg3", "KYB Org 3")
	if _, err := orgs.DecideVerification(ctx, org3.ID, true, "", "admin-1"); err != domain.ErrVerificationNotFound {
		t.Fatalf("expected verification not found, got %v", err)
	}
	queue, err := orgs.ListVerificationsByStatus(ctx, string(domain.OrgPendingReview))
	if err != nil {
		t.Fatalf("queue failed: %v", err)
	}
	if len(queue) != 0 {
		t.Fatalf("expected empty pending queue after decisions, got %d", len(queue))
	}
}

func TestDevOrgDuplicateSlug(t *testing.T) {
	orgs, _, _, _ := setupDev(t)
	ctx := context.Background()
	if _, err := orgs.CreateOrg(ctx, "dup", "Org 1"); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	if _, err := orgs.CreateOrg(ctx, "dup", "Org 2"); err != domain.ErrDuplicateOrgSlug {
		t.Fatalf("expected duplicate slug error, got %v", err)
	}
}
