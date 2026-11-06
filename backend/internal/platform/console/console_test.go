package console

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	catalogApp "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	catalogDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
	catalogInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/infrastructure"
	developerApp "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/application"
	developerInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/infrastructure"
	developerDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
	distApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	distDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/audit"
)

func consoleFixtures(t *testing.T) (*OrgAppResolver, *distApp.ArtifactService, string, string) {
	t.Helper()
	ctx := context.Background()
	devRepo := developerInfra.NewInMemoryDeveloperRepository()
	orgs := developerApp.NewOrgService(devRepo)
	members := developerApp.NewMemberService(devRepo)
	org, err := orgs.CreateOrg(ctx, "acme", "Acme")
	if err != nil {
		t.Fatalf("org create failed: %v", err)
	}
	if _, err := members.AddMember(ctx, org.ID, "user-1", developerDomain.RoleDevOwner); err != nil {
		t.Fatalf("member add failed: %v", err)
	}
	catalogRepo := catalogInfra.NewInMemoryCatalogRepository()
	orgIDCopy := org.ID
	catalogRepo.AddApp(
		&catalogDomain.CatalogApp{ID: "app-acme-01", CanonicalSlug: "acme-app", PrimaryPackageName: "com.acme.app", DeveloperOrgID: &orgIDCopy, CatalogStatus: catalogDomain.CatalogStatusPublished},
		nil, nil, nil, nil, nil,
	)
	catalogSvc := catalogApp.NewCatalogService(catalogRepo, catalogInfra.NewInMemoryCache())
	resolver := NewOrgAppResolver(orgs, members, catalogSvc)
	artifacts := distApp.NewArtifactService(distApp.NewInMemoryDistributionRepo(), nil)
	return resolver, artifacts, org.ID, "app-omni-01"
}

func TestResolverCurrentHandles(t *testing.T) {
	ctx := context.Background()
	resolver, _, orgID, _ := consoleFixtures(t)
	gotOrg, gotApp, err := resolver.ResolveOrgApp(ctx, "user-1", "current", "current")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	_ = gotApp
	if gotOrg != orgID {
		t.Fatalf("expected org %s, got %s", orgID, gotOrg)
	}
	if _, _, err := resolver.ResolveOrgApp(ctx, "stranger", "", ""); err != ErrNoOrg {
		t.Fatalf("expected no-org for stranger, got %v", err)
	}
	if _, _, err := resolver.ResolveOrgApp(ctx, "stranger", orgID, ""); err != ErrOrgAccessDenied {
		t.Fatalf("expected access denied, got %v", err)
	}
}

func TestUploadInitValidation(t *testing.T) {
	ctx := context.Background()
	resolver, artifacts, orgID, appID := consoleFixtures(t)
	_ = resolver
	uploads := NewUploadService(artifacts)
	sha := strings.Repeat("a", 64)
	session, err := uploads.InitUpload(ctx, orgID, appID, "release.apk", 1024, sha)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if session.ID == "" || session.Status != "INITIATED" {
		t.Fatalf("unexpected session: %+v", session)
	}
	for _, tc := range []struct {
		name string
		file string
		size int64
		sha  string
	}{
		{"bad extension", "release.zip", 1024, sha},
		{"empty name", "", 1024, sha},
		{"zero size", "release.apk", 0, sha},
		{"oversize", "release.apk", 3 << 30, sha},
		{"short sha", "release.apk", 1024, "abc"},
		{"non-hex sha", "release.apk", 1024, strings.Repeat("z", 64)},
	} {
		if _, err := uploads.InitUpload(ctx, orgID, appID, tc.file, tc.size, tc.sha); err == nil {
			t.Fatalf("%s: expected validation error", tc.name)
		}
	}
	res, err := uploads.CompleteUpload(ctx, session.ID, sha, nil)
	if err != nil {
		t.Fatalf("metadata-only complete failed: %v", err)
	}
	if res.Status != "COMPLETED" || res.Details != nil {
		t.Fatalf("unexpected receipt: %+v", res)
	}
	if _, err := uploads.CompleteUpload(ctx, session.ID, sha, nil); err == nil {
		t.Fatalf("expected state error on double complete")
	}
	if _, err := uploads.CompleteUpload(ctx, "missing", sha, nil); err != ErrUploadNotFound {
		t.Fatalf("expected not-found, got %v", err)
	}
}

func TestUploadBinaryPathRejectsGarbage(t *testing.T) {
	ctx := context.Background()
	_, artifacts, orgID, appID := consoleFixtures(t)
	uploads := NewUploadService(artifacts)
	garbage := []byte("not-an-apk")
	sum := sha256.Sum256(garbage)
	sha := hex.EncodeToString(sum[:])
	session, err := uploads.InitUpload(ctx, orgID, appID, "evil.apk", int64(len(garbage)), sha)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := uploads.CompleteUpload(ctx, session.ID, sha, garbage); err == nil {
		t.Fatalf("expected pipeline rejection for garbage bytes")
	}
}

