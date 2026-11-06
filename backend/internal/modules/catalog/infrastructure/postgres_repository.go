package infrastructure

import (
	"context"
	"fmt"
	"strconv"
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
		       total_installs, source_code_url, license_spdx, monetization_type, created_at, updated_at
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
	if q := strings.TrimSpace(filter.Query); q != "" {
		query += fmt.Sprintf(` AND (canonical_slug ILIKE $%d OR primary_package_name ILIKE $%d OR EXISTS (SELECT 1 FROM app_localizations l WHERE l.app_id = catalog_apps.id AND l.title ILIKE $%d))`, argIdx, argIdx, argIdx)
		args = append(args, "%"+q+"%")
		argIdx++
	}
	if filter.FreeOnly {
		query += " AND monetization_type IN ('FREE', 'FREE_WITH_ADS', 'FREE_WITH_IAP')"
	}
	if filter.DirectOnly {
		query += " AND EXISTS (SELECT 1 FROM app_distributions d WHERE d.app_id = catalog_apps.id AND d.distribution_mode = 'DIRECT_HOSTED')"
	}
	if cr, ct, ci, ok := domain.DecodeListCursor(filter.Cursor); ok {
		query += fmt.Sprintf(` AND (bayesian_rating < $%d OR (bayesian_rating = $%d AND total_installs < $%d) OR (bayesian_rating = $%d AND total_installs = $%d AND id > $%d))`, argIdx, argIdx, argIdx+1, argIdx, argIdx+1, argIdx+2)
		args = append(args, strconv.FormatFloat(cr, 'f', 3, 64), ct, ci)
		argIdx += 3
	}

	query += " ORDER BY bayesian_rating DESC, total_installs DESC, id ASC"

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit+1)

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
			&a.InstallCount, &srcCodeURL, &a.LicenseSPDX, &a.MonetizationType, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan catalog_app: %w", err)
		}
		a.DeveloperOrgID = ownerOrgID
		a.SourceCodeURL = srcCodeURL
		apps = append(apps, a)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	next := ""
	if len(apps) > limit {
		last := apps[limit-1]
		next = domain.EncodeListCursor(last.BayesianRating, last.InstallCount, last.ID)
		apps = apps[:limit]
	}
	return apps, next, nil
}

