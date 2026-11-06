package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
)

// Admin login without an MFA credential must report that enrollment is required,
// and enroll -> verify-confirm must yield MFA-verified session tokens.
func TestMFAEnrollmentFlow(t *testing.T) {
	ctx := context.Background()
	svc, uRepo, _, _, _ := setupTestAuthService(t)

	user, err := svc.Register(ctx, application.RegisterInput{
		Email:       "admin-enroll@appmarket.local",
		Password:    "SecureP@ssword123",
		DisplayName: "Enroll Admin",
	})
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	uRepo.roles[user.ID] = []string{"ADMIN_SECURITY"}

	loginRes, err := svc.Login(ctx, "admin-enroll@appmarket.local", "SecureP@ssword123", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if !loginRes.MFARequired || !loginRes.MFAEnrollmentRequired {
		t.Fatalf("expected MFARequired+enrollment flags, got %+v", loginRes)
	}

	enrolled, err := svc.EnrollMFA(ctx, "admin-enroll@appmarket.local", "SecureP@ssword123")
	if err != nil {
		t.Fatalf("enroll failed: %v", err)
	}
	if enrolled.Secret == "" || enrolled.ProvisioningURI == "" || len(enrolled.RecoveryCodes) != 10 {
		t.Fatalf("enroll returned incomplete material: %+v", enrolled)
	}

	mfaSvc := application.NewMFAService("BN App Market")
	code, err := mfaSvc.ComputeTOTP(enrolled.Secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("compute totp: %v", err)
	}
	final, err := svc.CompleteLoginWithMFA(ctx, user.ID, code, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("verify-confirm failed: %v", err)
	}
	if final.AccessToken == "" || final.RefreshToken == "" {
		t.Fatalf("expected session tokens after confirm, got %+v", final)
	}

	if _, err := svc.EnrollMFA(ctx, "admin-enroll@appmarket.local", "SecureP@ssword123"); err != domain.ErrMFAAlreadyEnrolled {
		t.Fatalf("expected ErrMFAAlreadyEnrolled on second enroll, got %v", err)
	}
}

func TestMFAEnrollRejectsBadPassword(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _ := setupTestAuthService(t)

	if _, err := svc.Register(ctx, application.RegisterInput{
		Email:       "victim@appmarket.local",
		Password:    "SecureP@ssword123",
		DisplayName: "Victim",
	}); err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	if _, err := svc.EnrollMFA(ctx, "victim@appmarket.local", "WrongPassword1!"); err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
