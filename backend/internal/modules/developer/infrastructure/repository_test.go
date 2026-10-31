package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/infrastructure"
)

func TestInMemoryDeveloperRepository_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewInMemoryDeveloperRepository()

	// 1. Create Org
	now := time.Now().UTC()
	org := &domain.DeveloperOrg{
		ID:                 "org-test-01",
		OrgSlug:            "acme-corp",
		OrgName:            "Acme Corp",
		VerificationStatus: domain.OrgUnverified,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := repo.CreateOrg(ctx, org); err != nil {
		t.Fatalf("CreateOrg failed: %v", err)
	}

	// 2. Get Org by ID & Slug
	fetched, err := repo.GetOrgByID(ctx, "org-test-01")
	if err != nil || fetched.OrgSlug != "acme-corp" {
		t.Fatalf("GetOrgByID failed: %v, org=%v", err, fetched)
	}
	bySlug, err := repo.GetOrgBySlug(ctx, "acme-corp")
	if err != nil || bySlug.ID != "org-test-01" {
		t.Fatalf("GetOrgBySlug failed: %v", err)
	}

	// 3. Update Org Status
	if err := repo.UpdateOrgStatus(ctx, "org-test-01", domain.OrgPendingReview); err != nil {
		t.Fatalf("UpdateOrgStatus failed: %v", err)
	}
	updated, _ := repo.GetOrgByID(ctx, "org-test-01")
	if updated.VerificationStatus != domain.OrgPendingReview {
		t.Errorf("expected PENDING_REVIEW, got %s", updated.VerificationStatus)
	}

	// 4. Add Member & RBAC Check
	member := &domain.OrgMember{
		OrgID:  "org-test-01",
		UserID: "usr-owner-01",
		Role:   domain.RoleDevOwner,
	}
	if err := repo.AddMember(ctx, member); err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}
	fetchedMember, err := repo.GetMember(ctx, "org-test-01", "usr-owner-01")
	if err != nil || fetchedMember.Role != domain.RoleDevOwner {
		t.Fatalf("GetMember failed: %v", err)
	}
	members, err := repo.ListMembers(ctx, "org-test-01")
	if err != nil || len(members) != 1 {
		t.Fatalf("ListMembers failed: %v", err)
	}

	// 5. API Keys
	key := &domain.APIKey{
		ID:        "key-01",
		OrgID:     "org-test-01",
		Name:      "CI Deploy Key",
		KeyHash:   "abc123hash",
		Scopes:    []string{"packages:write"},
		ExpiresAt: now.Add(24 * time.Hour),
		Revoked:   false,
		CreatedAt: now,
	}
	if err := repo.CreateAPIKey(ctx, key); err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}
	fetchedKey, err := repo.GetAPIKeyByHash(ctx, "abc123hash")
	if err != nil || fetchedKey.ID != "key-01" {
		t.Fatalf("GetAPIKeyByHash failed: %v", err)
	}
	if err := repo.RevokeAPIKey(ctx, "key-01"); err != nil {
		t.Fatalf("RevokeAPIKey failed: %v", err)
	}
	revokedKey, _ := repo.GetAPIKeyByHash(ctx, "abc123hash")
	if !revokedKey.Revoked {
		t.Errorf("expected key to be revoked")
	}
}

