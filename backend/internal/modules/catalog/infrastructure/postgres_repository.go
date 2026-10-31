package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
)

// PostgresCatalogRepository implements CatalogRepository backed by PostgreSQL 16 with in-memory fallback.
type PostgresCatalogRepository struct {
	pool     *pgxpool.Pool
	fallback *InMemoryCatalogRepository
}

func NewPostgresCatalogRepository(pool *pgxpool.Pool) *PostgresCatalogRepository {
	return &PostgresCatalogRepository{
		pool:     pool,
		fallback: NewInMemoryCatalogRepository(),
	}
}

func (r *PostgresCatalogRepository) ListPublishedApps(ctx context.Context, filter domain.AppListFilter) ([]domain.CatalogApp, string, error) {
	if r.pool == nil {
		return r.fallback.ListPublishedApps(ctx, filter)
	}

	query := `
		SELECT id, canonical_slug, primary_package_name, owner_org_id, primary_category_id,
		       global_age_rating, default_locale, catalog_status, average_rating, bayesian_rating,
		       total_installs, source_code_url, license_spdx, created_at, updated_at
		FROM catalog_apps
		WHERE catalog_status = 'PUBLISHED'
	`
	var args []any
	argIdx := 1

	if filter.CategoryID != "" {
		query += fmt.Sprintf(" AND primary_category_id = $%d", argIdx)
		args = append(args, filter.CategoryID)
		argIdx++
	}
	if filter.MinRating > 0 {
		query += fmt.Sprintf(" AND average_rating >= $%d", argIdx)
		args = append(args, filter.MinRating)
		argIdx++
	}

	query += " ORDER BY bayesian_rating DESC, total_installs DESC"

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return r.fallback.ListPublishedApps(ctx, filter)
	}
	defer rows.Close()

	var apps []domain.CatalogApp
	for rows.Next() {
		var a domain.CatalogApp
		var ownerOrgID, srcCodeURL *string
		err := rows.Scan(
			&a.ID, &a.CanonicalSlug, &a.PrimaryPackageName, &ownerOrgID, &a.PrimaryCategoryID,
			&a.ContentRating, &a.DefaultLocale, &a.CatalogStatus, &a.AverageRating, &a.BayesianRating,
			&a.InstallCount, &srcCodeURL, &a.LicenseSPDX, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan catalog_app: %w", err)
		}
		a.DeveloperOrgID = ownerOrgID
		a.SourceCodeURL = srcCodeURL
		apps = append(apps, a)
	}

	return apps, "", rows.Err()
}

func (r *PostgresCatalogRepository) GetAppBySlug(ctx context.Context, slug string) (*domain.CatalogApp, error) {
	if r.pool == nil {
		return r.fallback.GetAppBySlug(ctx, slug)
	}

	query := `
		SELECT id, canonical_slug, primary_package_name, owner_org_id, primary_category_id,
		       global_age_rating, default_locale, catalog_status, average_rating, bayesian_rating,
		       total_installs, source_code_url, license_spdx, created_at, updated_at
		FROM catalog_apps
		WHERE canonical_slug = $1
	`
	var a domain.CatalogApp
	var ownerOrgID, srcCodeURL *string
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&a.ID, &a.CanonicalSlug, &a.PrimaryPackageName, &ownerOrgID, &a.PrimaryCategoryID,
		&a.ContentRating, &a.DefaultLocale, &a.CatalogStatus, &a.AverageRating, &a.BayesianRating,
		&a.InstallCount, &srcCodeURL, &a.LicenseSPDX, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrAppNotFound
		}
		return r.fallback.GetAppBySlug(ctx, slug)
	}
	a.DeveloperOrgID = ownerOrgID
	a.SourceCodeURL = srcCodeURL
	return &a, nil
}

func (r *PostgresCatalogRepository) GetLocalizations(ctx context.Context, appID string) ([]domain.Localization, error) {
	if r.pool == nil {
		return r.fallback.GetLocalizations(ctx, appID)
	}

	query := `
		SELECT id, app_id, locale, title, short_description, full_description_sanitized, whats_new, ''
		FROM app_localizations
		WHERE app_id = $1
	`
	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return r.fallback.GetLocalizations(ctx, appID)
	}
	defer rows.Close()

	var locs []domain.Localization
	for rows.Next() {
		var l domain.Localization
		var whatsNew *string
		if err := rows.Scan(&l.ID, &l.AppID, &l.Locale, &l.Title, &l.ShortDescription, &l.FullDescription, &whatsNew, &l.PrivacyPolicyURL); err != nil {
			return nil, err
		}
		if whatsNew != nil {
			l.WhatsNew = *whatsNew
		}
		locs = append(locs, l)
	}
	return locs, rows.Err()
}

