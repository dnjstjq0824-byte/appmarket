package application_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
)

func TestSaveMediaAssetValidation(t *testing.T) {
	svc, _, _ := setupCatalog(t)
	ctx := context.Background()
	sha := strings.Repeat("b", 64)

	saved, err := svc.SaveMediaAsset(ctx, domain.MediaAsset{
		AppID: "app-media-01", Locale: "ko-KR", DeviceFormFactor: "PHONE",
		AssetType: "SCREENSHOT", URL: "media/app-media-01/shot.png",
		Width: 1080, Height: 1920, SHA256: sha, DisplayOrder: 0,
	})
	if err != nil {
		t.Fatalf("valid asset save failed: %v", err)
	}
	if saved.ID == "" || saved.AppID != "app-media-01" {
		t.Fatalf("unexpected saved asset: %+v", saved)
	}

	for _, tc := range []struct {
		name  string
		asset domain.MediaAsset
	}{
		{"bad asset type", domain.MediaAsset{AppID: "a", Locale: "ko-KR", DeviceFormFactor: "PHONE", AssetType: "ICON", URL: "u", Width: 10, Height: 10, SHA256: sha}},
		{"bad form factor", domain.MediaAsset{AppID: "a", Locale: "ko-KR", DeviceFormFactor: "WATCH", AssetType: "SCREENSHOT", URL: "u", Width: 10, Height: 10, SHA256: sha}},
		{"empty url", domain.MediaAsset{AppID: "a", Locale: "ko-KR", DeviceFormFactor: "PHONE", AssetType: "SCREENSHOT", Width: 10, Height: 10, SHA256: sha}},
		{"zero dims", domain.MediaAsset{AppID: "a", Locale: "ko-KR", DeviceFormFactor: "PHONE", AssetType: "SCREENSHOT", URL: "u", Width: 0, Height: 10, SHA256: sha}},
		{"bad sha", domain.MediaAsset{AppID: "a", Locale: "ko-KR", DeviceFormFactor: "PHONE", AssetType: "SCREENSHOT", URL: "u", Width: 10, Height: 10, SHA256: "zzz"}},
		{"empty app", domain.MediaAsset{Locale: "ko-KR", DeviceFormFactor: "PHONE", AssetType: "SCREENSHOT", URL: "u", Width: 10, Height: 10, SHA256: sha}},
	} {
		if _, err := svc.SaveMediaAsset(ctx, tc.asset); err == nil {
			t.Fatalf("%s: expected validation error, got nil", tc.name)
		}
	}
}

func TestDeleteMediaAssetRoundtrip(t *testing.T) {
	svc, _, _ := setupCatalog(t)
	ctx := context.Background()
	sha := strings.Repeat("c", 64)

	saved, err := svc.SaveMediaAsset(ctx, domain.MediaAsset{
		AppID: "app-media-02", Locale: "en-US", DeviceFormFactor: "PHONE",
		AssetType: "FEATURE_BANNER", URL: "media/app-media-02/banner.png",
		Width: 1024, Height: 500, SHA256: sha,
	})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	list, err := svc.ListMediaAssets(ctx, "app-media-02", "en-US")
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 asset, got %d err=%v", len(list), err)
	}
	if err := svc.DeleteMediaAsset(ctx, "app-media-02", saved.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	list, err = svc.ListMediaAssets(ctx, "app-media-02", "en-US")
	if err != nil || len(list) != 0 {
		t.Fatalf("expected 0 assets after delete, got %d err=%v", len(list), err)
	}
	if err := svc.DeleteMediaAsset(ctx, "app-media-02", "missing"); err != domain.ErrMediaAssetNotFound {
		t.Fatalf("expected not-found, got %v", err)
	}
}
