package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
)

// seedPublishableRelease stores a release plus one quarantined artifact whose
// bytes exist in storage under the session quarantine key.
func seedPublishableRelease(t *testing.T, ctx context.Context, repo *application.InMemoryDistributionRepo, storage *application.InMemoryObjectStorage, releaseID string, status domain.ReleaseStatus) (string, string) {
	t.Helper()
	content := []byte("release-binary-" + releaseID)
	sha := domain.ComputeSHA256(content)
	qkey := "quarantine/org-1/app-1/" + releaseID
	if err := storage.PutObject(ctx, "bn-quarantine-bucket", &application.StorageObject{
		Key: qkey, Content: content, SHA256: sha, ContentType: "application/vnd.android.package-archive",
	}); err != nil {
		t.Fatalf("quarantine seed failed: %v", err)
	}
	rel := &domain.AppRelease{
		ID: releaseID, DistributionID: "dist-1", AppID: "app-1",
		ReleaseTrack: domain.TrackProduction, VersionCode: 42, VersionName: "4.2",
		ReleaseStatus: status, CreatedBy: "dev-1",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.CreateRelease(ctx, rel); err != nil {
		t.Fatalf("release seed failed: %v", err)
	}
	art := &domain.ReleaseArtifact{
		ID: "art-" + releaseID, ReleaseID: releaseID, ArtifactType: domain.ArtifactTypeAPK,
		FileName: "app-42.apk", SizeBytes: int64(len(content)), SHA256Hash: sha,
		S3QuarantineKey: qkey, PackageName: "com.acme.app", VersionCode: 42,
		MinSDKVersion: 26, TargetSDKVersion: 34, Status: domain.ArtifactStatusQuarantined,
		CreatedAt: time.Now().UTC(),
	}
	if err := repo.CreateArtifact(ctx, art); err != nil {
		t.Fatalf("artifact seed failed: %v", err)
	}
	return sha, qkey
}

func TestSubmitReleaseForReview(t *testing.T) {
	ctx := context.Background()
	repo := application.NewInMemoryDistributionRepo()
	storage := application.NewInMemoryObjectStorage()
	svc := application.NewArtifactService(repo, nil)
	seedPublishableRelease(t, ctx, repo, storage, "rel-submit-01", domain.ReleaseSecurityScanning)

	updated, err := svc.SubmitReleaseForReview(ctx, "rel-submit-01")
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if updated.ReleaseStatus != domain.ReleaseSubmittedForReview {
		t.Fatalf("expected SUBMITTED_FOR_REVIEW, got %s", updated.ReleaseStatus)
	}
	if _, err := svc.SubmitReleaseForReview(ctx, "rel-submit-01"); !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Fatalf("expected invalid-transition on double submit, got %v", err)
	}

	seedPublishableRelease(t, ctx, repo, storage, "rel-submit-02", domain.ReleaseDraft)
	if _, err := svc.SubmitReleaseForReview(ctx, "rel-submit-02"); !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Fatalf("expected invalid-transition submitting a DRAFT, got %v", err)
	}
}

