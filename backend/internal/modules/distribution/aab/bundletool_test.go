package aab

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"
)

// TC-FUNC-009: AAB bundletool APK-Set Conversion and Universal APK Extraction
func TestTCFUNC009BundletoolUniversalAPK(t *testing.T) {
	ctx := context.Background()
	vault := NewInMemorySigningVault()
	signingKey, err := vault.CreateAppSigningKey(ctx, "com.example.universal")
	if err != nil {
		t.Fatalf("failed to create signing key: %v", err)
	}

	executor := NewSystemBundletoolExecutor("/opt/bundletool/bundletool.jar")

	t.Run("Sub-test 1: Valid AAB structure generates Universal APK with valid SHA-256", func(t *testing.T) {
		// Construct minimal valid AAB zip archive containing base/manifest/AndroidManifest.xml
		var aabBuf bytes.Buffer
		zw := zip.NewWriter(&aabBuf)

		mw, err := zw.Create("base/manifest/AndroidManifest.xml")
		if err != nil {
			t.Fatalf("failed to create zip entry: %v", err)
		}
		_, _ = mw.Write([]byte("<manifest package=\"com.example.universal\" versionCode=\"1\"/>"))

		dw, err := zw.Create("base/dex/classes.dex")
		if err != nil {
			t.Fatalf("failed to create dex entry: %v", err)
		}
		_, _ = dw.Write([]byte("dex\n035\x00data"))

		_ = zw.Close()

		res, err := executor.BuildUniversalAPK(ctx, aabBuf.Bytes(), signingKey)
		if err != nil {
			t.Fatalf("expected BuildUniversalAPK success, got %v", err)
		}

		if res.PackageName != "com.example.universal" {
			t.Fatalf("expected com.example.universal, got %s", res.PackageName)
		}
		if res.VersionCode != 1 {
			t.Fatalf("expected versionCode 1, got %d", res.VersionCode)
		}
		if len(res.SHA256Checksum) != 64 {
			t.Fatalf("expected 64-char sha256 checksum, got %s", res.SHA256Checksum)
		}
		if len(res.APKContent) == 0 {
			t.Fatalf("expected non-empty universal apk content")
		}
	})

	t.Run("Sub-test 2: Corrupted or Non-AAB Archive is Rejected with REJECTED_INVALID_APK", func(t *testing.T) {
		corruptedData := []byte("NOT_A_ZIP_OR_AAB_ARCHIVE_DATA")

		_, err := executor.BuildUniversalAPK(ctx, corruptedData, signingKey)
		if err == nil {
			t.Fatalf("expected error for corrupted archive, got nil")
		}

		// AAB without manifest
		var emptyZip bytes.Buffer
		zw := zip.NewWriter(&emptyZip)
		f, _ := zw.Create("random_file.txt")
		_, _ = f.Write([]byte("random"))
		_ = zw.Close()

		_, err = executor.BuildUniversalAPK(ctx, emptyZip.Bytes(), signingKey)
		if err == nil {
			t.Fatalf("expected error for zip missing AndroidManifest, got nil")
		}
	})
}
