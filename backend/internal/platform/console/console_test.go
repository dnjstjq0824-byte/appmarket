package console

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	catalogApp "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	catalogDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
	catalogInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/infrastructure"
	developerApp "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/application"
	developerDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
	developerInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/infrastructure"
	distApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	distDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
	searchDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/search/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/audit"
)

type recordingIndexer struct {
	docs []searchDomain.SearchDoc
	err  error
}

func (f *recordingIndexer) IndexDocument(_ context.Context, doc searchDomain.SearchDoc) error {
	f.docs = append(f.docs, doc)
	return f.err
}

type recordingObjectStorage struct {
	buckets []string
	puts    []*distApp.StorageObject
	putErr  error
}

func (f *recordingObjectStorage) GetObject(_ context.Context, _, _ string) (*distApp.StorageObject, error) {
	return nil, errors.New("not implemented")
}

func (f *recordingObjectStorage) PutObject(_ context.Context, bucket string, obj *distApp.StorageObject) error {
	if f.putErr != nil {
		return f.putErr
	}
	f.buckets = append(f.buckets, bucket)
	f.puts = append(f.puts, obj)
	return nil
}

func (f *recordingObjectStorage) DeleteObject(_ context.Context, _, _ string) error {
	return nil
}

func (f *recordingObjectStorage) TagObject(_ context.Context, _, _ string, _ map[string]string) error {
	return nil
}

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

func TestDecisionApprovalIndexesSearchDoc(t *testing.T) {
	ctx := context.Background()
	catalogRepo := catalogInfra.NewInMemoryCatalogRepository()
	catalogRepo.AddApp(
		&catalogDomain.CatalogApp{
			ID: "app-index-01", CanonicalSlug: "indexed-app", PrimaryPackageName: "com.acme.indexed",
			PrimaryCategoryID: "productivity", DefaultLocale: "ko-KR",
			CatalogStatus:  catalogDomain.CatalogStatusPendingInitialReview,
			BayesianRating: 4.5, InstallCount: 1200,
		},
		[]catalogDomain.Localization{
			{ID: "loc-1", AppID: "app-index-01", Locale: "ko-KR", Title: "색인 앱", FullDescription: "설명"},
		},
		nil, nil, nil, nil,
	)
	catalogRepo.AddApp(
		&catalogDomain.CatalogApp{
			ID: "app-index-02", CanonicalSlug: "rejected-app", PrimaryPackageName: "com.acme.rejected",
			CatalogStatus: catalogDomain.CatalogStatusPendingInitialReview,
		},
		nil, nil, nil, nil, nil,
	)
	catalogSvc := catalogApp.NewCatalogService(catalogRepo, catalogInfra.NewInMemoryCache())
	indexer := &recordingIndexer{}
	svc := NewDecisionService(audit.NewHashChainLogger(audit.NewInMemoryAuditRepository())).
		WithCatalog(catalogSvc).
		WithSearchIndexer(indexer)

	if _, err := svc.Decide(ctx, "app-index-01", "APPROVED", "ship it", nil, "admin-1", "", ""); err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if len(indexer.docs) != 1 {
		t.Fatalf("expected 1 indexed doc after approval, got %d", len(indexer.docs))
	}
	doc := indexer.docs[0]
	if doc.AppID != "app-index-01" || doc.PackageName != "com.acme.indexed" || doc.Category != "productivity" {
		t.Fatalf("unexpected indexed doc identity: %+v", doc)
	}
	if doc.TitleI18n["ko-KR"] != "색인 앱" || doc.DescriptionI18n["ko-KR"] != "설명" {
		t.Fatalf("unexpected indexed i18n: %+v / %+v", doc.TitleI18n, doc.DescriptionI18n)
	}
	if doc.BayesianRating != 4.5 || doc.DownloadCount != 1200 {
		t.Fatalf("unexpected indexed stats: %+v", doc)
	}

	if _, err := svc.Decide(ctx, "app-index-02", "REJECTED", "broken icon", nil, "admin-1", "", ""); err != nil {
		t.Fatalf("reject failed: %v", err)
	}
	if len(indexer.docs) != 1 {
		t.Fatalf("rejection must not index, got %d docs", len(indexer.docs))
	}
}