func TestPublishReleaseStagedAndFull(t *testing.T) {
	ctx := context.Background()
	repo := application.NewInMemoryDistributionRepo()
	storage := application.NewInMemoryObjectStorage()
	svc := application.NewArtifactService(repo, nil)
	publisher := application.NewCDNPublisher(storage, "bn-quarantine-bucket", "bn-public-cdn-bucket", "cdn.example.local")

	seedPublishableRelease(t, ctx, repo, storage, "rel-pub-staged", domain.ReleaseApprovedPendingPub)
	rel, urls, err := svc.PublishRelease(ctx, "rel-pub-staged", false, 0, publisher)
	if err != nil {
		t.Fatalf("staged publish failed: %v", err)
	}
	if rel.ReleaseStatus != domain.ReleasePublishedStaged || rel.RolloutPercentage != 20 {
		t.Fatalf("expected PUBLISHED_STAGED at 20%%, got %s at %d", rel.ReleaseStatus, rel.RolloutPercentage)
	}
	if rel.PublishedAt == nil {
		t.Fatalf("expected PublishedAt set")
	}
	if len(urls) != 1 || !strings.HasPrefix(urls[0], "https://cdn.example.local/packages/com.acme.app/") {
		t.Fatalf("unexpected cdn urls %v", urls)
	}
	got, err := storage.GetObject(ctx, "bn-public-cdn-bucket", strings.TrimPrefix(urls[0], "https://cdn.example.local/"))
	if err != nil || string(got.Content) != "release-binary-rel-pub-staged" {
		t.Fatalf("promoted bytes missing or wrong: %v %+v", err, got)
	}
	arts, _ := repo.ListArtifactsByRelease(ctx, "rel-pub-staged")
	if len(arts) != 1 || arts[0].Status != domain.ArtifactStatusPromotedPublic || arts[0].S3PublicCDNKey == nil {
		t.Fatalf("expected promoted artifact record, got %+v", arts)
	}

	seedPublishableRelease(t, ctx, repo, storage, "rel-pub-full", domain.ReleaseApprovedPendingPub)
	rel, urls, err = svc.PublishRelease(ctx, "rel-pub-full", true, 0, publisher)
	if err != nil {
		t.Fatalf("full publish failed: %v", err)
	}
	if rel.ReleaseStatus != domain.ReleasePublishedFull || rel.RolloutPercentage != 100 || len(urls) != 1 {
		t.Fatalf("unexpected full publish: %s at %d urls=%v", rel.ReleaseStatus, rel.RolloutPercentage, urls)
	}

	if _, _, err := svc.PublishRelease(ctx, "rel-pub-full", true, 0, publisher); !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Fatalf("expected invalid-transition republishing a live release, got %v", err)
	}
}

func TestPublishReleaseFailClosed(t *testing.T) {
	ctx := context.Background()
	repo := application.NewInMemoryDistributionRepo()
	storage := application.NewInMemoryObjectStorage()
	svc := application.NewArtifactService(repo, nil)
	publisher := application.NewCDNPublisher(storage, "bn-quarantine-bucket", "bn-public-cdn-bucket", "cdn.example.local")

	// Not approved: refused, status untouched.
	seedPublishableRelease(t, ctx, repo, storage, "rel-unapproved", domain.ReleaseSubmittedForReview)
	if _, _, err := svc.PublishRelease(ctx, "rel-unapproved", true, 0, publisher); !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Fatalf("expected invalid-transition publishing unapproved release, got %v", err)
	}
	rel, _ := repo.GetRelease(ctx, "rel-unapproved")
	if rel.ReleaseStatus != domain.ReleaseSubmittedForReview {
		t.Fatalf("status must not move on refused publish, got %s", rel.ReleaseStatus)
	}

	// Quarantine bytes missing: CDN promotion fails, status untouched.
	seedPublishableRelease(t, ctx, repo, storage, "rel-nobytes", domain.ReleaseApprovedPendingPub)
	if err := storage.DeleteObject(ctx, "bn-quarantine-bucket", "quarantine/org-1/app-1/rel-nobytes"); err != nil {
		t.Fatalf("seed delete failed: %v", err)
	}
	if _, _, err := svc.PublishRelease(ctx, "rel-nobytes", true, 0, publisher); err == nil {
		t.Fatalf("expected promotion failure without quarantine bytes")
	}
	rel, _ = repo.GetRelease(ctx, "rel-nobytes")
	if rel.ReleaseStatus != domain.ReleaseApprovedPendingPub {
		t.Fatalf("status must not move on failed promotion, got %s", rel.ReleaseStatus)
	}

	// Infected artifact: refuses publication.
	seedPublishableRelease(t, ctx, repo, storage, "rel-infected", domain.ReleaseApprovedPendingPub)
	if err := repo.UpdateArtifactStatus(ctx, "art-rel-infected", domain.ArtifactStatusInfected); err != nil {
		t.Fatalf("status seed failed: %v", err)
	}
	if _, _, err := svc.PublishRelease(ctx, "rel-infected", true, 0, publisher); err == nil {
		t.Fatalf("expected refusal for infected artifact")
	}

	// No publisher configured: fail closed.
	seedPublishableRelease(t, ctx, repo, storage, "rel-nopub", domain.ReleaseApprovedPendingPub)
	if _, _, err := svc.PublishRelease(ctx, "rel-nopub", true, 0, nil); err == nil {
		t.Fatalf("expected failure without publisher")
	}
}
