package console

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	distApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	distDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
	distInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/infrastructure"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/security/scanner"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/sanitize"
)

const maxUploadBytes = 2 * 1024 * 1024 * 1024

var (
	ErrUploadInvalid    = errors.New("ERR_UPLOAD_INVALID")
	ErrUploadNotFound   = errors.New("ERR_UPLOAD_NOT_FOUND")
	ErrUploadState      = errors.New("ERR_UPLOAD_STATE")
	ErrUploadExpired    = errors.New("ERR_UPLOAD_EXPIRED")
	ErrChecksumMismatch = errors.New("ERR_UPLOAD_CHECKSUM_MISMATCH")
	ErrQuarantineStore  = errors.New("ERR_QUARANTINE_STORE_FAILED")
)

// UploadService drives quarantine upload sessions plus APK/AAB completion.
type UploadService struct {
	artifacts        *distApp.ArtifactService
	quarantineStore  distApp.ObjectStorage
	quarantineBucket string
}

func NewUploadService(artifacts *distApp.ArtifactService) *UploadService {
	return &UploadService{artifacts: artifacts}
}

// WithQuarantineStorage enables persistence of uploaded binaries at the
// session's S3QuarantineKey before scanning. Without it completion keeps the
// previous scan-only behavior (unit-test mode).
func (s *UploadService) WithQuarantineStorage(store distApp.ObjectStorage, bucket string) *UploadService {
	s.quarantineStore = store
	s.quarantineBucket = bucket
	return s
}

// InitUpload validates the declared binary and opens a quarantine session.
func (s *UploadService) InitUpload(ctx context.Context, orgID, appID, fileName string, fileSize int64, expectedSHA256 string) (*distDomain.UploadSession, error) {
	name := sanitize.Strict(strings.TrimSpace(fileName))
	if name == "" {
		return nil, fmt.Errorf("%w: file name required", ErrUploadInvalid)
	}
	lower := strings.ToLower(name)
	artifactType := distDomain.ArtifactTypeAPK
	switch {
	case strings.HasSuffix(lower, ".apk"):
		artifactType = distDomain.ArtifactTypeAPK
	case strings.HasSuffix(lower, ".aab"):
		artifactType = distDomain.ArtifactTypeAAB
	default:
		return nil, fmt.Errorf("%w: only .apk and .aab are accepted", ErrUploadInvalid)
	}
	if fileSize <= 0 || fileSize > maxUploadBytes {
		return nil, fmt.Errorf("%w: file size out of range 1..2GiB", ErrUploadInvalid)
	}
	sha := strings.ToLower(strings.TrimSpace(expectedSHA256))
	if len(sha) != 64 {
		return nil, fmt.Errorf("%w: expectedSha256 must be 64 hex chars", ErrUploadInvalid)
	}
	if _, err := hex.DecodeString(sha); err != nil {
		return nil, fmt.Errorf("%w: expectedSha256 is not hex", ErrUploadInvalid)
	}
	return s.artifacts.InitiateUploadSession(ctx, orgID, appID, artifactType, name, fileSize, sha)
}

// BinaryDetails carries parsed manifest facts back to the console.
type BinaryDetails struct {
	PackageName          string   `json:"packageName"`
	VersionCode          int64    `json:"versionCode"`
	VersionName          string   `json:"versionName"`
	MinSDK               int      `json:"minSdk"`
	TargetSDK            int      `json:"targetSdk"`
	SHA256               string   `json:"sha256"`
	CertSHA256           string   `json:"certSha256"`
	DangerousPermissions []string `json:"dangerousPermissions"`
}

// CompleteResult is the completion receipt. Details is nil when the caller
// completed metadata-only (no binary bytes supplied).
type CompleteResult struct {
	UploadID string                 `json:"uploadId"`
	Status   string                 `json:"status"`
	Release  *distDomain.AppRelease `json:"release,omitempty"`
	Details  *BinaryDetails         `json:"details,omitempty"`
}

