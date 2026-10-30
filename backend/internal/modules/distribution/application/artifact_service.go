package application

import (
	"context"
	"fmt"
	"io"
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

	SaveSignature(ctx context.Context, sig *domain.APKSignature) error
	GetLatestSignatureByDistribution(ctx context.Context, distributionID string) (*domain.APKSignature, error)

	CreateRelease(ctx context.Context, r *domain.AppRelease) error
	GetRelease(ctx context.Context, id string) (*domain.AppRelease, error)
	UpdateRelease(ctx context.Context, r *domain.AppRelease) error
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

