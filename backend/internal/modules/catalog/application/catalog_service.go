package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/platform/sanitize"
	"golang.org/x/sync/singleflight"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
)

// CatalogService implements catalog query use cases.
type CatalogService struct {
	repo     CatalogRepository
	cache    Cache
	sfGroup  singleflight.Group
	cacheTTL int
}

// NewCatalogService constructs a CatalogService.
func NewCatalogService(repo CatalogRepository, cache Cache) *CatalogService {
	return &CatalogService{
		repo:     repo,
		cache:    cache,
		cacheTTL: 300,
	}
}

// SanitizeHTML strips untrusted markup via the platform sanitize policy.
func (s *CatalogService) SanitizeHTML(input string) string {
	return sanitize.Strict(input)
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

// LatestRelease returns the newest published release across an app's
// distributions, or ErrAppNotFound when none is published yet.
func (s *CatalogService) LatestRelease(ctx context.Context, appID string) (*domain.Release, error) {
	dists, err := s.repo.GetDistributions(ctx, appID)
	if err != nil {
		return nil, err
	}
	var latest *domain.Release
	for _, d := range dists {
		rel, err := s.repo.GetLatestRelease(ctx, d.ID)
		if err != nil || rel == nil {
			continue
		}
		if latest == nil || rel.VersionCode > latest.VersionCode {
			cp := *rel
			latest = &cp
		}
	}
	if latest == nil {
		return nil, domain.ErrAppNotFound
	}
	return latest, nil
}

// ListLocalizations returns every locale stored for an app.
func (s *CatalogService) ListLocalizations(ctx context.Context, appID string) ([]domain.Localization, error) {
	return s.repo.GetLocalizations(ctx, appID)
}

// ListAppsByOrg returns every app owned by a developer org, any status.
func (s *CatalogService) ListAppsByOrg(ctx context.Context, orgID string) ([]domain.CatalogApp, error) {
	return s.repo.ListAppsByOrg(ctx, orgID)
}

// SaveLocalization validates, sanitizes, and persists one locale's metadata.
func (s *CatalogService) SaveLocalization(ctx context.Context, loc domain.Localization) (*domain.Localization, error) {
	if loc.AppID == "" || loc.Locale == "" || strings.TrimSpace(loc.Title) == "" {
		return nil, fmt.Errorf("app_id, locale and title are required")
	}
	loc.Title = s.SanitizeHTML(loc.Title)
	loc.ShortDescription = s.SanitizeHTML(loc.ShortDescription)
	loc.FullDescription = s.SanitizeHTML(loc.FullDescription)
	loc.WhatsNew = s.SanitizeHTML(loc.WhatsNew)
	return s.repo.UpsertLocalization(ctx, loc)
}

// SaveMediaAsset validates and persists a store-listing media asset.
func (s *CatalogService) SaveMediaAsset(ctx context.Context, asset domain.MediaAsset) (*domain.MediaAsset, error) {
	if err := validateMediaAsset(asset); err != nil {
		return nil, err
	}
	asset.URL = sanitize.Strict(strings.TrimSpace(asset.URL))
	return s.repo.AddMediaAsset(ctx, asset)
}

// ListMediaAssets returns media assets for an app, optionally filtered by locale.
func (s *CatalogService) ListMediaAssets(ctx context.Context, appID, locale string) ([]domain.MediaAsset, error) {
	assets, err := s.repo.GetMediaAssets(ctx, appID, locale)
	if err != nil {
		return nil, err
	}
	if assets == nil {
		return []domain.MediaAsset{}, nil
	}
	return assets, nil
}

// DeleteMediaAsset removes a media asset.
func (s *CatalogService) DeleteMediaAsset(ctx context.Context, appID, assetID string) error {
	if strings.TrimSpace(appID) == "" || strings.TrimSpace(assetID) == "" {
		return fmt.Errorf("app_id and asset_id are required")
	}
	return s.repo.DeleteMediaAsset(ctx, appID, assetID)
}

// ApplyReviewDecision moves an app to PUBLISHED (approved) or back to DRAFT
// (rejected) after verifying the app exists.
func (s *CatalogService) ApplyReviewDecision(ctx context.Context, appID string, approved bool) error {
	if _, err := s.repo.GetAppByID(ctx, appID); err != nil {
		return err
	}
	status := domain.CatalogStatusDraft
	if approved {
		status = domain.CatalogStatusPublished
	}
	return s.repo.UpdateCatalogStatus(ctx, appID, status)
}

// SubmitForReview moves a DRAFT app into the review queue.
func (s *CatalogService) SubmitForReview(ctx context.Context, appID string) error {
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return err
	}
	if app.CatalogStatus != domain.CatalogStatusDraft && app.CatalogStatus != domain.CatalogStatusUnlisted {
		return fmt.Errorf("only DRAFT or UNLISTED apps can be submitted for review")
	}
	return s.repo.UpdateCatalogStatus(ctx, appID, domain.CatalogStatusPendingInitialReview)
}

