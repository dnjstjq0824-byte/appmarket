package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/sync/singleflight"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
)

// CatalogService implements catalog query use cases.
type CatalogService struct {
	repo     CatalogRepository
	cache    Cache
	policy   *bluemonday.Policy
	sfGroup  singleflight.Group
	cacheTTL int
}

// NewCatalogService constructs a CatalogService.
func NewCatalogService(repo CatalogRepository, cache Cache) *CatalogService {
	return &CatalogService{
		repo:     repo,
		cache:    cache,
		policy:   bluemonday.StrictPolicy(),
		cacheTTL: 300,
	}
}

// SanitizeHTML sanitizes untrusted HTML using bluemonday StrictPolicy.
func (s *CatalogService) SanitizeHTML(input string) string {
	return s.policy.Sanitize(input)
}

// LocaleFallbackChain returns the ordered list of locales to try.
func LocaleFallbackChain(requested, defaultLocale string) []string {
	chain := []string{}
	seen := map[string]bool{}
	add := func(l string) {
		l = strings.TrimSpace(l)
		if l == "" || seen[l] {
			return
		}
		seen[l] = true
		chain = append(chain, l)
	}
	add(requested)
	if idx := strings.Index(requested, "-"); idx > 0 {
		add(requested[:idx])
	}
	add(defaultLocale)
	if idx := strings.Index(defaultLocale, "-"); idx > 0 {
		add(defaultLocale[:idx])
	}
	add("en")
	add("en-US")
	return chain
}

// ResolveLocalization picks the best localization given a fallback chain.
func ResolveLocalization(localizations []domain.Localization, requested, defaultLocale string) (*domain.Localization, string) {
	ptrs := map[string]*domain.Localization{}
	for i := range localizations {
		lc := localizations[i]
		ptrs[lc.Locale] = &localizations[i]
	}
	for _, loc := range LocaleFallbackChain(requested, defaultLocale) {
		if l, ok := ptrs[loc]; ok {
			return l, loc
		}
		for k, v := range ptrs {
			if strings.EqualFold(k, loc) {
				return v, k
			}
		}
	}
	if len(localizations) > 0 {
		return &localizations[0], localizations[0].Locale
	}
	return nil, ""
}

// ListApps returns a page of published apps.
func (s *CatalogService) ListApps(ctx context.Context, filter domain.AppListFilter) ([]domain.AppSummary, string, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	apps, cursor, err := s.repo.ListPublishedApps(ctx, filter)
	if err != nil {
		return nil, "", err
	}
	summaries := make([]domain.AppSummary, 0, len(apps))
	for _, a := range apps {
		summaries = append(summaries, domain.AppSummary{
			ID:                 a.ID,
			CanonicalSlug:      a.CanonicalSlug,
			Title:              a.CanonicalSlug,
			PrimaryPackageName: a.PrimaryPackageName,
			PrimaryCategoryID:  a.PrimaryCategoryID,
			AverageRating:      a.AverageRating,
			BayesianRating:     a.BayesianRating,
			InstallCount:       a.InstallCount,
		})
	}
	return summaries, cursor, nil
}

// GetAppDetail returns aggregated app detail with locale fallback, distributions, releases.
func (s *CatalogService) GetAppDetail(ctx context.Context, slug, requestedLocale string) (*domain.AppDetail, error) {
	cacheKey := fmt.Sprintf("catalog:app:%s:%s", slug, requestedLocale)
	if s.cache != nil {
		if b, err := s.cache.Get(ctx, cacheKey); err == nil && len(b) > 0 {
			var detail domain.AppDetail
			if err := json.Unmarshal(b, &detail); err == nil {
				return &detail, nil
			}
		}
	}

	v, err, _ := s.sfGroup.Do(cacheKey, func() (interface{}, error) {
		return s.loadAppDetail(ctx, slug, requestedLocale)
	})
	if err != nil {
		return nil, err
	}
	detail := v.(*domain.AppDetail)

	if s.cache != nil {
		if b, err := json.Marshal(detail); err == nil {
			_ = s.cache.Set(ctx, cacheKey, b, s.cacheTTL)
		}
	}
	return detail, nil
}

func (s *CatalogService) loadAppDetail(ctx context.Context, slug, requestedLocale string) (*domain.AppDetail, error) {
	app, err := s.repo.GetAppBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, domain.ErrAppNotFound
	}

	locs, err := s.repo.GetLocalizations(ctx, app.ID)
	if err != nil {
		return nil, err
	}
	loc, resolved := ResolveLocalization(locs, requestedLocale, app.DefaultLocale)
	if loc != nil {
		loc.ShortDescription = s.SanitizeHTML(loc.ShortDescription)
		loc.FullDescription = s.SanitizeHTML(loc.FullDescription)
		loc.WhatsNew = s.SanitizeHTML(loc.WhatsNew)
		loc.Title = s.SanitizeHTML(loc.Title)
	}

	media, _ := s.repo.GetMediaAssets(ctx, app.ID, resolved)

	dists, err := s.repo.GetDistributions(ctx, app.ID)
	if err != nil {
		return nil, err
	}
	for i := range dists {
		rel, err := s.repo.GetLatestRelease(ctx, dists[i].ID)
		if err == nil {
			dists[i].LatestRelease = rel
		}
		dists[i].SignatureCompatibilityGroup = dists[i].SigningCertSHA256
	}

	tags, _ := s.repo.GetTags(ctx, app.ID)

	detail := &domain.AppDetail{
		App:            *app,
		Localization:   loc,
		MediaAssets:    media,
		Distributions:  dists,
		Tags:           tags,
		ResolvedLocale: resolved,
	}
	if media == nil {
		detail.MediaAssets = []domain.MediaAsset{}
	}
	if dists == nil {
		detail.Distributions = []domain.Distribution{}
	}
	if tags == nil {
		detail.Tags = []domain.Tag{}
	}
	return detail, nil
}

// GetCategories returns all categories.
func (s *CatalogService) GetCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repo.GetCategories(ctx)
}

var _ = sync.Mutex{}
var _ = time.Second
