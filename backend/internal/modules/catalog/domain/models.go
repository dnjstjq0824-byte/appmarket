package domain

import (
	"errors"
	"time"
)

// CatalogStatus represents the exposure state of a catalog app.
type CatalogStatus string

const (
	CatalogStatusDraft                CatalogStatus = "DRAFT"
	CatalogStatusPendingInitialReview CatalogStatus = "PENDING_INITIAL_REVIEW"
	CatalogStatusPublished            CatalogStatus = "PUBLISHED"
	CatalogStatusUnlisted             CatalogStatus = "UNLISTED"
	CatalogStatusSuspendedByAdmin     CatalogStatus = "SUSPENDED_BY_ADMIN"
	CatalogStatusArchived             CatalogStatus = "ARCHIVED"
)

// DistributionSourceType enumerates supported distribution sources.
type DistributionSourceType string

const (
	SourceBNHosted          DistributionSourceType = "BN_HOSTED"
	SourceFDroidRepo        DistributionSourceType = "FDROID_REPO"
	SourceGooglePlayLink    DistributionSourceType = "GOOGLE_PLAY_LINK"
	SourceGalaxyStoreLink   DistributionSourceType = "GALAXY_STORE_LINK"
	SourceAppleAppStoreLink DistributionSourceType = "APPLE_APPSTORE_LINK"
	SourceOfficialSiteLink  DistributionSourceType = "OFFICIAL_SITE_LINK"
)

// DistributionMode enumerates how an app can be installed.
type DistributionMode string

const (
	ModeDirectAPK        DistributionMode = "DIRECT_APK"
	ModeExternalLinkOnly DistributionMode = "EXTERNAL_LINK_ONLY"
	ModeFederatedRepoBin DistributionMode = "FEDERATED_REPO_BINARY"
	ModeSearchInfoOnly   DistributionMode = "SEARCH_INFO_ONLY"
)

// Errors
var (
	ErrAppNotFound      = errors.New("app not found")
	ErrInvalidLocale    = errors.New("invalid locale")
	ErrCacheUnavailable = errors.New("cache unavailable")
	ErrInvalidFilter    = errors.New("invalid filter")
)

// CatalogApp is the logical app entity.
type CatalogApp struct {
	ID                 string        `json:"id"`
	CanonicalSlug      string        `json:"canonical_slug"`
	PrimaryPackageName string        `json:"primary_package_name"`
	DeveloperOrgID     *string       `json:"developer_org_id,omitempty"`
	PrimaryCategoryID  string        `json:"primary_category_id"`
	ContentRating      string        `json:"content_rating"`
	DefaultLocale      string        `json:"default_locale"`
	CatalogStatus      CatalogStatus `json:"catalog_status"`
	AverageRating      float64       `json:"average_rating"`
	BayesianRating     float64       `json:"bayesian_rating"`
	InstallCount       int64         `json:"install_count"`
	SourceCodeURL      *string       `json:"source_code_url,omitempty"`
	LicenseSPDX        string        `json:"license_spdx"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// Localization holds per-locale metadata.
type Localization struct {
	ID               string `json:"id"`
	AppID            string `json:"app_id"`
	Locale           string `json:"locale"`
	Title            string `json:"title"`
	ShortDescription string `json:"short_description"`
	FullDescription  string `json:"full_description"`
	WhatsNew         string `json:"whats_new"`
	PrivacyPolicyURL string `json:"privacy_policy_url"`
}

// MediaAsset holds an image/video asset.
type MediaAsset struct {
	ID               string `json:"id"`
	AppID            string `json:"app_id"`
	Locale           string `json:"locale"`
	DeviceFormFactor string `json:"device_form_factor"`
	AssetType        string `json:"asset_type"`
	URL              string `json:"url"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
}

// Distribution is a per-source distribution entity.
type Distribution struct {
	ID                         string                 `json:"id"`
	AppID                      string                 `json:"app_id"`
	SourceType                 DistributionSourceType `json:"source_type"`
	DistributionMode           DistributionMode       `json:"distribution_mode"`
	PackageName                string                 `json:"package_name"`
	SigningCertSHA256          string                 `json:"signing_cert_sha256"`
	SignatureCompatibilityGroup string                 `json:"signature_compatibility_group"`
	OfficialStoreURL           string                 `json:"official_store_url"`
	LastSyncedAt               *time.Time             `json:"last_synced_at,omitempty"`
	LatestRelease              *Release               `json:"latest_release,omitempty"`
}
// Release is a version within a distribution.
type Release struct {
	ID             string     `json:"id"`
	DistributionID string     `json:"distribution_id"`
	AppID          string     `json:"app_id"`
	ReleaseTrack   string     `json:"release_track"`
	VersionCode    int64      `json:"version_code"`
	VersionName    string     `json:"version_name"`
	MinSDK         int        `json:"min_sdk"`
	TargetSDK      int        `json:"target_sdk"`
	SupportedABIs  []string   `json:"supported_abis"`
	FileSizeBytes  int64      `json:"file_size_bytes"`
	ReleaseNotes   string     `json:"release_notes"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
}

// Category is a hierarchical taxonomy node.
type Category struct {
	ID       string            `json:"id"`
	ParentID *string           `json:"parent_id,omitempty"`
	NameI18n map[string]string `json:"name_i18n"`
	AppCount int               `json:"app_count"`
}

// Tag is a semantic tag.
type Tag struct {
	ID       string            `json:"id"`
	NameI18n map[string]string `json:"name_i18n"`
}

// AppListFilter holds query filters for listing apps.
type AppListFilter struct {
	CategoryID string
	TagID      string
	DirectOnly bool
	MinRating  float64
	FreeOnly   bool
	Cursor     string
	Limit      int
}

// AppDetail is the aggregated detail response.
type AppDetail struct {
	App            CatalogApp     `json:"app"`
	Localization   *Localization  `json:"localization"`
	MediaAssets    []MediaAsset   `json:"media_assets"`
	Distributions  []Distribution `json:"distributions"`
	Tags           []Tag          `json:"tags"`
	ResolvedLocale string         `json:"resolved_locale"`
}

// AppSummary is a list item.
type AppSummary struct {
	ID                 string  `json:"id"`
	CanonicalSlug      string  `json:"canonical_slug"`
	Title              string  `json:"title"`
	PrimaryPackageName string  `json:"primary_package_name"`
	PrimaryCategoryID  string  `json:"primary_category_id"`
	AverageRating      float64 `json:"average_rating"`
	BayesianRating     float64 `json:"bayesian_rating"`
	InstallCount       int64   `json:"install_count"`
	IconURL            string  `json:"icon_url,omitempty"`
}

// IsSignatureCompatible reports whether two distributions share a signature group.
func (d Distribution) IsSignatureCompatible(other Distribution) bool {
	return d.SignatureCompatibilityGroup != "" && d.SignatureCompatibilityGroup == other.SignatureCompatibilityGroup
}

