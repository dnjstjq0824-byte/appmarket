package console

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"

	catalogApp "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	catalogDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
	catalogInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/infrastructure"
	developerApp "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/application"
	developerDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
	developerInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/infrastructure"
	distApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
)

func mediaFixtures(t *testing.T) (*MetadataService, distApp.ObjectStorage, string, string) {
	t.Helper()
	ctx := context.Background()
	devRepo := developerInfra.NewInMemoryDeveloperRepository()
	orgs := developerApp.NewOrgService(devRepo)
	members := developerApp.NewMemberService(devRepo)
	org, err := orgs.CreateOrg(ctx, "acme-media", "Acme Media")
	if err != nil {
		t.Fatalf("org create failed: %v", err)
	}
	if _, err := members.AddMember(ctx, org.ID, "user-1", developerDomain.RoleDevOwner); err != nil {
		t.Fatalf("member add failed: %v", err)
	}
	catalogRepo := catalogInfra.NewInMemoryCatalogRepository()
	orgIDCopy := org.ID
	catalogRepo.AddApp(
		&catalogDomain.CatalogApp{ID: "app-media-01", CanonicalSlug: "media-app", PrimaryPackageName: "com.acme.media", DeveloperOrgID: &orgIDCopy, CatalogStatus: catalogDomain.CatalogStatusPublished},
		nil, nil, nil, nil, nil,
	)
	catalogSvc := catalogApp.NewCatalogService(catalogRepo, catalogInfra.NewInMemoryCache())
	resolver := NewOrgAppResolver(orgs, members, catalogSvc)
	storage := distApp.NewInMemoryObjectStorage()
	metadata := NewMetadataService(resolver, catalogSvc).WithMediaStorage(storage, "bn-public-media")
	return metadata, storage, org.ID, "app-media-01"
}

func encodeTestPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{R: 200, G: 30, B: 30, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode failed: %v", err)
	}
	return buf.Bytes()
}

func TestSaveMediaImageRoundtrip(t *testing.T) {
	ctx := context.Background()
	metadata, storage, orgID, appID := mediaFixtures(t)

	raw := encodeTestPNG(t, 1080, 1920)
	asset, err := metadata.SaveMediaImage(ctx, "user-1", orgID, appID, "ko-KR", "PHONE", "SCREENSHOT", "shot.png", raw)
	if err != nil {
		t.Fatalf("save media image failed: %v", err)
	}
	if asset.Width != 1080 || asset.Height != 1920 {
		t.Fatalf("dims not decoded from bytes: %+v", asset)
	}
	if len(asset.SHA256) != 64 {
		t.Fatalf("sha not computed: %+v", asset)
	}
	stored, err := storage.GetObject(ctx, "bn-public-media", asset.URL)
	if err != nil || !bytes.Equal(stored.Content, raw) {
		t.Fatalf("bytes not stored under key %q: %v", asset.URL, err)
	}

	listed, err := metadata.ListMedia(ctx, "user-1", orgID, appID, "ko-KR")
	if err != nil || len(listed) != 1 {
		t.Fatalf("expected 1 listed asset, got %d err=%v", len(listed), err)
	}
	if err := metadata.DeleteMedia(ctx, "user-1", orgID, appID, asset.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := storage.GetObject(ctx, "bn-public-media", asset.URL); err == nil {
		t.Fatal("expected stored bytes removed after delete")
	}
}

func TestSaveMediaImageRejects(t *testing.T) {
	ctx := context.Background()
	metadata, _, orgID, appID := mediaFixtures(t)

	if _, err := metadata.SaveMediaImage(ctx, "user-1", orgID, appID, "ko-KR", "PHONE", "SCREENSHOT", "evil.txt", []byte("not an image")); err == nil {
		t.Fatal("expected non-image rejection, got nil")
	}
	if _, err := metadata.SaveMediaImage(ctx, "user-1", orgID, appID, "ko-KR", "PHONE", "ICON", "a.png", encodeTestPNG(t, 4, 4)); err == nil {
		t.Fatal("expected bad asset-type rejection, got nil")
	}
	if _, err := metadata.SaveMediaImage(ctx, "stranger", orgID, appID, "ko-KR", "PHONE", "SCREENSHOT", "a.png", encodeTestPNG(t, 4, 4)); err == nil {
		t.Fatal("expected tenant denial for stranger, got nil")
	}
	oversize := bytes.Repeat([]byte{0x89}, 21<<20)
	if _, err := metadata.SaveMediaImage(ctx, "user-1", orgID, appID, "ko-KR", "PHONE", "SCREENSHOT", "big.png", oversize); err == nil {
		t.Fatal("expected oversize rejection, got nil")
	}
}