// CompleteUpload finishes a session. When apkBytes are supplied the real
// quarantine pipeline runs (AXML parse, v2/v3 signature verify, lineage,
// artifact registration); otherwise the session completes metadata-only and
// the binary is expected through the quarantine object-store key.
func (s *UploadService) CompleteUpload(ctx context.Context, uploadID, sha256Hex string, apkBytes []byte) (*CompleteResult, error) {
	session, err := s.artifacts.GetUploadSession(ctx, uploadID)
	if err != nil {
		return nil, ErrUploadNotFound
	}
	if session.Status != "INITIATED" {
		return nil, fmt.Errorf("%w: session is %s", ErrUploadState, session.Status)
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, ErrUploadExpired
	}
	if !strings.EqualFold(strings.TrimSpace(sha256Hex), session.ExpectedSHA256) {
		return nil, ErrChecksumMismatch
	}
	if len(apkBytes) == 0 {
		if err := s.artifacts.MarkUploadComplete(ctx, uploadID); err != nil {
			return nil, err
		}
		return &CompleteResult{UploadID: uploadID, Status: "COMPLETED"}, nil
	}
	sum := sha256.Sum256(apkBytes)
	if !strings.EqualFold(hex.EncodeToString(sum[:]), session.ExpectedSHA256) {
		return nil, ErrChecksumMismatch
	}
	// Quarantine-first: land the verified bytes in isolated storage before any
	// parsing, so even rejected binaries stay available for forensics and the
	// promotion path always finds the object behind S3QuarantineKey.
	if s.quarantineStore != nil && s.quarantineBucket != "" {
		if err := s.quarantineStore.PutObject(ctx, s.quarantineBucket, &distApp.StorageObject{
			Key:         session.S3QuarantineKey,
			Content:     apkBytes,
			SHA256:      session.ExpectedSHA256,
			ContentType: quarantineContentType(session.ArtifactType),
			Tags: map[string]string{
				"status":     "quarantined",
				"session_id": session.ID,
				"app_id":     session.AppID,
			},
		}); err != nil {
			return nil, ErrQuarantineStore
		}
	}
	now := time.Now().UTC()
	release := &distDomain.AppRelease{
		ID:                uuid.New().String(),
		DistributionID:    "dist-" + session.AppID,
		AppID:             session.AppID,
		ReleaseTrack:      distDomain.TrackProduction,
		VersionCode:       0,
		VersionName:       "pending-parse",
		ReleaseStatus:     distDomain.ReleaseDraft,
		RolloutPercentage: 0,
		CreatedBy:         session.OrgID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.artifacts.CreateRelease(ctx, release); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(apkBytes)
	manifest, err := distInfra.InspectAPK(reader, int64(len(apkBytes)), session.ExpectedSHA256)
	if err != nil {
		return nil, err
	}
	artifact, signature, err := s.artifacts.ProcessUploadedAPK(ctx, uploadID, release.ID, reader, int64(len(apkBytes)))
	if err != nil {
		return nil, err
	}
	permResult := scanner.CalculatePermissionRiskScore(manifest.Permissions)
	details := &BinaryDetails{
		PackageName:          manifest.PackageName,
		VersionCode:          manifest.VersionCode,
		VersionName:          manifest.VersionName,
		MinSDK:               manifest.MinSDKVersion,
		TargetSDK:            manifest.TargetSDKVersion,
		SHA256:               manifest.SHA256Hash,
		CertSHA256:           signature.CertSHA256Fingerprint,
		DangerousPermissions: permResult.DangerousPermissions,
	}
	if details.DangerousPermissions == nil {
		details.DangerousPermissions = []string{}
	}
	updated, err := s.artifacts.GetRelease(ctx, release.ID)
	if err != nil {
		return nil, err
	}
	_ = artifact
	return &CompleteResult{UploadID: uploadID, Status: "COMPLETED", Release: updated, Details: details}, nil
}

func quarantineContentType(artifactType distDomain.ArtifactType) string {
	if artifactType == distDomain.ArtifactTypeAAB {
		return "application/octet-stream"
	}
	return "application/vnd.android.package-archive"
}
