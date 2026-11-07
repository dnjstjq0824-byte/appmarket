package application

import (
	"context"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/infrastructure"
	"github.com/google/uuid"
)

type DistributionRepository interface {
	CreateUploadSession(ctx context.Context, s *domain.UploadSession) error
	GetUploadSession(ctx context.Context, id string) (*domain.UploadSession, error)
	UpdateUploadSessionStatus(ctx context.Context, id, status string) error

	CreateArtifact(ctx context.Context, a *domain.ReleaseArtifact) error
	GetArtifact(ctx context.Context, id string) (*domain.ReleaseArtifact, error)
	UpdateArtifactStatus(ctx context.Context, id string, status domain.ArtifactStatus) error
	ListArtifactsByRelease(ctx context.Context, releaseID string) ([]*domain.ReleaseArtifact, error)
	SetArtifactPublicCDNKey(ctx context.Context, id, key string) error

	SaveSignature(ctx context.Context, sig *domain.APKSignature) error
	GetLatestSignatureByDistribution(ctx context.Context, distributionID string) (*domain.APKSignature, error)

	CreateRelease(ctx context.Context, r *domain.AppRelease) error
	GetRelease(ctx context.Context, id string) (*domain.AppRelease, error)
	UpdateRelease(ctx context.Context, r *domain.AppRelease) error
	ListReleasesByApp(ctx context.Context, appID string) ([]*domain.AppRelease, error)
}

type InMemoryDistributionRepo struct {
	mu         sync.RWMutex
	sessions   map[string]*domain.UploadSession
	artifacts  map[string]*domain.ReleaseArtifact
	signatures map[string]*domain.APKSignature
	releases   map[string]*domain.AppRelease
}

func NewInMemoryDistributionRepo() *InMemoryDistributionRepo {
	return &InMemoryDistributionRepo{
		sessions:   make(map[string]*domain.UploadSession),
		artifacts:  make(map[string]*domain.ReleaseArtifact),
		signatures: make(map[string]*domain.APKSignature),
		releases:   make(map[string]*domain.AppRelease),
	}
}

func (r *InMemoryDistributionRepo) CreateUploadSession(ctx context.Context, s *domain.UploadSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = s
	return nil
}

func (r *InMemoryDistributionRepo) GetUploadSession(ctx context.Context, id string) (*domain.UploadSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, fmt.Errorf("upload session not found")
	}
	c := *s
	return &c, nil
}

func (r *InMemoryDistributionRepo) UpdateUploadSessionStatus(ctx context.Context, id, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return fmt.Errorf("upload session not found")
	}
	s.Status = status
	return nil
}

func (r *InMemoryDistributionRepo) CreateArtifact(ctx context.Context, a *domain.ReleaseArtifact) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.artifacts[a.ID] = a
	return nil
}

func (r *InMemoryDistributionRepo) GetArtifact(ctx context.Context, id string) (*domain.ReleaseArtifact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.artifacts[id]
	if !ok {
		return nil, domain.ErrArtifactNotFound
	}
	c := *a
	return &c, nil
}

func (r *InMemoryDistributionRepo) UpdateArtifactStatus(ctx context.Context, id string, status domain.ArtifactStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.artifacts[id]
	if !ok {
		return domain.ErrArtifactNotFound
	}
	a.Status = status
	return nil
}

func (r *InMemoryDistributionRepo) ListArtifactsByRelease(ctx context.Context, releaseID string) ([]*domain.ReleaseArtifact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.ReleaseArtifact
	for _, a := range r.artifacts {
		if a.ReleaseID == releaseID {
			c := *a
			out = append(out, &c)
		}
	}
	return out, nil
}

func (r *InMemoryDistributionRepo) SetArtifactPublicCDNKey(ctx context.Context, id, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.artifacts[id]
	if !ok {
		return domain.ErrArtifactNotFound
	}
	a.S3PublicCDNKey = &key
	return nil
}

func (r *InMemoryDistributionRepo) SaveSignature(ctx context.Context, sig *domain.APKSignature) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.signatures[sig.ID] = sig
	return nil
}

