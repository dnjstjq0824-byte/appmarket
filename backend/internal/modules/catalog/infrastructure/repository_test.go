package infrastructure_test

import (
	"context"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/infrastructure"
)

func TestInMemoryCatalogRepository_CRUDAndFiltering(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewInMemoryCatalogRepository()

	// 1. Test ListPublishedApps
	apps, cursor, err := repo.ListPublishedApps(ctx, domain.AppListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListPublishedApps failed: %v", err)
	}
	if len(apps) < 2 {
		t.Fatalf("expected at least 2 seeded apps, got %d", len(apps))
	}
	_ = cursor

	// 2. Test GetAppBySlug
	app, err := repo.GetAppBySlug(ctx, "omni-productivity-suite")
	if err != nil {
		t.Fatalf("GetAppBySlug failed: %v", err)
	}
	if app.PrimaryPackageName != "com.omniwork.pro" {
		t.Errorf("expected package com.omniwork.pro, got %s", app.PrimaryPackageName)
	}

	// 3. Test GetLocalizations
	locs, err := repo.GetLocalizations(ctx, app.ID)
	if err != nil {
		t.Fatalf("GetLocalizations failed: %v", err)
	}
	if len(locs) < 2 {
		t.Errorf("expected 2 localizations (en, ko), got %d", len(locs))
	}

	// 4. Test GetCategories
	cats, err := repo.GetCategories(ctx)
	if err != nil {
		t.Fatalf("GetCategories failed: %v", err)
	}
	if len(cats) < 5 {
		t.Errorf("expected at least 5 categories, got %d", len(cats))
	}

	// 5. Test GetDistributions and LatestRelease
	dists, err := repo.GetDistributions(ctx, app.ID)
	if err != nil {
		t.Fatalf("GetDistributions failed: %v", err)
	}
	if len(dists) == 0 {
		t.Fatalf("expected at least 1 distribution")
	}
	rel, err := repo.GetLatestRelease(ctx, dists[0].ID)
	if err != nil {
		t.Fatalf("GetLatestRelease failed: %v", err)
	}
	if rel == nil || rel.VersionCode != 100 {
		t.Errorf("expected release versionCode 100, got %v", rel)
	}

	// 6. Test Filtering by Category
	prodApps, _, err := repo.ListPublishedApps(ctx, domain.AppListFilter{CategoryID: "productivity"})
	if err != nil {
		t.Fatalf("filter failed: %v", err)
	}
	if len(prodApps) != 1 || prodApps[0].CanonicalSlug != "omni-productivity-suite" {
		t.Errorf("expected 1 productivity app, got %d", len(prodApps))
	}
}

func TestInMemoryCache_SetGetExpiry(t *testing.T) {
	ctx := context.Background()
	cache := infrastructure.NewInMemoryCache()

	err := cache.Set(ctx, "test-key", []byte("test-payload"), 60)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := cache.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(val) != "test-payload" {
		t.Errorf("expected test-payload, got %s", string(val))
	}

	_, err = cache.Get(ctx, "non-existent")
	if err == nil {
		t.Errorf("expected error for non-existent key")
	}
}

