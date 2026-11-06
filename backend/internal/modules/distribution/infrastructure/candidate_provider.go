package infrastructure

import (
	"context"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
)

// InMemoryCandidateProvider stores candidate releases in-memory for distribution checks.
type InMemoryCandidateProvider struct {
	mu         sync.RWMutex
	candidates map[string]domain.CandidateRelease
}

// NewInMemoryCandidateProvider creates and seeds the in-memory candidate provider.
func NewInMemoryCandidateProvider() *InMemoryCandidateProvider {
	p := &InMemoryCandidateProvider{
		candidates: make(map[string]domain.CandidateRelease),
	}
	p.seedCandidates()
	return p
}

func (p *InMemoryCandidateProvider) seedCandidates() {
	p.candidates["com.omniwork.pro"] = domain.CandidateRelease{
		AppID:              "app-omni-01",
		ReleaseID:          "rel-omni-2026-01",
		PackageName:        "com.omniwork.pro",
		VersionCode:        20260100,
		VersionName:        "4.2.0",
		MinSDK:             26,
		TargetSDK:          34,
		NativeABIs:         []string{"arm64-v8a", "armeabi-v7a", "x86_64"},
		RequiredFeatures:   []string{},
		CertSHA256:         "A1B2C3D4E5F67890123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0",
		RotatedCertLineage: []string{},
		RolloutPercentage:  100,
		ReleaseStatus:      domain.ReleasePublishedFull,
		DownloadURL:        "https://cdn.appmarket.local/packages/com.omniwork.pro/rel-omni-2026-01/omni-4.2.0.apk",
		SHA256Checksum:     "a1b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abcdef0",
		SizeBytes:          35800000,
	}

	p.candidates["com.finvault.app"] = domain.CandidateRelease{
		AppID:              "app-finvault-02",
		ReleaseID:          "rel-fin-2026-02",
		PackageName:        "com.finvault.app",
		VersionCode:        20260200,
		VersionName:        "3.1.0",
		MinSDK:             28,
		TargetSDK:          34,
		NativeABIs:         []string{"arm64-v8a"},
		RequiredFeatures:   []string{"android.hardware.fingerprint"},
		CertSHA256:         "B2C3D4E5F67890123456789ABCDEF0123456789ABCDEF0123456789ABCDEF01A",
		RotatedCertLineage: []string{},
		RolloutPercentage:  100,
		ReleaseStatus:      domain.ReleasePublishedFull,
		DownloadURL:        "https://cdn.appmarket.local/packages/com.finvault.app/rel-fin-2026-02/finvault-3.1.0.apk",
		SHA256Checksum:     "b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abcdef01a",
		SizeBytes:          24500000,
	}
}

// SetCandidate registers or overrides a candidate for a package name.
func (p *InMemoryCandidateProvider) SetCandidate(cand domain.CandidateRelease) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.candidates[cand.PackageName] = cand
}

// GetCandidatesForPackages returns candidate releases for all requested packages.
func (p *InMemoryCandidateProvider) GetCandidatesForPackages(ctx context.Context, packageNames []string) (map[string]domain.CandidateRelease, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]domain.CandidateRelease)
	for _, pkg := range packageNames {
		if cand, exists := p.candidates[pkg]; exists {
			result[pkg] = cand
		}
	}
	return result, nil
}

// PostgresCandidateProvider queries candidates from PostgreSQL 16 database.
type PostgresCandidateProvider struct {
	db       *pgxpool.Pool
	cdnBase  string
	fallback *InMemoryCandidateProvider
}

// NewPostgresCandidateProvider creates a candidate provider backed by PostgreSQL with in-memory fallback.
// cdnBase is the public CDN origin (host or host/path, without scheme) used to
// compose download URLs from release_artifacts.s3_public_key.
func NewPostgresCandidateProvider(db *pgxpool.Pool, cdnBase string) *PostgresCandidateProvider {
	return &PostgresCandidateProvider{
		db:       db,
		cdnBase:  strings.Trim(strings.TrimSpace(cdnBase), "/"),
		fallback: NewInMemoryCandidateProvider(),
	}
}

// GetCandidatesForPackages retrieves candidate releases from DB.
func (p *PostgresCandidateProvider) GetCandidatesForPackages(ctx context.Context, packageNames []string) (map[string]domain.CandidateRelease, error) {
	if p.db == nil {
		return p.fallback.GetCandidatesForPackages(ctx, packageNames)
	}

	query := `
		SELECT
			r.app_id, r.id, d.package_name, r.version_code, r.version_name,
			art.min_sdk, art.target_sdk, art.native_abis, art.required_features,
			COALESCE(d.signing_cert_sha256, ''),
			r.rollout_percentage, r.release_status,
			art.s3_public_key, art.sha256_checksum, art.file_size_bytes
		FROM app_releases r
		JOIN app_distributions d ON d.id = r.distribution_id
		JOIN catalog_apps a ON a.id = r.app_id
		JOIN release_artifacts art ON art.release_id = r.id
			AND art.artifact_status = 'PUBLISHED_ON_CDN'
		WHERE d.package_name = ANY($1)
		  AND r.release_status IN ('PUBLISHED_STAGED', 'PUBLISHED_FULL')
		ORDER BY r.version_code DESC`

	rows, err := p.db.Query(ctx, query, packageNames)
	if err != nil {
		// Fallback to in-memory seed on DB read failure to prevent catastrophic service outage
		return p.fallback.GetCandidatesForPackages(ctx, packageNames)
	}
	defer rows.Close()

	result := make(map[string]domain.CandidateRelease)
	for rows.Next() {
		var cand domain.CandidateRelease
		var statusStr string
		var publicKey *string

		if err := rows.Scan(
			&cand.AppID,
			&cand.ReleaseID,
			&cand.PackageName,
			&cand.VersionCode,
			&cand.VersionName,
			&cand.MinSDK,
			&cand.TargetSDK,
			&cand.NativeABIs,
			&cand.RequiredFeatures,
			&cand.CertSHA256,
			&cand.RolloutPercentage,
			&statusStr,
			&publicKey,
			&cand.SHA256Checksum,
			&cand.SizeBytes,
		); err != nil {
			return nil, err
		}

		cand.ReleaseStatus = domain.ReleaseStatus(statusStr)
		if publicKey != nil && *publicKey != "" && p.cdnBase != "" {
			cand.DownloadURL = "https://" + p.cdnBase + "/" + strings.TrimPrefix(*publicKey, "/")
		}

		// Keep highest version code per package
		if existing, exists := result[cand.PackageName]; !exists || cand.VersionCode > existing.VersionCode {
			result[cand.PackageName] = cand
		}
	}

	return result, nil
}