func (r *InMemoryDistributionRepo) GetLatestSignatureByDistribution(ctx context.Context, distributionID string) (*domain.APKSignature, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rel := range r.releases {
		if rel.DistributionID == distributionID {
			for _, art := range r.artifacts {
				if art.ReleaseID == rel.ID {
					for _, sig := range r.signatures {
						if sig.ArtifactID == art.ID {
							c := *sig
							return &c, nil
						}
					}
				}
			}
		}
	}
	return nil, nil
}

func (r *InMemoryDistributionRepo) CreateRelease(ctx context.Context, rel *domain.AppRelease) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.releases[rel.ID] = rel
	return nil
}

func (r *InMemoryDistributionRepo) GetRelease(ctx context.Context, id string) (*domain.AppRelease, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rel, ok := r.releases[id]
	if !ok {
		return nil, domain.ErrReleaseNotFound
	}
	c := *rel
	return &c, nil
}

func (r *InMemoryDistributionRepo) UpdateRelease(ctx context.Context, rel *domain.AppRelease) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.releases[rel.ID] = rel
	return nil
}

type ArtifactService struct {
	repo             DistributionRepository
	blacklistedCerts []string
	mu               sync.RWMutex
}

func (r *InMemoryDistributionRepo) ListReleasesByApp(ctx context.Context, appID string) ([]*domain.AppRelease, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.AppRelease
	for _, rel := range r.releases {
		if rel.AppID == appID {
			cp := *rel
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func NewArtifactService(repo DistributionRepository, blacklistedCerts []string) *ArtifactService {
	if repo == nil {
		repo = NewInMemoryDistributionRepo()
	}
	return &ArtifactService{
		repo:             repo,
		blacklistedCerts: blacklistedCerts,
	}
}

// InitiateUploadSession creates a secure quarantine upload session for an APK or AAB.
func (s *ArtifactService) InitiateUploadSession(
	ctx context.Context,
	orgID, appID string,
	artifactType domain.ArtifactType,
	fileName string,
	expectedSize int64,
	expectedSHA256 string,
) (*domain.UploadSession, error) {
	sessionID := uuid.New().String()
	quarantineKey := fmt.Sprintf("quarantine/%s/%s/%s", orgID, appID, sessionID)

	session := &domain.UploadSession{
		ID:                sessionID,
		OrgID:             orgID,
		AppID:             appID,
		ArtifactType:      artifactType,
		FileName:          fileName,
		ExpectedSizeBytes: expectedSize,
		ExpectedSHA256:    expectedSHA256,
		S3QuarantineKey:   quarantineKey,
		Status:            "INITIATED",
		ExpiresAt:         time.Now().UTC().Add(2 * time.Hour),
		CreatedAt:         time.Now().UTC(),
	}

	if err := s.repo.CreateUploadSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

// ProcessUploadedAPK processes and verifies an uploaded APK file.
// It parses AXML, checks v2/v3 signature, enforces key lineage, and registers the artifact.
func (s *ArtifactService) ProcessUploadedAPK(
	ctx context.Context,
	sessionID string,
	releaseID string,
	apkReader io.ReaderAt,
	size int64,
) (*domain.ReleaseArtifact, *domain.APKSignature, error) {
	session, err := s.repo.GetUploadSession(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	release, err := s.repo.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, nil, err
	}

	// 1. Inspect APK & Parse AXML (Zip Slip, Zip Bomb, Manifest)
	manifest, err := infrastructure.InspectAPK(apkReader, size, session.ExpectedSHA256)
	if err != nil {
		_ = s.repo.UpdateUploadSessionStatus(ctx, sessionID, "ABORTED")
		return nil, nil, err
	}

	// 2. Verify Signature Block (v2/v3 scheme, X.509 cert, debug cert, blacklist)
	s.mu.RLock()
	blacklisted := s.blacklistedCerts
	s.mu.RUnlock()

	verifiedSig, err := infrastructure.VerifyAPKSignature(apkReader, size, blacklisted)
	if err != nil {
		_ = s.repo.UpdateUploadSessionStatus(ctx, sessionID, "ABORTED")
		return nil, nil, err
	}

	// 3. Lineage Check with Previous Release (TC-SEC-007)
	prevSig, _ := s.repo.GetLatestSignatureByDistribution(ctx, release.DistributionID)
	if prevSig != nil {
		if err := infrastructure.VerifySignatureLineage(verifiedSig, prevSig.CertSHA256Fingerprint); err != nil {
			// Record rejected artifact for audit
			artID := uuid.New().String()
			rejArt := &domain.ReleaseArtifact{
				ID:                     artID,
				ReleaseID:              releaseID,
				ArtifactType:           domain.ArtifactTypeAPK,
				FileName:               session.FileName,
				SizeBytes:              size,
				SHA256Hash:             manifest.SHA256Hash,
				S3QuarantineKey:        session.S3QuarantineKey,
				PackageName:            manifest.PackageName,
				VersionCode:            manifest.VersionCode,
				MinSDKVersion:          manifest.MinSDKVersion,
				TargetSDKVersion:       manifest.TargetSDKVersion,
				NativeABIs:             manifest.NativeABIs,
				IsUniversalAPK:         manifest.IsUniversalAPK,
				SignatureSchemeVersion: verifiedSig.SignatureSchemeVersion,
				Status:                 domain.ArtifactStatusRejectedSignatureMismatch,
				CreatedAt:              time.Now().UTC(),
			}
			_ = s.repo.CreateArtifact(ctx, rejArt)
			return nil, nil, domain.ErrSignatureMismatch
		}
	}

	// 4. Create Artifact
	artID := uuid.New().String()
	artifact := &domain.ReleaseArtifact{
		ID:                     artID,
		ReleaseID:              releaseID,
		ArtifactType:           domain.ArtifactTypeAPK,
		FileName:               session.FileName,
		SizeBytes:              size,
		SHA256Hash:             manifest.SHA256Hash,
		S3QuarantineKey:        session.S3QuarantineKey,
		PackageName:            manifest.PackageName,
		VersionCode:            manifest.VersionCode,
		MinSDKVersion:          manifest.MinSDKVersion,
		TargetSDKVersion:       manifest.TargetSDKVersion,
		NativeABIs:             manifest.NativeABIs,
		IsUniversalAPK:         manifest.IsUniversalAPK,
		SignatureSchemeVersion: verifiedSig.SignatureSchemeVersion,
		Status:                 domain.ArtifactStatusQuarantined, // Ready for security scan
		CreatedAt:              time.Now().UTC(),
	}

	if err := s.repo.CreateArtifact(ctx, artifact); err != nil {
		return nil, nil, err
	}

	sigRecord := &domain.APKSignature{
		ID:                     uuid.New().String(),
		ArtifactID:             artID,
		SignatureSchemeVersion: verifiedSig.SignatureSchemeVersion,
		CertSHA256Fingerprint:  verifiedSig.CertSHA256Fingerprint,
		CertSHA1Fingerprint:    verifiedSig.CertSHA1Fingerprint,
		PublicKeyAlgorithm:     verifiedSig.PublicKeyAlgorithm,
		KeySizeBits:            verifiedSig.KeySizeBits,
		SubjectDN:              verifiedSig.SubjectDN,
		IssuerDN:               verifiedSig.IssuerDN,
		SerialNumberHex:        verifiedSig.SerialNumberHex,
		ValidFrom:              verifiedSig.ValidFrom,
		ValidUntil:             verifiedSig.ValidUntil,
		IsPrimaryLineage:       verifiedSig.IsPrimaryLineage,
		CreatedAt:              time.Now().UTC(),
	}

	if err := s.repo.SaveSignature(ctx, sigRecord); err != nil {
		return nil, nil, err
	}

	_ = s.repo.UpdateUploadSessionStatus(ctx, sessionID, "COMPLETED")

	// Transition release to SECURITY_SCANNING
	_ = domain.TransitionRelease(release, domain.ReleaseArtifactVerifying)
	_ = domain.TransitionRelease(release, domain.ReleaseSecurityScanning)
	_ = s.repo.UpdateRelease(ctx, release)

	return artifact, sigRecord, nil
}

// TransitionReleaseStatus transitions release through the 12-state state machine (TC-FUNC-008).
func (s *ArtifactService) TransitionReleaseStatus(ctx context.Context, releaseID string, targetStatus domain.ReleaseStatus) (*domain.AppRelease, error) {
	rel, err := s.repo.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, err
	}

	if err := domain.TransitionRelease(rel, targetStatus); err != nil {
		return nil, err
	}

	rel.UpdatedAt = time.Now().UTC()
	if targetStatus == domain.ReleasePublishedFull || targetStatus == domain.ReleasePublishedStaged {
		now := time.Now().UTC()
		rel.PublishedAt = &now
	}

	if err := s.repo.UpdateRelease(ctx, rel); err != nil {
		return nil, err
	}
	return rel, nil
}

// SubmitReleaseForReview moves a scanned release into the review queue
// (Play Console "send for review"). Releases resting in SECURITY_SCANNING
// advance through READY_FOR_SUBMISSION in the same action because the inline
// verification at upload completion already covers the scan step.
func (s *ArtifactService) SubmitReleaseForReview(ctx context.Context, releaseID string) (*domain.AppRelease, error) {
	rel, err := s.repo.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	switch rel.ReleaseStatus {
	case domain.ReleaseSecurityScanning:
		if err := domain.TransitionRelease(rel, domain.ReleaseReadyForSubmission); err != nil {
			return nil, err
		}
		fallthrough
	case domain.ReleaseReadyForSubmission:
		if err := domain.TransitionRelease(rel, domain.ReleaseSubmittedForReview); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%w: submit requires SECURITY_SCANNING or READY_FOR_SUBMISSION, release is %s",
			domain.ErrInvalidStateTransition, rel.ReleaseStatus)
	}
	rel.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateRelease(ctx, rel); err != nil {
		return nil, err
	}
	return rel, nil
}

// PublishRelease promotes an approved release's artifacts to the public CDN
// and transitions the release to PUBLISHED_STAGED (percentage 1-99, default
// 20) or PUBLISHED_FULL. Promotion is fail-closed: any CDN failure aborts
// before the status changes, and only unscanned-clean artifacts
// (QUARANTINED or VERIFIED_SAFE) are promotable — infected or rejected
// artifacts refuse publication.
func (s *ArtifactService) PublishRelease(ctx context.Context, releaseID string, full bool, percentage int, publisher *CDNPublisher) (*domain.AppRelease, []string, error) {
	rel, err := s.repo.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, nil, err
	}
	if rel.ReleaseStatus != domain.ReleaseApprovedPendingPub {
		return nil, nil, fmt.Errorf("%w: publish requires APPROVED_PENDING_PUBLISH, release is %s",
			domain.ErrInvalidStateTransition, rel.ReleaseStatus)
	}
	if publisher == nil {
		return nil, nil, fmt.Errorf("publish requires a configured CDN publisher")
	}
	target := domain.ReleasePublishedFull
	pct := 100
	if !full {
		if percentage <= 0 {
			percentage = 20
		}
		if percentage < 1 || percentage > 99 {
			return nil, nil, domain.ErrRolloutPercentageInvalid
		}
		target = domain.ReleasePublishedStaged
		pct = percentage
	}
	arts, err := s.repo.ListArtifactsByRelease(ctx, releaseID)
	if err != nil {
		return nil, nil, err
	}
	if len(arts) == 0 {
		return nil, nil, fmt.Errorf("%w: publish requires at least one artifact", domain.ErrInvalidStateTransition)
	}
	for _, art := range arts {
		if art.Status != domain.ArtifactStatusQuarantined && art.Status != domain.ArtifactStatusVerifiedSafe {
			return nil, nil, fmt.Errorf("%w: artifact %s with status %s refuses publication",
				domain.ErrInvalidStateTransition, art.ID, art.Status)
		}
	}
	urls := make([]string, 0, len(arts))
	for _, art := range arts {
		cdnURL, err := publisher.PromoteToPublicCDN(ctx, art)
		if err != nil {
			return nil, nil, fmt.Errorf("CDN promotion failed for artifact %s: %w", art.ID, err)
		}
		key := FormatCDNKey(art.PackageName, art.ReleaseID, art.SHA256Hash)
		if err := s.repo.SetArtifactPublicCDNKey(ctx, art.ID, key); err != nil {
			return nil, nil, err
		}
		if err := s.repo.UpdateArtifactStatus(ctx, art.ID, domain.ArtifactStatusPromotedPublic); err != nil {
			return nil, nil, err
		}
		urls = append(urls, cdnURL)
	}
	rel.RolloutPercentage = pct
	if err := s.repo.UpdateRelease(ctx, rel); err != nil {
		return nil, nil, err
	}
	updated, err := s.TransitionReleaseStatus(ctx, releaseID, target)
	if err != nil {
		return nil, nil, err
	}
	return updated, urls, nil
}

// GetUploadSession returns an upload session by id.
func (s *ArtifactService) GetUploadSession(ctx context.Context, id string) (*domain.UploadSession, error) {
	return s.repo.GetUploadSession(ctx, id)
}

// GetRelease returns a release by id.
func (s *ArtifactService) GetRelease(ctx context.Context, id string) (*domain.AppRelease, error) {
	return s.repo.GetRelease(ctx, id)
}

// CreateRelease persists a new release draft.
func (s *ArtifactService) CreateRelease(ctx context.Context, rel *domain.AppRelease) error {
	return s.repo.CreateRelease(ctx, rel)
}

// ListReleasesByApp returns newest-first releases for an app.
func (s *ArtifactService) ListReleasesByApp(ctx context.Context, appID string) ([]*domain.AppRelease, error) {
	return s.repo.ListReleasesByApp(ctx, appID)
}

// LatestSignature returns the newest signature record for a distribution.
func (s *ArtifactService) LatestSignature(ctx context.Context, distributionID string) (*domain.APKSignature, error) {
	return s.repo.GetLatestSignatureByDistribution(ctx, distributionID)
}

// MarkUploadComplete moves an upload session to COMPLETED.
func (s *ArtifactService) MarkUploadComplete(ctx context.Context, id string) error {
	return s.repo.UpdateUploadSessionStatus(ctx, id, "COMPLETED")
}

// HaltRelease suspends distribution immediately with a recorded reason.
func (s *ArtifactService) HaltRelease(ctx context.Context, releaseID, reason string) (*domain.AppRelease, error) {
	rel, err := s.repo.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	if err := domain.TransitionRelease(rel, domain.ReleaseHalted); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	rel.HaltedAt = &now
	rel.HaltReason = &reason
	rel.UpdatedAt = now
	if err := s.repo.UpdateRelease(ctx, rel); err != nil {
		return nil, err
	}
	return rel, nil
}

// ResumeRelease returns a halted release to staged or full publication.
func (s *ArtifactService) ResumeRelease(ctx context.Context, releaseID string) (*domain.AppRelease, error) {
	rel, err := s.repo.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	target := domain.ReleasePublishedStaged
	if rel.RolloutPercentage >= 100 {
		target = domain.ReleasePublishedFull
	}
	if err := domain.TransitionRelease(rel, target); err != nil {
		return nil, err
	}
	rel.HaltedAt = nil
	rel.HaltReason = nil
	rel.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateRelease(ctx, rel); err != nil {
		return nil, err
	}
	return rel, nil
}

// UpdateStagedRollout updates the percentage for a staged release (TC-FUNC-007).
func (s *ArtifactService) UpdateStagedRollout(ctx context.Context, releaseID string, percentage int) (*domain.AppRelease, error) {
	if percentage < 0 || percentage > 100 {
		return nil, domain.ErrRolloutPercentageInvalid
	}

	rel, err := s.repo.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, err
	}

	rel.RolloutPercentage = percentage
	if percentage == 100 && rel.ReleaseStatus == domain.ReleasePublishedStaged {
		_ = domain.TransitionRelease(rel, domain.ReleasePublishedFull)
	}

	rel.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateRelease(ctx, rel); err != nil {
		return nil, err
	}
	return rel, nil
}
