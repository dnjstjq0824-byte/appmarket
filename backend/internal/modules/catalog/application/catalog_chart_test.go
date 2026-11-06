package application_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
)

func TestGetChartOrdersAndValidates(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := setupCatalog(t)
	low := seedApp(repo, "low", "en-US", []domain.Localization{{ID: "l1", AppID: "app-low", Locale: "en-US", Title: "Low"}}, nil)
	low.AverageRating = 3.0
	low.BayesianRating = 2.9
	low.InstallCount = 10
	high := seedApp(repo, "high", "en-US", []domain.Localization{{ID: "l2", AppID: "app-high", Locale: "en-US", Title: "High"}}, nil)
	high.AverageRating = 4.9
	high.BayesianRating = 4.85
	high.InstallCount = 50000

	for _, chart := range []string{"top_free", "top_paid", "trending", "top_grossing"} {
		items, err := svc.GetChart(ctx, chart)
		if err != nil {
			t.Fatalf("%s failed: %v", chart, err)
		}
		if len(items) != 2 {
			t.Fatalf("%s: expected 2 items, got %d", chart, len(items))
		}
	}
	trending, err := svc.GetChart(ctx, "trending")
	if err != nil {
		t.Fatalf("trending failed: %v", err)
	}
	if trending[0].CanonicalSlug != "high" {
		t.Fatalf("expected high-rated first, got %s", trending[0].CanonicalSlug)
	}
	if _, err := svc.GetChart(ctx, "nope"); err != domain.ErrInvalidFilter {
		t.Fatalf("expected invalid filter, got %v", err)
	}
}

func TestSaveLocalizationSanitizes(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := setupCatalog(t)
	seedApp(repo, "loc", "en-US", nil, nil)
	saved, err := svc.SaveLocalization(ctx, domain.Localization{
		AppID: "app-loc", Locale: "en-US", Title: "<b>Title</b>",
		ShortDescription: "ok", FullDescription: `<img src=x onerror=alert(1)>desc`, WhatsNew: "new",
	})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if strings.Contains(saved.FullDescription, "onerror") || strings.Contains(saved.Title, "<b>") {
		t.Fatalf("localization not sanitized: %+v", saved)
	}
	if _, err := svc.SaveLocalization(ctx, domain.Localization{AppID: "app-loc", Locale: "en-US"}); err == nil {
		t.Fatalf("expected title validation error")
	}
}

func TestListAppsByOrgService(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := setupCatalog(t)
	orgID := "org-chart-01"
	app := seedApp(repo, "orgapp", "en-US", nil, nil)
	app.DeveloperOrgID = &orgID
	list, err := svc.ListAppsByOrg(ctx, orgID)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 org app, got %d / %v", len(list), err)
	}
}
