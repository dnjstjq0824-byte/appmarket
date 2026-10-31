package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/infrastructure"
)

func TestIAMRepositories_Lifecycle(t *testing.T) {
	ctx := context.Background()

	// 1. User Repository
	userRepo := infrastructure.NewInMemoryUserRepository()

	// Test pre-seeded admin
	admin, err := userRepo.GetUserByID(ctx, "usr-admin-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("expected admin user to be found, got error: %v", err)
	}
	if admin.Email != "admin@appmarket.local" {
		t.Fatalf("expected admin@appmarket.local, got %s", admin.Email)
	}

	roles, _, err := userRepo.GetUserRoles(ctx, admin.ID)
	if err != nil || len(roles) == 0 || roles[0] != "ADMIN_SUPER" {
		t.Fatalf("expected ADMIN_SUPER role for admin, got %v", roles)
	}

	// Test create new user
	newUserID := "usr-new-0001"
	blind := infrastructure.ComputeBlindIndex("new@test.local", "seed-hmac-salt")
	newUser := &domain.User{
		ID:                  newUserID,
		Email:               "new@test.local",
		EmailBlindIndex:     blind,
		DisplayName:         "New User",
		Status:              domain.StatusActive,
		FailedLoginAttempts: 0,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}
	if err := userRepo.CreateUser(ctx, newUser); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	foundByBlind, err := userRepo.GetUserByEmailBlindIndex(ctx, blind)
	if err != nil || foundByBlind.ID != newUserID {
		t.Fatalf("expected user to be found by blind index, got %v, err=%v", foundByBlind, err)
	}

	// Update user status
	lockedTime := time.Now().UTC().Add(30 * time.Minute)
	if err := userRepo.UpdateUserStatus(ctx, newUserID, domain.StatusLockedBruteForce, 5, &lockedTime); err != nil {
		t.Fatalf("failed to update user status: %v", err)
	}

	updated, _ := userRepo.GetUserByID(ctx, newUserID)
	if updated.Status != domain.StatusLockedBruteForce || updated.FailedLoginAttempts != 5 {
		t.Fatalf("unexpected user state: %+v", updated)
	}

	// 2. Session Repository
	sessionRepo := infrastructure.NewInMemorySessionRepository()
	s1 := &domain.AuthSession{
		ID:                "sess-1",
		UserID:            newUserID,
		TokenFamilyID:     "fam-1",
		RefreshTokenHash:  "hash-1",
		DeviceFingerprint: "fp-1",
		ExpiresAt:         time.Now().UTC().Add(24 * time.Hour),
		CreatedAt:         time.Now().UTC(),
	}
	if err := sessionRepo.CreateSession(ctx, s1); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	foundSess, err := sessionRepo.GetSessionByTokenHash(ctx, "hash-1")
	if err != nil || foundSess.ID != "sess-1" {
		t.Fatalf("expected session-1, got %v, err=%v", foundSess, err)
	}

	// Rotate session
	s2 := &domain.AuthSession{
		ID:                "sess-2",
		UserID:            newUserID,
		TokenFamilyID:     "fam-1",
		RefreshTokenHash:  "hash-2",
		DeviceFingerprint: "fp-1",
		ExpiresAt:         time.Now().UTC().Add(24 * time.Hour),
		CreatedAt:         time.Now().UTC(),
	}
	if err := sessionRepo.RotateSession(ctx, "sess-1", s2); err != nil {
		t.Fatalf("failed to rotate session: %v", err)
	}

	// Verify old session rotated
	oldSess, _ := sessionRepo.GetSessionByTokenHash(ctx, "hash-1")
	if oldSess.RotatedAt == nil || oldSess.ReplacedByHash != "hash-2" {
		t.Fatalf("expected old session marked rotated, got %+v", oldSess)
	}

	// Revoke Family
	if err := sessionRepo.RevokeFamily(ctx, "fam-1"); err != nil {
		t.Fatalf("failed to revoke family: %v", err)
	}
	revokedSess, _ := sessionRepo.GetSessionByTokenHash(ctx, "hash-2")
	if !revokedSess.IsRevoked {
		t.Fatalf("expected session 2 revoked")
	}

	// 3. MFA Repository
	mfaRepo := infrastructure.NewInMemoryMFARepository()
	cred := &domain.MFACredential{
		ID:                "mfa-1",
		UserID:            newUserID,
		Type:              "TOTP",
		SecretEncrypted:   "encrypted-secret-blob",
		BackupCodesHashed: []string{"hashcode1", "hashcode2"},
		IsVerified:        true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := mfaRepo.SaveMFACredential(ctx, cred); err != nil {
		t.Fatalf("failed to save MFA credential: %v", err)
	}
	foundCred, err := mfaRepo.GetMFACredential(ctx, newUserID)
	if err != nil || foundCred == nil || foundCred.SecretEncrypted != "encrypted-secret-blob" {
		t.Fatalf("expected MFA cred recovered: %+v, err=%v", foundCred, err)
	}
}