func TestDecisionApprovalSurvivesIndexerOutage(t *testing.T) {
	ctx := context.Background()
	catalogRepo := catalogInfra.NewInMemoryCatalogRepository()
	catalogRepo.AddApp(
		&catalogDomain.CatalogApp{
			ID: "app-index-03", CanonicalSlug: "outage-app", PrimaryPackageName: "com.acme.outage",
			CatalogStatus: catalogDomain.CatalogStatusPendingInitialReview,
		},
		nil, nil, nil, nil, nil,
	)
	catalogSvc := catalogApp.NewCatalogService(catalogRepo, catalogInfra.NewInMemoryCache())
	svc := NewDecisionService(audit.NewHashChainLogger(audit.NewInMemoryAuditRepository())).
		WithCatalog(catalogSvc).
		WithSearchIndexer(&recordingIndexer{err: errors.New("opensearch down")})

	d, err := svc.Decide(ctx, "app-index-03", "APPROVED", "ship it", nil, "admin-1", "", "")
	if err != nil {
		t.Fatalf("decision must succeed despite indexer outage: %v", err)
	}
	if d.Decision != "APPROVED" {
		t.Fatalf("unexpected decision: %+v", d)
	}
	got, _ := catalogRepo.GetAppBySlug(ctx, "outage-app")
	if got == nil || got.CatalogStatus != catalogDomain.CatalogStatusPublished {
		t.Fatalf("expected PUBLISHED after approve, got %+v", got)
	}
}

func TestUploadQuarantineStoresBytesBeforeScan(t *testing.T) {
	ctx := context.Background()
	_, artifacts, orgID, appID := consoleFixtures(t)
	store := &recordingObjectStorage{}
	uploads := NewUploadService(artifacts).WithQuarantineStorage(store, "bn-quarantine-bucket")

	garbage := []byte("not-an-apk")
	sum := sha256.Sum256(garbage)
	sha := hex.EncodeToString(sum[:])
	session, err := uploads.InitUpload(ctx, orgID, appID, "release.apk", int64(len(garbage)), sha)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// Garbage must still be rejected by the scanner, but only AFTER the bytes
	// landed in quarantine storage.
	if _, err := uploads.CompleteUpload(ctx, session.ID, sha, garbage); err == nil {
		t.Fatalf("expected scanner rejection for garbage bytes")
	}
	if len(store.puts) != 1 || store.buckets[0] != "bn-quarantine-bucket" {
		t.Fatalf("expected 1 quarantine put, got %d in %v", len(store.puts), store.buckets)
	}
	put := store.puts[0]
	if put.Key != session.S3QuarantineKey || string(put.Content) != string(garbage) {
		t.Fatalf("unexpected quarantined object: %+v", put)
	}
	if put.SHA256 != sha || put.Tags["status"] != "quarantined" || put.Tags["app_id"] != appID {
		t.Fatalf("unexpected quarantine metadata: %+v / %+v", put.SHA256, put.Tags)
	}
}

func TestUploadQuarantineStoreFailureFailsClosed(t *testing.T) {
	ctx := context.Background()
	_, artifacts, orgID, appID := consoleFixtures(t)
	uploads := NewUploadService(artifacts).
		WithQuarantineStorage(&recordingObjectStorage{putErr: errors.New("s3 down")}, "bn-quarantine-bucket")

	payload := []byte("payload")
	sum := sha256.Sum256(payload)
	sha := hex.EncodeToString(sum[:])
	session, err := uploads.InitUpload(ctx, orgID, appID, "release.apk", int64(len(payload)), sha)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := uploads.CompleteUpload(ctx, session.ID, sha, payload); err != ErrQuarantineStore {
		t.Fatalf("expected ErrQuarantineStore, got %v", err)
	}
}

