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

func TestListPublishedApps_QueryAndFlagFilters(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewInMemoryCatalogRepository()

	matched, _, err := repo.ListPublishedApps(ctx, domain.AppListFilter{Query: "omni"})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(matched) != 1 || matched[0].ID != "app-omni-01" {
		t.Fatalf("expected omni match, got %+v", matched)
	}
	matched, _, err = repo.ListPublishedApps(ctx, domain.AppListFilter{Query: "com.securesuite"})
	if err != nil || len(matched) != 1 || matched[0].ID != "app-auth-02" {
		t.Fatalf("expected package match, got %+v / %v", matched, err)
	}
	matched, _, err = repo.ListPublishedApps(ctx, domain.AppListFilter{Query: "authenticator"})
	if err != nil || len(matched) != 1 || matched[0].ID != "app-auth-02" {
		t.Fatalf("expected title match, got %+v / %v", matched, err)
	}
	matched, _, err = repo.ListPublishedApps(ctx, domain.AppListFilter{Query: "no-such-app-xyz"})
	if err != nil || len(matched) != 0 {
		t.Fatalf("expected empty query result, got %+v / %v", matched, err)
	}
	free, _, err := repo.ListPublishedApps(ctx, domain.AppListFilter{FreeOnly: true})
	if err != nil || len(free) != 2 {
		t.Fatalf("expected 2 free apps, got %d / %v", len(free), err)
	}
	direct, _, err := repo.ListPublishedApps(ctx, domain.AppListFilter{DirectOnly: true})
	if err != nil || len(direct) != 2 {
		t.Fatalf("expected 2 direct apps, got %d / %v", len(direct), err)
	}
}

func TestListAppsByOrgAndUpsertLocalization(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewInMemoryCatalogRepository()
	orgID := "org-test-01"
	repo.AddApp(
		&domain.CatalogApp{ID: "app-org-01", CanonicalSlug: "org-app", PrimaryPackageName: "com.org.app", DeveloperOrgID: &orgID, CatalogStatus: domain.CatalogStatusPublished},
		[]domain.Localization{{ID: "loc-1", AppID: "app-org-01", Locale: "en-US", Title: "Old"}},
		nil, nil, nil, nil,
	)
	apps, err := repo.ListAppsByOrg(ctx, orgID)
	if err != nil || len(apps) != 1 {
		t.Fatalf("expected 1 org app, got %d / %v", len(apps), err)
	}
	saved, err := repo.UpsertLocalization(ctx, domain.Localization{AppID: "app-org-01", Locale: "en-US", Title: "New", ShortDescription: "s", FullDescription: "f"})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	if saved.Title != "New" || saved.ID != "loc-1" {
		t.Fatalf("expected in-place update, got %+v", saved)
	}
	added, err := repo.UpsertLocalization(ctx, domain.Localization{AppID: "app-org-01", Locale: "ko-KR", Title: "제목", ShortDescription: "s", FullDescription: "f"})
	if err != nil || added.ID == "" {
		t.Fatalf("expected ko-KR insert, got %+v / %v", added, err)
	}
	locs, err := repo.GetLocalizations(ctx, "app-org-01")
	if err != nil || len(locs) != 2 {
		t.Fatalf("expected 2 locales, got %d / %v", len(locs), err)
	}
	if _, err := repo.UpsertLocalization(ctx, domain.Localization{AppID: "missing", Locale: "en-US"}); err != domain.ErrAppNotFound {
		t.Fatalf("expected app-not-found, got %v", err)
	}
}

func TestListPublishedApps_CursorPaginationRoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewInMemoryCatalogRepository()

	page1, cursor1, err := repo.ListPublishedApps(ctx, domain.AppListFilter{Limit: 1})
	if err != nil {
		t.Fatalf("page 1 failed: %v", err)
	}
	if len(page1) != 1 {
		t.Fatalf("expected 1 app on page 1, got %d", len(page1))
	}
	if cursor1 == "" {
		t.Fatalf("expected non-empty next cursor when more results remain")
	}

	page2, cursor2, err := repo.ListPublishedApps(ctx, domain.AppListFilter{Limit: 1, Cursor: cursor1})
	if err != nil {
		t.Fatalf("page 2 failed: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("expected 1 app on page 2, got %d", len(page2))
	}
	if page2[0].ID == page1[0].ID {
		t.Fatalf("cursor did not advance: page 2 repeats %s", page1[0].ID)
	}
	if cursor2 != "" {
		t.Fatalf("expected empty cursor at end of results, got %q", cursor2)
	}

	var walked []string
	cursor := ""
	for {
		page, next, err := repo.ListPublishedApps(ctx, domain.AppListFilter{Limit: 1, Cursor: cursor})
		if err != nil {
			t.Fatalf("walk failed: %v", err)
		}
		for _, a := range page {
			walked = append(walked, a.ID)
		}
		if next == "" {
			break
		}
		cursor = next
	}
	if len(walked) != 2 {
		t.Fatalf("expected full walk to visit 2 seeded apps, visited %v", walked)
	}
	if walked[0] == walked[1] {
		t.Fatalf("walk visited duplicate app %s", walked[0])
	}
}