// GetAppByID returns the catalog app with the given ID.
func (s *CatalogService) GetAppByID(ctx context.Context, appID string) (*domain.CatalogApp, error) {
	return s.repo.GetAppByID(ctx, appID)
}

// ListReviewQueue returns apps awaiting initial review.
func (s *CatalogService) ListReviewQueue(ctx context.Context) ([]domain.CatalogApp, error) {
	apps, err := s.repo.ListAppsByStatus(ctx, []domain.CatalogStatus{domain.CatalogStatusPendingInitialReview})
	if err != nil {
		return nil, err
	}
	if apps == nil {
		return []domain.CatalogApp{}, nil
	}
	return apps, nil
}

func validateMediaAsset(asset domain.MediaAsset) error {
	if strings.TrimSpace(asset.AppID) == "" || strings.TrimSpace(asset.Locale) == "" {
		return fmt.Errorf("app_id and locale are required")
	}
	switch asset.AssetType {
	case "SCREENSHOT", "FEATURE_BANNER":
	default:
		return fmt.Errorf("invalid asset_type %q: must be SCREENSHOT or FEATURE_BANNER", asset.AssetType)
	}
	switch asset.DeviceFormFactor {
	case "PHONE", "TABLET_7", "TABLET_10", "TV", "FOLDABLE":
	default:
		return fmt.Errorf("invalid device_form_factor %q", asset.DeviceFormFactor)
	}
	if strings.TrimSpace(asset.URL) == "" {
		return fmt.Errorf("url is required")
	}
	if asset.Width <= 0 || asset.Height <= 0 {
		return fmt.Errorf("width and height must be positive")
	}
	if len(asset.SHA256) != 64 {
		return fmt.Errorf("sha256 must be 64 hex chars")
	}
	for _, c := range strings.ToLower(asset.SHA256) {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return fmt.Errorf("sha256 is not hex")
		}
	}
	return nil
}

// GetChart returns a ranked app list for a chart type: top_free, top_paid,
// trending, or top_grossing.
func (s *CatalogService) GetChart(ctx context.Context, chartType string) ([]domain.AppSummary, error) {
	filter := domain.AppListFilter{Limit: 20}
	switch strings.ToLower(chartType) {
	case "top_free":
		filter.FreeOnly = true
	case "top_paid", "top_grossing":
		// Paid ranking shares the published list until paid inventory exists;
		// ordering stays install-driven so ranks are always real positions.
	case "trending":
	default:
		return nil, domain.ErrInvalidFilter
	}
	summaries, _, err := s.ListApps(ctx, filter)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(chartType, "trending") {
		sort.Slice(summaries, func(i, j int) bool {
			if summaries[i].BayesianRating != summaries[j].BayesianRating {
				return summaries[i].BayesianRating > summaries[j].BayesianRating
			}
			return summaries[i].InstallCount > summaries[j].InstallCount
		})
	}
	return summaries, nil
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
		if !errors.Is(err, domain.ErrAppNotFound) {
			return nil, err
		}
		// Deep links may carry the stable app ID instead of the slug.
		app, err = s.repo.GetAppByID(ctx, slug)
		if err != nil {
			return nil, err
		}
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
