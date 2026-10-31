package infrastructure

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
)

// InMemoryCatalogRepository provides thread-safe in-memory storage for catalog data,
// pre-seeded with commercial reference applications.
type InMemoryCatalogRepository struct {
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

func NewInMemoryCatalogRepository() *InMemoryCatalogRepository {
	repo := &InMemoryCatalogRepository{
		apps:     make(map[string]*domain.CatalogApp),
		byID:     make(map[string]*domain.CatalogApp),
		locs:     make(map[string][]domain.Localization),
		media:    make(map[string][]domain.MediaAsset),
		dists:    make(map[string][]domain.Distribution),
		releases: make(map[string]*domain.Release),
		tags:     make(map[string][]domain.Tag),
		categories: []domain.Category{
			{ID: "productivity", ParentID: nil, NameI18n: map[string]string{"en-US": "Productivity", "ko-KR": "생산성"}, AppCount: 42},
			{ID: "utilities", ParentID: nil, NameI18n: map[string]string{"en-US": "Tools & Utilities", "ko-KR": "도구 및 유틸리티"}, AppCount: 35},
			{ID: "communication", ParentID: nil, NameI18n: map[string]string{"en-US": "Communication", "ko-KR": "커뮤니케이션"}, AppCount: 20},
			{ID: "finance", ParentID: nil, NameI18n: map[string]string{"en-US": "Finance", "ko-KR": "금융"}, AppCount: 18},
			{ID: "media_video", ParentID: nil, NameI18n: map[string]string{"en-US": "Media & Video", "ko-KR": "미디어 및 동영상"}, AppCount: 15},
			{ID: "games_action", ParentID: nil, NameI18n: map[string]string{"en-US": "Action Games", "ko-KR": "액션 게임"}, AppCount: 28},
			{ID: "games_rpg", ParentID: nil, NameI18n: map[string]string{"en-US": "Role Playing", "ko-KR": "롤플레잉"}, AppCount: 19},
		},
	}

	repo.seedReferenceData()
	return repo
}

func (r *InMemoryCatalogRepository) seedReferenceData() {
	now := time.Now().UTC()

	// 1. Omni Productivity Suite
	app1 := &domain.CatalogApp{
		ID:                 "app-omni-01",
		CanonicalSlug:      "omni-productivity-suite",
		PrimaryPackageName: "com.omniwork.pro",
		PrimaryCategoryID:  "productivity",
		ContentRating:      "ALL",
		DefaultLocale:      "en-US",
		CatalogStatus:      domain.CatalogStatusPublished,
		AverageRating:      4.85,
		BayesianRating:     4.82,
		InstallCount:       125000,
		LicenseSPDX:        "Apache-2.0",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	r.AddApp(
		app1,
		[]domain.Localization{
			{
				ID:               "loc-omni-en",
				AppID:            app1.ID,
				Locale:           "en-US",
				Title:            "Omni Productivity Suite",
				ShortDescription: "All-in-one workspace, offline docs, and team collaboration",
				FullDescription:  "Omni Productivity Suite empowers independent professionals with secure offline docs, seamless real-time syncing, and complete privacy.",
				WhatsNew:         "Performance optimizations and dark mode support.",
				PrivacyPolicyURL: "https://omniwork.local/privacy",
			},
			{
				ID:               "loc-omni-ko",
				AppID:            app1.ID,
				Locale:           "ko-KR",
				Title:            "옴니 프로덕티비티 스위트",
				ShortDescription: "올인원 업무 공간, 오프라인 문서 및 팀 협업 툴",
				FullDescription:  "옴니 프로덕티비티 스위트는 안전한 오프라인 문서 편집, 실시간 동기화, 완벽한 개인정보 보호를 제공합니다.",
				WhatsNew:         "성능 최적화 및 다크 모드 지원 추가.",
				PrivacyPolicyURL: "https://omniwork.local/privacy",
			},
		},
		[]domain.MediaAsset{
			{
				ID:               "media-omni-icon",
				AppID:            app1.ID,
				Locale:           "en-US",
				DeviceFormFactor: "PHONE",
				AssetType:        "ICON",
				URL:              "https://cdn.appmarket.local/icons/omni-pro.png",
				Width:            512,
				Height:           512,
			},
		},
		[]domain.Distribution{
			{
				ID:                "dist-omni-bn",
				AppID:             app1.ID,
				SourceType:        domain.SourceBNHosted,
				DistributionMode:  domain.ModeDirectAPK,
				PackageName:       "com.omniwork.pro",
				SigningCertSHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
		},
		map[string]*domain.Release{
			"dist-omni-bn": {
				ID:             "rel-omni-100",
				DistributionID: "dist-omni-bn",
				AppID:          app1.ID,
				ReleaseTrack:   "PRODUCTION",
				VersionCode:    100,
				VersionName:    "1.0.0",
				MinSDK:         26,
				TargetSDK:      35,
				SupportedABIs:  []string{"arm64-v8a", "armeabi-v7a"},
				FileSizeBytes:  24500100,
				ReleaseNotes:   "Initial stable release with full offline support.",
				PublishedAt:    &now,
			},
		},
		[]domain.Tag{
			{ID: "tag-office", NameI18n: map[string]string{"en-US": "Office", "ko-KR": "사무"}},
			{ID: "tag-cloud", NameI18n: map[string]string{"en-US": "Cloud", "ko-KR": "클라우드"}},
		},
	)

	// 2. Cloud Authenticator Pro
	app2 := &domain.CatalogApp{
		ID:                 "app-auth-02",
		CanonicalSlug:      "cloud-authenticator-pro",
		PrimaryPackageName: "com.securesuite.authenticator",
		PrimaryCategoryID:  "utilities",
		ContentRating:      "ALL",
		DefaultLocale:      "en-US",
		CatalogStatus:      domain.CatalogStatusPublished,
		AverageRating:      4.92,
		BayesianRating:     4.90,
		InstallCount:       89000,
		LicenseSPDX:        "MIT",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	r.AddApp(
		app2,
		[]domain.Localization{
			{
				ID:               "loc-auth-en",
				AppID:            app2.ID,
				Locale:           "en-US",
				Title:            "Cloud Authenticator Pro",
				ShortDescription: "Encrypted 2FA authenticator with hardware security key backup",
				FullDescription:  "Protect your accounts with hardware-backed encryption, encrypted cloud backup, and zero-knowledge privacy.",
				WhatsNew:         "FIDO2 WebAuthn passkey support.",
				PrivacyPolicyURL: "https://securesuite.local/privacy",
			},
		},
		nil,
		[]domain.Distribution{
			{
				ID:                "dist-auth-bn",
				AppID:             app2.ID,
				SourceType:        domain.SourceBNHosted,
				DistributionMode:  domain.ModeDirectAPK,
				PackageName:       "com.securesuite.authenticator",
				SigningCertSHA256: "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90",
			},
		},
		map[string]*domain.Release{
			"dist-auth-bn": {
				ID:             "rel-auth-210",
				DistributionID: "dist-auth-bn",
				AppID:          app2.ID,
				ReleaseTrack:   "PRODUCTION",
				VersionCode:    210,
				VersionName:    "2.1.0",
				MinSDK:         26,
				TargetSDK:      35,
				SupportedABIs:  []string{"arm64-v8a", "x86_64"},
				FileSizeBytes:  12400000,
				ReleaseNotes:   "Added WebAuthn passkeys and TOTP cloud sync.",
				PublishedAt:    &now,
			},
		},
		[]domain.Tag{{ID: "tag-security", NameI18n: map[string]string{"en-US": "Security", "ko-KR": "보안"}}},
	)
}

func (r *InMemoryCatalogRepository) AddApp(
	app *domain.CatalogApp,
	locs []domain.Localization,
	media []domain.MediaAsset,
	dists []domain.Distribution,
	releases map[string]*domain.Release,
	tags []domain.Tag,
) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.apps[app.CanonicalSlug] = app
	r.byID[app.ID] = app
	r.locs[app.ID] = locs
	r.media[app.ID] = media
	r.dists[app.ID] = dists
	for distID, rel := range releases {
		r.releases[distID] = rel
	}
	r.tags[app.ID] = tags
}

func (r *InMemoryCatalogRepository) ListPublishedApps(ctx context.Context, filter domain.AppListFilter) ([]domain.CatalogApp, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.CatalogApp
	for _, app := range r.byID {
		if app.CatalogStatus != domain.CatalogStatusPublished {
			continue
		}
		if filter.CategoryID != "" && !strings.EqualFold(app.PrimaryCategoryID, filter.CategoryID) {
			continue
		}
		if filter.MinRating > 0 && app.AverageRating < filter.MinRating {
			continue
		}
		if filter.TagID != "" {
			hasTag := false
			for _, t := range r.tags[app.ID] {
				if strings.EqualFold(t.ID, filter.TagID) || strings.EqualFold(t.NameI18n["en-US"], filter.TagID) {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}
		result = append(result, *app)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].InstallCount > result[j].InstallCount
	})

	limit := filter.Limit
	if limit <= 0 || limit > len(result) {
		limit = len(result)
	}

	return result[:limit], "", nil
}

func (r *InMemoryCatalogRepository) GetAppBySlug(ctx context.Context, slug string) (*domain.CatalogApp, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	app, ok := r.apps[slug]
	if !ok {
		return nil, domain.ErrAppNotFound
	}
	cp := *app
	return &cp, nil
}

func (r *InMemoryCatalogRepository) GetLocalizations(ctx context.Context, appID string) ([]domain.Localization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]domain.Localization, len(r.locs[appID]))
	copy(out, r.locs[appID])
	return out, nil
}

func (r *InMemoryCatalogRepository) GetMediaAssets(ctx context.Context, appID, locale string) ([]domain.MediaAsset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []domain.MediaAsset
	for _, m := range r.media[appID] {
		if locale == "" || m.Locale == locale {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *InMemoryCatalogRepository) GetDistributions(ctx context.Context, appID string) ([]domain.Distribution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]domain.Distribution, len(r.dists[appID]))
	copy(out, r.dists[appID])
	return out, nil
}

func (r *InMemoryCatalogRepository) GetLatestRelease(ctx context.Context, distributionID string) (*domain.Release, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rel, ok := r.releases[distributionID]
	if !ok {
		return nil, nil
	}
	cp := *rel
	return &cp, nil
}

func (r *InMemoryCatalogRepository) GetCategories(ctx context.Context) ([]domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]domain.Category, len(r.categories))
	copy(out, r.categories)
	return out, nil
}

func (r *InMemoryCatalogRepository) GetTags(ctx context.Context, appID string) ([]domain.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]domain.Tag, len(r.tags[appID]))
	copy(out, r.tags[appID])
	return out, nil
}
