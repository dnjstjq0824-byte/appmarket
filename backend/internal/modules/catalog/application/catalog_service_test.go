package application_test

import (
	"context"
	"sync"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
)

type mockCatalogRepo struct {
	mu         sync.RWMutex
	apps       map[string]*domain.CatalogApp
	byID       map[string]*domain.CatalogApp
	locs       map[string][]domain.Localization
	media      map[string][]domain.MediaAsset
	dists      map[string][]domain.Distribution
	releases   map[string]*domain.Release
	categories []domain.Category
	tags       map[string][]domain.Tag
}

func newMockCatalogRepo() *mockCatalogRepo {
	return &mockCatalogRepo{
		apps:     make(map[string]*domain.CatalogApp),
		byID:     make(map[string]*domain.CatalogApp),
		locs:     make(map[string][]domain.Localization),
		media:    make(map[string][]domain.MediaAsset),
		dists:    make(map[string][]domain.Distribution),
		releases: make(map[string]*domain.Release),
		tags:     make(map[string][]domain.Tag),
	}
}

func (r *mockCatalogRepo) ListPublishedApps(ctx context.Context, filter domain.AppListFilter) ([]domain.CatalogApp, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.CatalogApp
	for _, a := range r.byID {
		if a.CatalogStatus == domain.CatalogStatusPublished {
			out = append(out, *a)
		}
	}
	return out, "", nil
}
func (r *mockCatalogRepo) GetAppBySlug(ctx context.Context, slug string) (*domain.CatalogApp, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.apps[slug]
	if !ok {
		return nil, domain.ErrAppNotFound
	}
	cp := *a
	return &cp, nil
}
func (r *mockCatalogRepo) GetLocalizations(ctx context.Context, appID string) ([]domain.Localization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Localization, len(r.locs[appID]))
	copy(out, r.locs[appID])
	return out, nil
}
func (r *mockCatalogRepo) GetMediaAssets(ctx context.Context, appID, locale string) ([]domain.MediaAsset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.MediaAsset
	for _, m := range r.media[appID] {
		if m.Locale == locale || locale == "" {
			out = append(out, m)
		}
	}
	return out, nil
}
func (r *mockCatalogRepo) GetDistributions(ctx context.Context, appID string) ([]domain.Distribution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Distribution, len(r.dists[appID]))
	copy(out, r.dists[appID])
	return out, nil
}
func (r *mockCatalogRepo) GetLatestRelease(ctx context.Context, distributionID string) (*domain.Release, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rel, ok := r.releases[distributionID]
	if !ok {
		return nil, nil
	}
	cp := *rel
	return &cp, nil
}
func (r *mockCatalogRepo) GetCategories(ctx context.Context) ([]domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.categories, nil
}
func (r *mockCatalogRepo) GetTags(ctx context.Context, appID string) ([]domain.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tags[appID], nil
}

type mockCache struct {
	mu    sync.RWMutex
	store map[string][]byte
}

func newMockCache() *mockCache { return &mockCache{store: make(map[string][]byte)} }

func (c *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.store[key]
	if !ok {
		return nil, domain.ErrCacheUnavailable
	}
	cp := make([]byte, len(v))
	copy(cp, v)
	return cp, nil
}
func (c *mockCache) Set(ctx context.Context, key string, value []byte, ttlSeconds int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]byte, len(value))
	copy(cp, value)
	c.store[key] = cp
	return nil
}

func setupCatalog(t *testing.T) (*application.CatalogService, *mockCatalogRepo, *mockCache) {
	t.Helper()
	repo := newMockCatalogRepo()
	cache := newMockCache()
	svc := application.NewCatalogService(repo, cache)
	return svc, repo, cache
}

func seedApp(repo *mockCatalogRepo, slug, defaultLocale string, locs []domain.Localization, dists []domain.Distribution) *domain.CatalogApp {
	app := &domain.CatalogApp{
		ID:                 "app-" + slug,
		CanonicalSlug:      slug,
		PrimaryPackageName: "com.example." + slug,
		PrimaryCategoryID:  "apps",
		DefaultLocale:      defaultLocale,
		CatalogStatus:      domain.CatalogStatusPublished,
		AverageRating:      4.5,
		BayesianRating:     4.3,
		InstallCount:       1000,
	}
	repo.apps[slug] = app
	repo.byID[app.ID] = app
	repo.locs[app.ID] = locs
	repo.dists[app.ID] = dists
	return app
}

func TestCatalogLocaleFallback(t *testing.T) {
	svc, repo, _ := setupCatalog(t)
	locs := []domain.Localization{
		{ID: "l1", AppID: "app-demo", Locale: "ko-KR", Title: "데모 앱", FullDescription: "한국어 설명"},
		{ID: "l2", AppID: "app-demo", Locale: "en-US", Title: "Demo App", FullDescription: "English description"},
	}
	dists := []domain.Distribution{
		{ID: "d1", AppID: "app-demo", SourceType: domain.SourceBNHosted, DistributionMode: domain.ModeDirectAPK, PackageName: "com.example.demo", SigningCertSHA256: "cert-aaa"},
	}
	seedApp(repo, "demo", "en-US", locs, dists)

	detail, err := svc.GetAppDetail(context.Background(), "demo", "ko-KR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.ResolvedLocale != "ko-KR" {
		t.Fatalf("expected resolved locale ko-KR, got %s", detail.ResolvedLocale)
	}
	if detail.Localization == nil || detail.Localization.Title != "데모 앱" {
		t.Fatalf("expected Korean title, got %+v", detail.Localization)
	}

	detail2, err := svc.GetAppDetail(context.Background(), "demo", "fr-FR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail2.ResolvedLocale != "en-US" {
		t.Fatalf("expected fallback to en-US, got %s", detail2.ResolvedLocale)
	}
	if detail2.Localization == nil || detail2.Localization.Title != "Demo App" {
		t.Fatalf("expected English title after fallback, got %+v", detail2.Localization)
	}
}