func TestRolloutHaltResumeFlow(t *testing.T) {
	ctx := context.Background()
	_, artifacts, _, appID := consoleFixtures(t)
	rollouts := NewRolloutService(artifacts)
	if _, err := rollouts.CurrentRelease(ctx, appID); err != ErrNoRelease {
		t.Fatalf("expected no-release, got %v", err)
	}
	now := time.Now().UTC()
	rel := &distDomain.AppRelease{
		ID: "rel-1", DistributionID: "dist-" + appID, AppID: appID,
		ReleaseTrack: distDomain.TrackProduction, VersionCode: 10, VersionName: "1.0",
		ReleaseStatus: distDomain.ReleasePublishedStaged, RolloutPercentage: 10,
		CreatedBy: "user-1", CreatedAt: now, UpdatedAt: now,
	}
	if err := artifacts.CreateRelease(ctx, rel); err != nil {
		t.Fatalf("create release failed: %v", err)
	}
	updated, err := rollouts.UpdateRollout(ctx, appID, 50)
	if err != nil || updated.RolloutPercentage != 50 {
		t.Fatalf("expected 50%%, got %+v / %v", updated, err)
	}
	if _, err := rollouts.UpdateRollout(ctx, appID, 40); err == nil {
		t.Fatalf("expected rejection for decreasing percentage")
	}
	halted, err := rollouts.Halt(ctx, appID, "bad crash rate")
	if err != nil || halted.ReleaseStatus != distDomain.ReleaseHalted {
		t.Fatalf("expected halted, got %+v / %v", halted, err)
	}
	if _, err := rollouts.UpdateRollout(ctx, appID, 60); err == nil {
		t.Fatalf("expected rejection while halted")
	}
	resumed, err := rollouts.Resume(ctx, appID)
	if err != nil || resumed.ReleaseStatus != distDomain.ReleasePublishedStaged {
		t.Fatalf("expected resume to staged, got %+v / %v", resumed, err)
	}
	if _, err := rollouts.Resume(ctx, appID); err == nil {
		t.Fatalf("expected rejection when not halted")
	}
}

func TestDecisionFlipsCatalogStatus(t *testing.T) {
	ctx := context.Background()
	repo := audit.NewInMemoryAuditRepository()
	catalogRepo := catalogInfra.NewInMemoryCatalogRepository()
	catalogRepo.AddApp(
		&catalogDomain.CatalogApp{ID: "app-review-01", CanonicalSlug: "review-app", PrimaryPackageName: "com.acme.review", CatalogStatus: catalogDomain.CatalogStatusPendingInitialReview},
		nil, nil, nil, nil, nil,
	)
	catalogSvc := catalogApp.NewCatalogService(catalogRepo, catalogInfra.NewInMemoryCache())
	svc := NewDecisionService(audit.NewHashChainLogger(repo)).WithCatalog(catalogSvc)
	queue, err := catalogSvc.ListReviewQueue(ctx)
	if err != nil || len(queue) != 1 {
		t.Fatalf("expected 1 queued app, got %d err=%v", len(queue), err)
	}
	if _, err := svc.Decide(ctx, "ghost-app", "APPROVED", "x", nil, "admin-1", "", ""); err != catalogDomain.ErrAppNotFound {
		t.Fatalf("expected app-not-found for ghost decision, got %v", err)
	}
	if _, err := svc.Decide(ctx, "app-review-01", "APPROVED", "ship it", nil, "admin-1", "", ""); err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	got, _ := catalogRepo.GetAppBySlug(ctx, "review-app")
	if got == nil || got.CatalogStatus != catalogDomain.CatalogStatusPublished {
		t.Fatalf("expected PUBLISHED after approve, got %+v", got)
	}
	queue, _ = catalogSvc.ListReviewQueue(ctx)
	if len(queue) != 0 {
		t.Fatalf("expected empty queue after approve, got %d", len(queue))
	}
	catalogRepo.AddApp(
		&catalogDomain.CatalogApp{ID: "app-review-02", CanonicalSlug: "review-app-2", PrimaryPackageName: "com.acme.review2", CatalogStatus: catalogDomain.CatalogStatusPendingInitialReview},
		nil, nil, nil, nil, nil,
	)
	if _, err := svc.Decide(ctx, "app-review-02", "REJECTED", "broken icon", nil, "admin-1", "", ""); err != nil {
		t.Fatalf("reject failed: %v", err)
	}
	got2, _ := catalogRepo.GetAppBySlug(ctx, "review-app-2")
	if got2 == nil || got2.CatalogStatus != catalogDomain.CatalogStatusDraft {
		t.Fatalf("expected DRAFT after reject, got %+v", got2)
	}
}

func TestDecisionsValidateAndAudit(t *testing.T) {
	ctx := context.Background()
	repo := audit.NewInMemoryAuditRepository()
	svc := NewDecisionService(audit.NewHashChainLogger(repo))
	d, err := svc.Decide(ctx, "app-1", "APPROVED", "looks good", map[string]bool{"icons": true}, "admin-1", "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("decide failed: %v", err)
	}
	if d.Decision != "APPROVED" || d.DecidedBy != "admin-1" {
		t.Fatalf("unexpected decision: %+v", d)
	}
	if _, err := svc.Decide(ctx, "app-2", "REJECTED", "", nil, "admin-1", "", ""); err == nil {
		t.Fatalf("expected notes requirement for rejection")
	}
	if _, err := svc.Decide(ctx, "app-2", "MAYBE", "x", nil, "admin-1", "", ""); err == nil {
		t.Fatalf("expected verdict validation error")
	}
	got, err := svc.GetDecision(ctx, "app-1")
	if err != nil || got.AppID != "app-1" {
		t.Fatalf("expected stored decision, got %+v / %v", got, err)
	}
	records, err := repo.GetAllRecords()
	if err != nil || len(records) != 1 {
		t.Fatalf("expected 1 audit record, got %d / %v", len(records), err)
	}
}