func (r *PostgresCatalogRepository) GetMediaAssets(ctx context.Context, appID, locale string) ([]domain.MediaAsset, error) {
	if r.pool == nil {
		return r.fallback.GetMediaAssets(ctx, appID, locale)
	}

	query := `
		SELECT id, app_id, locale, device_form_factor, asset_type, s3_public_key, width_px, height_px
		FROM app_media_assets
		WHERE app_id = $1
	`
	var args []any
	args = append(args, appID)
	if locale != "" {
		query += " AND locale = $2"
		args = append(args, locale)
	}
	query += " ORDER BY display_order ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return r.fallback.GetMediaAssets(ctx, appID, locale)
	}
	defer rows.Close()

	var assets []domain.MediaAsset
	for rows.Next() {
		var m domain.MediaAsset
		if err := rows.Scan(&m.ID, &m.AppID, &m.Locale, &m.DeviceFormFactor, &m.AssetType, &m.URL, &m.Width, &m.Height); err != nil {
			return nil, err
		}
		assets = append(assets, m)
	}
	return assets, rows.Err()
}

func (r *PostgresCatalogRepository) GetDistributions(ctx context.Context, appID string) ([]domain.Distribution, error) {
	if r.pool == nil {
		return r.fallback.GetDistributions(ctx, appID)
	}

	query := `
		SELECT id, app_id, source_type, distribution_mode, package_name, COALESCE(signing_cert_sha256, '')
		FROM app_distributions
		WHERE app_id = $1
	`
	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return r.fallback.GetDistributions(ctx, appID)
	}
	defer rows.Close()

	var dists []domain.Distribution
	for rows.Next() {
		var d domain.Distribution
		if err := rows.Scan(&d.ID, &d.AppID, &d.SourceType, &d.DistributionMode, &d.PackageName, &d.SigningCertSHA256); err != nil {
			return nil, err
		}
		dists = append(dists, d)
	}
	return dists, rows.Err()
}

func (r *PostgresCatalogRepository) GetLatestRelease(ctx context.Context, distributionID string) (*domain.Release, error) {
	if r.pool == nil {
		return r.fallback.GetLatestRelease(ctx, distributionID)
	}

	query := `
		SELECT r.id, r.distribution_id, r.app_id, r.release_track, r.version_code, r.version_name,
		       r.min_sdk, r.target_sdk, COALESCE(a.file_size_bytes, 0), COALESCE(r.release_notes, ''), r.published_at
		FROM app_releases r
		LEFT JOIN release_artifacts a ON a.release_id = r.id
		WHERE r.distribution_id = $1 AND r.release_status = 'PUBLISHED'
		ORDER BY r.version_code DESC
		LIMIT 1
	`
	var rel domain.Release
	err := r.pool.QueryRow(ctx, query, distributionID).Scan(
		&rel.ID, &rel.DistributionID, &rel.AppID, &rel.ReleaseTrack, &rel.VersionCode, &rel.VersionName,
		&rel.MinSDK, &rel.TargetSDK, &rel.FileSizeBytes, &rel.ReleaseNotes, &rel.PublishedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return r.fallback.GetLatestRelease(ctx, distributionID)
	}
	return &rel, nil
}

func (r *PostgresCatalogRepository) GetCategories(ctx context.Context) ([]domain.Category, error) {
	if r.pool == nil {
		return r.fallback.GetCategories(ctx)
	}

	query := `
		SELECT id, parent_id, name_i18n, (SELECT COUNT(*) FROM catalog_apps WHERE primary_category_id = c.id AND catalog_status = 'PUBLISHED')
		FROM categories c
		ORDER BY display_order ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return r.fallback.GetCategories(ctx)
	}
	defer rows.Close()

	var cats []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.ParentID, &c.NameI18n, &c.AppCount); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *PostgresCatalogRepository) GetTags(ctx context.Context, appID string) ([]domain.Tag, error) {
	if r.pool == nil {
		return r.fallback.GetTags(ctx, appID)
	}

	query := `
		SELECT t.id, t.name_i18n
		FROM tags t
		JOIN app_tags at ON at.tag_id = t.id
		WHERE at.app_id = $1
	`
	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return nil, nil
		}
		return r.fallback.GetTags(ctx, appID)
	}
	defer rows.Close()

	var tags []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.NameI18n); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}
