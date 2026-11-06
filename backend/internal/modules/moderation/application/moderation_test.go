package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/domain"
)

// TC-SEC-008: Verified-Installer Only Reviews & Anti-Fraud Review Detection
func TestTCSEC008VerifiedInstallerReviewsAndAntiFraud(t *testing.T) {
	ctx := context.Background()
	verifier := NewInMemoryInstallVerifier()
	detector := NewAntiFraudDetector()
	service := NewReviewService(verifier, detector)

	appID := "app-secure-vault"

	t.Run("Sub-test 1: Unverified user is blocked from submitting review", func(t *testing.T) {
		_, err := service.PostReview(ctx, CreateReviewInput{
			AppID:           appID,
			UserID:          "user-not-installed",
			Rating:          5,
			Title:           "Great app!",
			Comment:         "Works nicely",
			VersionCode:     100,
			ClientIP:        "192.168.1.10",
			AccountAgeHours: 48.0,
		})

		if !errors.Is(err, domain.ErrReviewRequiresInstallation) {
			t.Fatalf("expected ErrReviewRequiresInstallation for unverified user, got %v", err)
		}
	})

	t.Run("Sub-test 2: Verified installer submits review successfully with version tag", func(t *testing.T) {
		verifier.RecordInstallation("user-verified", appID)

		rev, err := service.PostReview(ctx, CreateReviewInput{
			AppID:           appID,
			UserID:          "user-verified",
			Rating:          5,
			Title:           "Legit and smooth",
			Comment:         "Installed and running perfectly on Android 14",
			VersionCode:     102,
			ClientIP:        "192.168.1.20",
			AccountAgeHours: 120.0,
		})

		if err != nil {
			t.Fatalf("expected review post success, got %v", err)
		}
		if !rev.IsVerifiedInstall {
			t.Fatalf("expected IsVerifiedInstall=true")
		}
		if rev.VersionCode != 102 {
			t.Fatalf("expected versionCode 102, got %d", rev.VersionCode)
		}
		if rev.Status != domain.ReviewPublished {
			t.Fatalf("expected ReviewPublished for clean review, got %s", rev.Status)
		}
	})

	t.Run("Sub-test 2b: Stored review markup is stripped", func(t *testing.T) {
		rev, err := service.PostReview(ctx, CreateReviewInput{
			AppID:           appID,
			UserID:          "user-verified",
			Rating:          4,
			Title:           `<script>alert(1)</script>Good`,
			Comment:         `Nice <a href="javascript:alert(2)">link</a> app`,
			VersionCode:     102,
			ClientIP:        "192.168.1.21",
			AccountAgeHours: 120.0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(rev.Title, "<script") || strings.Contains(rev.Comment, "javascript:") {
			t.Fatalf("stored XSS survived: title=%q comment=%q", rev.Title, rev.Comment)
		}
	})

	t.Run("Sub-test 3: Brand new account submitting 1-star gets flagged suspicious", func(t *testing.T) {
		verifier.RecordInstallation("user-new-spammer", appID)

		rev, err := service.PostReview(ctx, CreateReviewInput{
			AppID:           appID,
			UserID:          "user-new-spammer",
			Rating:          1,
			Title:           "Terrible!",
			Comment:         "Hate this app completely",
			VersionCode:     102,
			ClientIP:        "192.168.1.30",
			AccountAgeHours: 0.5, // Account created 30 mins ago
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rev.Status != domain.ReviewFlaggedSuspicious {
			t.Fatalf("expected ReviewFlaggedSuspicious for 30m old account 1-star, got %s", rev.Status)
		}
	})
}

// TC-FUNC-013: Rejection Appeal Workflow & Dual-Control Separation
func TestTCFUNC013AppealWorkflowDualControl(t *testing.T) {
	ctx := context.Background()
	appealService := NewAppealService()

	originalReviewer := "admin_reviewer_original@bookingnetwork.com"
	seniorReviewer := "admin_senior_audit@bookingnetwork.com"

	appeal, err := appealService.SubmitAppeal(ctx, "org-game", "rel-rejected-101", "False positive permission flag", originalReviewer)
	if err != nil {
		t.Fatalf("failed to submit appeal: %v", err)
	}

	t.Run("Sub-test 1: Original reviewer cannot judge their own appeal", func(t *testing.T) {
		_, err := appealService.DecideAppeal(ctx, appeal.ID, originalReviewer, domain.AppealApproved, "Reconsidering my initial decision")
		if !errors.Is(err, domain.ErrAppealSameReviewerForbidden) {
			t.Fatalf("expected ErrAppealSameReviewerForbidden, got %v", err)
		}
	})

	t.Run("Sub-test 2b: Stored appeal text is sanitized", func(t *testing.T) {
		appeal2, err := appealService.SubmitAppeal(ctx, "org-x", "rel-y", `<img src=x onerror=alert(1)> clean reason`, originalReviewer)
		if err != nil {
			t.Fatalf("submit failed: %v", err)
		}
		if strings.Contains(appeal2.Reason, "onerror") || strings.Contains(appeal2.Reason, "<img") {
			t.Fatalf("stored XSS survived in appeal reason: %q", appeal2.Reason)
		}
	})

	t.Run("Sub-test 2: Distinct senior reviewer decides appeal successfully", func(t *testing.T) {
		decidedAppeal, err := appealService.DecideAppeal(ctx, appeal.ID, seniorReviewer, domain.AppealApproved, "Permission is justified for NFC wallet")
		if err != nil {
			t.Fatalf("expected senior approval success, got %v", err)
		}
		if decidedAppeal.Status != domain.AppealApproved {
			t.Fatalf("expected AppealApproved, got %s", decidedAppeal.Status)
		}
		if decidedAppeal.AssignedReviewerID != seniorReviewer {
			t.Fatalf("expected assigned reviewer %s, got %s", seniorReviewer, decidedAppeal.AssignedReviewerID)
		}
	})
}

