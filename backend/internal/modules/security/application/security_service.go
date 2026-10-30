package application

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/security/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/security/scanner"
	"github.com/google/uuid"
)

type SecurityRepository interface {
	SaveScanReport(ctx context.Context, report *domain.SecurityScanReport) error
	GetScanReportByArtifact(ctx context.Context, artifactID string) (*domain.SecurityScanReport, error)
	SetKillSwitch(ctx context.Context, state *domain.KillSwitchState) error
	GetKillSwitch(ctx context.Context, appID string) (*domain.KillSwitchState, error)
}

type InMemorySecurityRepo struct {
	mu           sync.RWMutex
	reports      map[string]*domain.SecurityScanReport
	killSwitches map[string]*domain.KillSwitchState
}

func NewInMemorySecurityRepo() *InMemorySecurityRepo {
	return &InMemorySecurityRepo{
		reports:      make(map[string]*domain.SecurityScanReport),
		killSwitches: make(map[string]*domain.KillSwitchState),
	}
}

func (r *InMemorySecurityRepo) SaveScanReport(ctx context.Context, report *domain.SecurityScanReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reports[report.ArtifactID] = report
	return nil
}

func (r *InMemorySecurityRepo) GetScanReportByArtifact(ctx context.Context, artifactID string) (*domain.SecurityScanReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rep, ok := r.reports[artifactID]
	if !ok {
		return nil, fmt.Errorf("scan report not found")
	}
	c := *rep
	return &c, nil
}

func (r *InMemorySecurityRepo) SetKillSwitch(ctx context.Context, state *domain.KillSwitchState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.killSwitches[state.AppID] = state
	return nil
}

func (r *InMemorySecurityRepo) GetKillSwitch(ctx context.Context, appID string) (*domain.KillSwitchState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	state, ok := r.killSwitches[appID]
	if !ok {
		return nil, nil
	}
	c := *state
	return &c, nil
}

type SecurityService struct {
	repo           SecurityRepository
	malwareScanner *scanner.MalwareScanner
	engineVersion  string
	mu             sync.RWMutex
}

func NewSecurityService(repo SecurityRepository) *SecurityService {
	if repo == nil {
		repo = NewInMemorySecurityRepo()
	}
	return &SecurityService{
		repo:           repo,
		malwareScanner: scanner.NewMalwareScanner(),
		engineVersion:  "ClamAV-1.4.1/YARA-4.3.2-BN-Sec-Core",
	}
}

// PerformSecurityScan analyzes an artifact for malware, dangerous permissions, and brand impersonation.
func (s *SecurityService) PerformSecurityScan(
	ctx context.Context,
	artifactID string,
	appTitle string,
	packageName string,
	orgID string,
	permissions []string,
	artifactReader io.Reader,
) (*domain.SecurityScanReport, error) {
	// 1. Scan for malware (EICAR / YARA)
	scanResult, err := s.malwareScanner.ScanArtifact(ctx, artifactReader)
	if err != nil {
		return nil, err
	}

	// 2. Score permissions
	permResult := scanner.CalculatePermissionRiskScore(permissions)

	// 3. Detect impersonation
	impResult := scanner.DetectBrandImpersonation(appTitle, packageName, orgID)

	// Composite verdict determination
	verdict := domain.VerdictCleanLowRisk
	if scanResult.IsInfected {
		verdict = domain.VerdictInfectedQuarantined
	} else if permResult.RequiresManualReview || impResult.IsImpersonating {
		verdict = domain.VerdictCleanHighRiskManual
	}

	report := &domain.SecurityScanReport{
		ID:                   uuid.New().String(),
		ArtifactID:           artifactID,
		ScannerEngineVersion: s.engineVersion,
		ClamAVStatus:         scanResult.ClamAVStatus,
		YARAMatchedRules:     scanResult.YARAMatchedRules,
		DangerousPermissions: permResult.DangerousPermissions,
		ImpersonationScore:   impResult.SimilarityScore,
		RiskScore:            permResult.TotalRiskScore,
		Verdict:              verdict,
		RawReport:            scanResult.Details,
		ScannedAt:            time.Now().UTC(),
	}

	if err := s.repo.SaveScanReport(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

// TriggerKillSwitch activates the emergency kill-switch for an app (TC-SEC-010).
func (s *SecurityService) TriggerKillSwitch(ctx context.Context, appID, reason, adminID string) error {
	state := &domain.KillSwitchState{
		AppID:       appID,
		IsTriggered: true,
		Reason:      reason,
		TriggeredBy: adminID,
		TriggeredAt: time.Now().UTC(),
	}
	return s.repo.SetKillSwitch(ctx, state)
}

// RevokeKillSwitch deactivates the emergency kill-switch.
func (s *SecurityService) RevokeKillSwitch(ctx context.Context, appID string) error {
	state := &domain.KillSwitchState{
		AppID:       appID,
		IsTriggered: false,
		Reason:      "Resolved by admin",
		TriggeredAt: time.Now().UTC(),
	}
	return s.repo.SetKillSwitch(ctx, state)
}

// ValidateDownloadRequest verifies whether the app is allowed for download or blocked by Kill Switch (TC-SEC-010).
func (s *SecurityService) ValidateDownloadRequest(ctx context.Context, appID string) error {
	state, err := s.repo.GetKillSwitch(ctx, appID)
	if err != nil {
		return err
	}
	if state != nil && state.IsTriggered {
		return domain.ErrKillSwitchActive
	}
	return nil
}

