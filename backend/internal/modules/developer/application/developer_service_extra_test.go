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
