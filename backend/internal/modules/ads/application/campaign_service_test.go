package application

import (
	"context"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/ads/domain"
)

func TestCampaignLifecycle(t *testing.T) {
	ctx := context.Background()
	svc := NewCampaignService(nil)
	c, err := svc.CreateCampaign(ctx, CreateCampaignInput{
		OrgID: "org-1", AppID: "app-1", AppName: "Demo App",
		TargetKeywords: []string{"launcher", "tools"}, MaxBidCPC: 300, DailyBudget: 50000,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if c.Status != domain.CampaignPendingReview {
		t.Fatalf("expected PENDING_REVIEW, got %s", c.Status)
	}
	if _, err := svc.ReviewCampaign(ctx, c.ID, true); err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	got, err := svc.GetCampaign(ctx, c.ID)
	if err != nil || got.Status != domain.CampaignActive {
		t.Fatalf("expected ACTIVE, got %+v / %v", got, err)
	}
	if _, err := svc.ReviewCampaign(ctx, c.ID, false); err != domain.ErrCampaignNotReviewable {
		t.Fatalf("expected not-reviewable on second review, got %v", err)
	}
	paused, err := svc.SetPaused(ctx, "org-1", c.ID, true)
	if err != nil || paused.Status != domain.CampaignPausedManual {
		t.Fatalf("expected pause, got %+v / %v", paused, err)
	}
	if _, err := svc.SetPaused(ctx, "org-2", c.ID, false); err != domain.ErrCampaignAccessDenied {
		t.Fatalf("expected access denied for other org, got %v", err)
	}
	resumed, err := svc.SetPaused(ctx, "org-1", c.ID, false)
	if err != nil || resumed.Status != domain.CampaignActive {
		t.Fatalf("expected resume, got %+v / %v", resumed, err)
	}
	if list := svc.ListCampaignsByOrg(ctx, "org-1"); len(list) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(list))
	}
	if pending := svc.ListCampaignsByStatus(ctx, domain.CampaignPendingReview); len(pending) != 0 {
		t.Fatalf("expected 0 pending, got %d", len(pending))
	}
}

func TestCampaignRejectAndValidation(t *testing.T) {
	ctx := context.Background()
	svc := NewCampaignService(nil)
	c, err := svc.CreateCampaign(ctx, CreateCampaignInput{
		OrgID: "org-1", AppID: "app-1", TargetKeywords: []string{"x"}, MaxBidCPC: 100, DailyBudget: 1000,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	rejected, err := svc.ReviewCampaign(ctx, c.ID, false)
	if err != nil || rejected.Status != domain.CampaignRejected {
		t.Fatalf("expected reject, got %+v / %v", rejected, err)
	}
	if _, err := svc.CreateCampaign(ctx, CreateCampaignInput{OrgID: "org-1"}); err == nil {
		t.Fatalf("expected validation error for empty campaign")
	}
	if _, err := svc.GetCampaign(ctx, "missing"); err != domain.ErrAdCampaignNotFound {
		t.Fatalf("expected not-found, got %v", err)
	}
}