func (r *PostgresCatalogRepository) ListAppsByOrg(ctx context.Context, orgID string) ([]domain.CatalogApp, error) {
	if r.pool == nil {
		return r.fallback.ListAppsByOrg(ctx, orgID)
	}
	query := `
		SELECT id, canonical_slug, primary_package_name, owner_org_id, primary_category_id,
		       global_age_rating, default_locale, catalog_status, average_rating, bayesian_rating,
		       total_installs, source_code_url, license_spdx, monetization_type, created_at, updated_at
		FROM catalog_apps
		WHERE owner_org_id = $1
		ORDER BY updated_at DESC
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return r.fallback.ListAppsByOrg(ctx, orgID)
	}
	defer rows.Close()
	var apps []domain.CatalogApp
	for rows.Next() {
		var a domain.CatalogApp
		var ownerOrgID, srcCodeURL *string
		if err := rows.Scan(
			&a.ID, &a.CanonicalSlug, &a.PrimaryPackageName, &ownerOrgID, &a.PrimaryCategoryID,
			&a.ContentRating, &a.DefaultLocale, &a.CatalogStatus, &a.AverageRating, &a.BayesianRating,
			&a.InstallCount, &srcCodeURL, &a.LicenseSPDX, &a.MonetizationType, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan catalog_app: %w", err)
		}
		a.DeveloperOrgID = ownerOrgID
		a.SourceCodeURL = srcCodeURL
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

func (r *PostgresCatalogRepository) UpsertLocalization(ctx context.Context, loc domain.Localization) (*domain.Localization, error) {
	if r.pool == nil {
		return r.fallback.UpsertLocalization(ctx, loc)
	}
	query := `
		INSERT INTO app_localizations (id, app_id, locale, title, short_description, full_description_sanitized, whats_new, icon_s3_key)
		VALUES (COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, NULLIF($7, ''), COALESCE(NULLIF($8, ''), 'pending-upload'))
		ON CONFLICT (app_id, locale) DO UPDATE SET
			title = EXCLUDED.title,
			short_description = EXCLUDED.short_description,
			full_description_sanitized = EXCLUDED.full_description_sanitized,
			whats_new = EXCLUDED.whats_new,
			updated_at = NOW()
		RETURNING id, app_id, locale, title, short_description, full_description_sanitized, whats_new
	`
	var out domain.Localization
	var whatsNew *string
	err := r.pool.QueryRow(ctx, query, loc.ID, loc.AppID, loc.Locale, loc.Title, loc.ShortDescription, loc.FullDescription, loc.WhatsNew, "").Scan(
		&out.ID, &out.AppID, &out.Locale, &out.Title, &out.ShortDescription, &out.FullDescription, &whatsNew,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert localization: %w", err)
	}
	if whatsNew != nil {
		out.WhatsNew = *whatsNew
	}
	return &out, nil
}

func (r *PostgresCatalogRepository) GetAppBySlug(ctx context.Context, slug string) (*domain.CatalogApp, error) {
	if r.pool == nil {
		return r.fallback.GetAppBySlug(ctx, slug)
	}

	query := `
		SELECT id, canonical_slug, primary_package_name, owner_org_id, primary_category_id,
		       global_age_rating, default_locale, catalog_status, average_rating, bayesian_rating,
		       total_installs, source_code_url, license_spdx, monetization_type, created_at, updated_at
		FROM catalog_apps
		WHERE canonical_slug = $1
	`
	var a domain.CatalogApp
	var ownerOrgID, srcCodeURL *string
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&a.ID, &a.CanonicalSlug, &a.PrimaryPackageName, &ownerOrgID, &a.PrimaryCategoryID,
		&a.ContentRating, &a.DefaultLocale, &a.CatalogStatus, &a.AverageRating, &a.BayesianRating,
		&a.InstallCount, &srcCodeURL, &a.LicenseSPDX, &a.MonetizationType, &a.CreatedAt, &a.UpdatedAt,
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

func (r *PostgresCatalogRepository) GetAppByID(ctx context.Context, appID string) (*domain.CatalogApp, error) {
	if r.pool == nil {
		return r.fallback.GetAppByID(ctx, appID)
	}

	query := `
		SELECT id, canonical_slug, primary_package_name, owner_org_id, primary_category_id,
		       global_age_rating, default_locale, catalog_status, average_rating, bayesian_rating,
		       total_installs, source_code_url, license_spdx, monetization_type, created_at, updated_at
		FROM catalog_apps
		WHERE id = $1::uuid
	`
	var a domain.CatalogApp
	var ownerOrgID, srcCodeURL *string
	err := r.pool.QueryRow(ctx, query, appID).Scan(
		&a.ID, &a.CanonicalSlug, &a.PrimaryPackageName, &ownerOrgID, &a.PrimaryCategoryID,
		&a.ContentRating, &a.DefaultLocale, &a.CatalogStatus, &a.AverageRating, &a.BayesianRating,
		&a.InstallCount, &srcCodeURL, &a.LicenseSPDX, &a.MonetizationType, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrAppNotFound
		}
		return r.fallback.GetAppByID(ctx, appID)
	}
	a.DeveloperOrgID = ownerOrgID
	a.SourceCodeURL = srcCodeURL
	return &a, nil
}

func (r *PostgresCatalogRepository) UpdateCatalogStatus(ctx context.Context, appID string, status domain.CatalogStatus) error {
	if r.pool == nil {
		return r.fallback.UpdateCatalogStatus(ctx, appID, status)
	}
	tag, err := r.pool.Exec(ctx, `UPDATE catalog_apps SET catalog_status = $2, updated_at = NOW() WHERE id = $1::uuid`, appID, string(status))
	if err != nil {
		return fmt.Errorf("failed to update catalog status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAppNotFound
	}
	return nil
}

func (r *PostgresCatalogRepository) ListAppsByStatus(ctx context.Context, statuses []domain.CatalogStatus) ([]domain.CatalogApp, error) {
	if r.pool == nil {
		return r.fallback.ListAppsByStatus(ctx, statuses)
	}
	strs := make([]string, 0, len(statuses))
	for _, s := range statuses {
		strs = append(strs, string(s))
	}
	query := `
		SELECT id, canonical_slug, primary_package_name, owner_org_id, primary_category_id,
		       global_age_rating, default_locale, catalog_status, average_rating, bayesian_rating,
		       total_installs, source_code_url, license_spdx, monetization_type, created_at, updated_at
		FROM catalog_apps
		WHERE catalog_status = ANY($1)
		ORDER BY updated_at ASC
	`
	rows, err := r.pool.Query(ctx, query, strs)
	if err != nil {
		return r.fallback.ListAppsByStatus(ctx, statuses)
	}
	defer rows.Close()
	var out []domain.CatalogApp
	for rows.Next() {
		var a domain.CatalogApp
		var ownerOrgID, srcCodeURL *string
		if err := rows.Scan(
			&a.ID, &a.CanonicalSlug, &a.PrimaryPackageName, &ownerOrgID, &a.PrimaryCategoryID,
			&a.ContentRating, &a.DefaultLocale, &a.CatalogStatus, &a.AverageRating, &a.BayesianRating,
			&a.InstallCount, &srcCodeURL, &a.LicenseSPDX, &a.MonetizationType, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		a.DeveloperOrgID = ownerOrgID
		a.SourceCodeURL = srcCodeURL
		out = append(out, a)
	}
	return out, rows.Err()
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
		SELECT id, app_id, locale, device_form_factor, asset_type, s3_public_key, width_px, height_px, sha256_checksum, display_order
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
		if err := rows.Scan(&m.ID, &m.AppID, &m.Locale, &m.DeviceFormFactor, &m.AssetType, &m.URL, &m.Width, &m.Height, &m.SHA256, &m.DisplayOrder); err != nil {
			return nil, err
		}
		assets = append(assets, m)
	}
	return assets, rows.Err()
}

func (r *PostgresCatalogRepository) AddMediaAsset(ctx context.Context, asset domain.MediaAsset) (*domain.MediaAsset, error) {
	if r.pool == nil {
		return r.fallback.AddMediaAsset(ctx, asset)
	}
	query := `
		INSERT INTO app_media_assets (id, app_id, locale, device_form_factor, asset_type, s3_public_key, width_px, height_px, sha256_checksum, display_order)
		VALUES (COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()), $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	var id string
	if err := r.pool.QueryRow(ctx, query, asset.ID, asset.AppID, asset.Locale, asset.DeviceFormFactor, asset.AssetType, asset.URL, asset.Width, asset.Height, asset.SHA256, asset.DisplayOrder).Scan(&id); err != nil {
		return nil, fmt.Errorf("failed to insert media asset: %w", err)
	}
	asset.ID = id
	return &asset, nil
}