func seedConsoleRelease(t *testing.T, ctx context.Context, repo *distApp.InMemoryDistributionRepo, storage *distApp.InMemoryObjectStorage, appID, releaseID string, status distDomain.ReleaseStatus) {
	t.Helper()
	content := []byte("console-binary-" + releaseID)
	sha := distDomain.ComputeSHA256(content)
	qkey := "quarantine/org-1/" + appID + "/" + releaseID
	if err := storage.PutObject(ctx, "bn-quarantine-bucket", &distApp.StorageObject{
		Key: qkey, Content: content, SHA256: sha, ContentType: "application/vnd.android.package-archive",
	}); err != nil {
		t.Fatalf("quarantine seed failed: %v", err)
	}
	if err := repo.CreateRelease(ctx, &distDomain.AppRelease{
		ID: releaseID, DistributionID: "dist-" + appID, AppID: appID,
		ReleaseTrack: distDomain.TrackProduction, VersionCode: 7, VersionName: "1.0.7",
		ReleaseStatus: status, CreatedBy: "dev-1",
	}); err != nil {
		t.Fatalf("release seed failed: %v", err)
	}
	if err := repo.CreateArtifact(ctx, &distDomain.ReleaseArtifact{
		ID: "art-" + releaseID, ReleaseID: releaseID, ArtifactType: distDomain.ArtifactTypeAPK,
		FileName: "app-7.apk", SizeBytes: int64(len(content)), SHA256Hash: sha,
		S3QuarantineKey: qkey, PackageName: "com.acme.app", VersionCode: 7,
		Status: distDomain.ArtifactStatusQuarantined,
	}); err != nil {
		t.Fatalf("artifact seed failed: %v", err)
	}
}

func TestRolloutSubmitAndPublish(t *testing.T) {
	ctx := context.Background()
	distRepo := distApp.NewInMemoryDistributionRepo()
	artifacts := distApp.NewArtifactService(distRepo, nil)
	storage := distApp.NewInMemoryObjectStorage()
	publisher := distApp.NewCDNPublisher(storage, "bn-quarantine-bucket", "bn-public-cdn-bucket", "cdn.example.local")
	rollouts := NewRolloutService(artifacts).WithCDNPublisher(publisher)

	seedConsoleRelease(t, ctx, distRepo, storage, "app-roll-01", "rel-roll-01", distDomain.ReleaseSecurityScanning)
	submitted, err := rollouts.SubmitForReview(ctx, "app-roll-01")
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if submitted.ReleaseStatus != distDomain.ReleaseSubmittedForReview {
		t.Fatalf("expected SUBMITTED_FOR_REVIEW, got %s", submitted.ReleaseStatus)
	}
	if _, err := artifacts.TransitionReleaseStatus(ctx, "rel-roll-01", distDomain.ReleaseInReview); err != nil {
		t.Fatalf("review step failed: %v", err)
	}
	if _, err := artifacts.TransitionReleaseStatus(ctx, "rel-roll-01", distDomain.ReleaseApprovedPendingPub); err != nil {
		t.Fatalf("approve step failed: %v", err)
	}
	result, err := rollouts.Publish(ctx, "app-roll-01", false, 0)
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if result.Release.ReleaseStatus != distDomain.ReleasePublishedStaged || result.Rollout != 20 || len(result.CDNURLs) != 1 {
		t.Fatalf("unexpected publish result: %+v", result)
	}
}

func TestRolloutPublishFailsClosed(t *testing.T) {
	ctx := context.Background()
	distRepo := distApp.NewInMemoryDistributionRepo()
	artifacts := distApp.NewArtifactService(distRepo, nil)
	storage := distApp.NewInMemoryObjectStorage()
	publisher := distApp.NewCDNPublisher(storage, "bn-quarantine-bucket", "bn-public-cdn-bucket", "cdn.example.local")

	// Not approved: 409-class precondition error.
	seedConsoleRelease(t, ctx, distRepo, storage, "app-roll-02", "rel-roll-02", distDomain.ReleaseSecurityScanning)
	rollouts := NewRolloutService(artifacts).WithCDNPublisher(publisher)
	if _, err := rollouts.Publish(ctx, "app-roll-02", true, 0); err != ErrReleaseNotReady {
		t.Fatalf("expected ErrReleaseNotReady, got %v", err)
	}
	// No publisher configured: 503-class failure.
	seedConsoleRelease(t, ctx, distRepo, storage, "app-roll-03", "rel-roll-03", distDomain.ReleaseApprovedPendingPub)
	bare := NewRolloutService(artifacts)
	if _, err := bare.Publish(ctx, "app-roll-03", true, 0); err != ErrReleasePublishFail {
		t.Fatalf("expected ErrReleasePublishFail without publisher, got %v", err)
	}
	// No release at all: ErrNoRelease surfaces unchanged.
	if _, err := rollouts.Publish(ctx, "app-missing", true, 0); err != ErrNoRelease {
		t.Fatalf("expected ErrNoRelease, got %v", err)
	}
}

