package domain

import (
	"errors"
	"time"
)

type ArtifactType string

const (
	ArtifactTypeAPK ArtifactType = "APK"
	ArtifactTypeAAB ArtifactType = "AAB"
)

type ArtifactStatus string

const (
	ArtifactStatusQuarantined              ArtifactStatus = "QUARANTINED"
	ArtifactStatusVerifiedSafe             ArtifactStatus = "VERIFIED_SAFE"
	ArtifactStatusInfected                 ArtifactStatus = "INFECTED"
	ArtifactStatusRejectedInvalid          ArtifactStatus = "REJECTED_INVALID_APK"
	ArtifactStatusRejectedSignatureMismatch ArtifactStatus = "REJECTED_SIGNATURE_MISMATCH"
	ArtifactStatusPromotedPublic           ArtifactStatus = "PROMOTED_PUBLIC"
)

type ReleaseTrack string

const (
	TrackInternal   ReleaseTrack = "INTERNAL"
	TrackBeta       ReleaseTrack = "BETA"
	TrackProduction ReleaseTrack = "PRODUCTION"
)

type ReleaseStatus string

const (
	ReleaseDraft                 ReleaseStatus = "DRAFT"
	ReleaseArtifactVerifying     ReleaseStatus = "ARTIFACT_VERIFYING"
	ReleaseSecurityScanning      ReleaseStatus = "SECURITY_SCANNING"
	ReleaseScanFailedQuarantined ReleaseStatus = "SCAN_FAILED_QUARANTINED"
	ReleaseReadyForSubmission    ReleaseStatus = "READY_FOR_SUBMISSION"
	ReleaseSubmittedForReview    ReleaseStatus = "SUBMITTED_FOR_REVIEW"
	ReleaseInReview              ReleaseStatus = "IN_REVIEW"
	ReleaseChangesRequested      ReleaseStatus = "CHANGES_REQUESTED"
	ReleaseRejected              ReleaseStatus = "REJECTED"
	ReleaseApprovedPendingPub    ReleaseStatus = "APPROVED_PENDING_PUBLISH"
	ReleasePublishedStaged       ReleaseStatus = "PUBLISHED_STAGED"
	ReleasePublishedFull         ReleaseStatus = "PUBLISHED_FULL"
	ReleaseHalted                ReleaseStatus = "HALTED"
	ReleaseSuperseded            ReleaseStatus = "SUPERSEDED"
	ReleaseRevokedByAdmin        ReleaseStatus = "REVOKED_BY_ADMIN"
)

var (
	ErrInvalidStateTransition    = errors.New("ERR_INVALID_RELEASE_STATE_TRANSITION")
	ErrApkV2SignatureRequired    = errors.New("ERR_APK_V2_SIGNATURE_REQUIRED")
	ErrSignatureMismatch         = errors.New("REJECTED_SIGNATURE_MISMATCH")
	ErrZipBombDetected           = errors.New("ERR_ZIP_BOMB_DETECTED")
	ErrZipSlipDetected           = errors.New("ERR_ZIP_SLIP_DETECTED")
	ErrInvalidApkFile            = errors.New("REJECTED_INVALID_APK")
	ErrReleaseNotFound           = errors.New("release not found")
	ErrArtifactNotFound          = errors.New("artifact not found")
	ErrRolloutPercentageInvalid  = errors.New("rollout percentage must be between 0 and 100")
	ErrBlacklistedCertificate    = errors.New("ERR_BLACKLISTED_SIGNING_CERTIFICATE")
)

type UploadSession struct {
	ID                string         `json:"id"`
	OrgID             string         `json:"org_id"`
	AppID             string         `json:"app_id"`
	ArtifactType      ArtifactType   `json:"artifact_type"`
	FileName          string         `json:"file_name"`
	ExpectedSizeBytes int64          `json:"expected_size_bytes"`
	ExpectedSHA256    string         `json:"expected_sha256"`
	S3QuarantineKey   string         `json:"s3_quarantine_key"`
	Status            string         `json:"status"` // INITIATED, COMPLETED, ABORTED, EXPIRED
	ExpiresAt         time.Time      `json:"expires_at"`
	CreatedAt         time.Time      `json:"created_at"`
}

type ReleaseArtifact struct {
	ID                     string         `json:"id"`
	ReleaseID              string         `json:"release_id"`
	ArtifactType           ArtifactType   `json:"artifact_type"`
	FileName               string         `json:"file_name"`
	SizeBytes              int64          `json:"size_bytes"`
	SHA256Hash             string         `json:"sha256_hash"`
	S3QuarantineKey        string         `json:"s3_quarantine_key"`
	S3PublicCDNKey         *string        `json:"s3_public_cdn_key,omitempty"`
	PackageName            string         `json:"package_name"`
	VersionCode            int64          `json:"version_code"`
	MinSDKVersion          int            `json:"min_sdk_version"`
	TargetSDKVersion       int            `json:"target_sdk_version"`
	NativeABIs             []string       `json:"native_abis"`
	IsUniversalAPK         bool           `json:"is_universal_apk"`
	SignatureSchemeVersion int            `json:"signature_scheme_version"`
	Status                 ArtifactStatus `json:"status"`
	CreatedAt              time.Time      `json:"created_at"`
}

type APKSignature struct {
	ID                     string     `json:"id"`
	ArtifactID             string     `json:"artifact_id"`
	SignatureSchemeVersion int        `json:"signature_scheme_version"`
	CertSHA256Fingerprint  string     `json:"cert_sha256_fingerprint"`
	CertSHA1Fingerprint    string     `json:"cert_sha1_fingerprint"`
	PublicKeyAlgorithm     string     `json:"public_key_algorithm"`
	KeySizeBits            int        `json:"key_size_bits"`
	SubjectDN              string     `json:"subject_dn"`
	IssuerDN               string     `json:"issuer_dn"`
	SerialNumberHex        string     `json:"serial_number_hex"`
	ValidFrom              time.Time  `json:"valid_from"`
	ValidUntil             time.Time  `json:"valid_until"`
	ProofOfRotation        string     `json:"proof_of_rotation,omitempty"`
	IsPrimaryLineage       bool       `json:"is_primary_lineage"`
	CreatedAt              time.Time  `json:"created_at"`
}

type AppRelease struct {
	ID                string            `json:"id"`
	DistributionID    string            `json:"distribution_id"`
	AppID             string            `json:"app_id"`
	ReleaseTrack      ReleaseTrack      `json:"release_track"`
	VersionCode       int64             `json:"version_code"`
	VersionName       string            `json:"version_name"`
	ReleaseNotesI18n  map[string]string `json:"release_notes_i18n"`
	ReleaseStatus     ReleaseStatus     `json:"release_status"`
	RolloutPercentage int               `json:"rollout_percentage"`
	PreviousReleaseID *string           `json:"previous_release_id,omitempty"`
	CreatedBy         string            `json:"created_by"`
	PublishedAt       *time.Time        `json:"published_at,omitempty"`
	HaltedAt          *time.Time        `json:"halted_at,omitempty"`
	HaltReason        *string           `json:"halt_reason,omitempty"`
	Version           int               `json:"version"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