func (r *PostgresCatalogRepository) DeleteMediaAsset(ctx context.Context, appID, assetID string) error {
	if r.pool == nil {
		return r.fallback.DeleteMediaAsset(ctx, appID, assetID)
	}
	tag, err := r.pool.Exec(ctx, `DELETE FROM app_media_assets WHERE app_id = $1::uuid AND id = $2::uuid`, appID, assetID)
	if err != nil {
		return fmt.Errorf("failed to delete media asset: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrMediaAssetNotFound
	}
	return nil
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
		       COALESCE(a.min_sdk, 0), COALESCE(a.target_sdk, 0), COALESCE(a.native_abis, '{}'),
		       COALESCE(a.file_size_bytes, 0),
		       COALESCE(r.release_notes_i18n->>'ko-KR', r.release_notes_i18n->>'en-US', ''),
		       r.published_at
		FROM app_releases r
		LEFT JOIN LATERAL (
			SELECT ra.min_sdk, ra.target_sdk, ra.native_abis, ra.file_size_bytes
			FROM release_artifacts ra
			WHERE ra.release_id = r.id
			ORDER BY CASE WHEN ra.abi_filter = 'UNIVERSAL' THEN 0 ELSE 1 END, ra.file_size_bytes DESC
			LIMIT 1
		) a ON true
		WHERE r.distribution_id = $1 AND r.release_status IN ('PUBLISHED_STAGED', 'PUBLISHED_FULL')
		ORDER BY r.version_code DESC
		LIMIT 1
	`
	var rel domain.Release
	err := r.pool.QueryRow(ctx, query, distributionID).Scan(
		&rel.ID, &rel.DistributionID, &rel.AppID, &rel.ReleaseTrack, &rel.VersionCode, &rel.VersionName,
		&rel.MinSDK, &rel.TargetSDK, &rel.SupportedABIs, &rel.FileSizeBytes, &rel.ReleaseNotes, &rel.PublishedAt,
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