func TestDecisionAdvancesRelease(t *testing.T) {
	ctx := context.Background()
	catalogRepo := catalogInfra.NewInMemoryCatalogRepository()
	catalogRepo.AddApp(
		&catalogDomain.CatalogApp{
			ID: "app-dec-rel-01", CanonicalSlug: "dec-rel-app", PrimaryPackageName: "com.acme.decrel",
			CatalogStatus: catalogDomain.CatalogStatusPendingInitialReview,
		},
		nil, nil, nil, nil, nil,
	)
	catalogRepo.AddApp(
		&catalogDomain.CatalogApp{
			ID: "app-dec-rel-02", CanonicalSlug: "dec-rel-app-2", PrimaryPackageName: "com.acme.decrel2",
			CatalogStatus: catalogDomain.CatalogStatusPendingInitialReview,
		},
		nil, nil, nil, nil, nil,
	)
	catalogRepo.AddApp(
		&catalogDomain.CatalogApp{
			ID: "app-dec-rel-03", CanonicalSlug: "dec-rel-app-3", PrimaryPackageName: "com.acme.decrel3",
			CatalogStatus: catalogDomain.CatalogStatusPendingInitialReview,
		},
		nil, nil, nil, nil, nil,
	)
	catalogSvc := catalogApp.NewCatalogService(catalogRepo, catalogInfra.NewInMemoryCache())
	distRepo := distApp.NewInMemoryDistributionRepo()
	artifacts := distApp.NewArtifactService(distRepo, nil)
	storage := distApp.NewInMemoryObjectStorage()
	seedConsoleRelease(t, ctx, distRepo, storage, "app-dec-rel-01", "rel-dec-01", distDomain.ReleaseSubmittedForReview)
	seedConsoleRelease(t, ctx, distRepo, storage, "app-dec-rel-02", "rel-dec-02", distDomain.ReleaseInReview)

	svc := NewDecisionService(audit.NewHashChainLogger(audit.NewInMemoryAuditRepository())).
		WithCatalog(catalogSvc).
		WithReleases(artifacts)

	if _, err := svc.Decide(ctx, "app-dec-rel-01", "APPROVED", "ship it", nil, "admin-1", "", ""); err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	rel, _ := distRepo.GetRelease(ctx, "rel-dec-01")
	if rel.ReleaseStatus != distDomain.ReleaseApprovedPendingPub {
		t.Fatalf("expected APPROVED_PENDING_PUBLISH, got %s", rel.ReleaseStatus)
	}
	if _, err := svc.Decide(ctx, "app-dec-rel-02", "REJECTED", "broken icon", nil, "admin-1", "", ""); err != nil {
		t.Fatalf("reject failed: %v", err)
	}
	rel, _ = distRepo.GetRelease(ctx, "rel-dec-02")
	if rel.ReleaseStatus != distDomain.ReleaseChangesRequested {
		t.Fatalf("expected CHANGES_REQUESTED, got %s", rel.ReleaseStatus)
	}
	// No releases for the app: the app verdict still records (best-effort).
	if _, err := svc.Decide(ctx, "app-dec-rel-03", "APPROVED", "ship it", nil, "admin-1", "", ""); err != nil {
		t.Fatalf("approve without releases must succeed: %v", err)
	}
}
