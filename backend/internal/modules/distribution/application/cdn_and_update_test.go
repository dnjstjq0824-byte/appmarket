package application

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
)

// TC-FUNC-012: 4-Dimensional Device Compatibility & Signature Continuity Resolution
func TestTCFUNC012DeviceCompatibilityAndSignatureContinuity(t *testing.T) {
	deviceArm64 := domain.DeviceSpec{
		SDKInt:            34,
		SupportedABIs:     []string{"arm64-v8a", "armeabi-v7a"},
		DensityDPI:        440,
		AvailableFeatures: []string{"android.hardware.camera", "android.hardware.nfc"},
		DeviceHash:        "device-hash-12345",
	}

	installedApp := domain.InstalledAppSpec{
		PackageName:          "com.example.bank",
		VersionCode:          100,
		CertSHA256:           "AAAA1111222233334444555566667777888899990000AAAABBBBCCCCDDDDEEEEFF",
		InstallerPackageName: "com.bookingnetwork.appmarket",
	}

	baseCandidate := domain.CandidateRelease{
		AppID:              "bank-app-1",
		ReleaseID:          "rel-200",
		PackageName:        "com.example.bank",
		VersionCode:        200,
		VersionName:        "2.0.0",
		MinSDK:             26,
		TargetSDK:          34,
		NativeABIs:         []string{"arm64-v8a"},
		RequiredFeatures:   []string{"android.hardware.nfc"},
		CertSHA256:         "AAAA1111222233334444555566667777888899990000AAAABBBBCCCCDDDDEEEEFF", // Same signer
		RotatedCertLineage: []string{},
		RolloutPercentage:  100,
		ReleaseStatus:      domain.ReleasePublishedFull,
		DownloadURL:        "https://cdn.appmarket.local/packages/bank/rel-200/abc.apk",
		SHA256Checksum:     "deadbeef12345678deadbeef12345678deadbeef12345678deadbeef12345678",
		SizeBytes:          25000000,
	}

	t.Run("Scenario 1: Full Compatibility with Same Signer -> UPDATE_AVAILABLE_COMPATIBLE", func(t *testing.T) {
		status, res := domain.ResolveUpdate(deviceArm64, installedApp, baseCandidate)
		if status != domain.StatusUpdateAvailableCompatible {
			t.Fatalf("expected StatusUpdateAvailableCompatible, got %s", status)
		}
		if res.RequiresUserAction {
			t.Fatalf("expected RequiresUserAction=false for same signer")
		}
		if res.DownloadURL == "" {
			t.Fatalf("expected valid download URL")
		}
	})

	t.Run("Scenario 2: Key Rotation with v3 Proof-of-Rotation Lineage -> COMPATIBLE_KEY_ROTATION", func(t *testing.T) {
		candKeyRotation := baseCandidate
		candKeyRotation.CertSHA256 = "BBBB9999888877776666555544443333222211110000AAAABBBBCCCCDDDDEEEEFF" // New rotated key
		candKeyRotation.RotatedCertLineage = []string{
			installedApp.CertSHA256, // Old key in lineage
		}

		status, res := domain.ResolveUpdate(deviceArm64, installedApp, candKeyRotation)
		if status != domain.StatusUpdateAvailableKeyRotation {
			t.Fatalf("expected StatusUpdateAvailableKeyRotation, got %s", status)
		}
		if res.RequiresUserAction {
			t.Fatalf("expected seamless update without user action for valid rotation")
		}
	})

	t.Run("Scenario 3: Signer Mismatch without Rotation -> UPDATE_AVAILABLE_SIGNER_CONFLICT", func(t *testing.T) {
		candSignerConflict := baseCandidate
		candSignerConflict.CertSHA256 = "CCCC9999888877776666555544443333222211110000AAAABBBBCCCCDDDDEEEEFF" // Different key
		candSignerConflict.RotatedCertLineage = []string{} // No lineage connecting them

		status, res := domain.ResolveUpdate(deviceArm64, installedApp, candSignerConflict)
		if status != domain.StatusUpdateAvailableSignerConflict {
			t.Fatalf("expected StatusUpdateAvailableSignerConflict, got %s", status)
		}
		if !res.RequiresUserAction {
			t.Fatalf("expected RequiresUserAction=true when signer conflict requires reinstallation")
		}
	})

	t.Run("Scenario 4: Incompatible SDK -> INCOMPATIBLE_DEVICE_SDK_ABI", func(t *testing.T) {
		candHighSDK := baseCandidate
		candHighSDK.MinSDK = 35 // Device is 34

		status, res := domain.ResolveUpdate(deviceArm64, installedApp, candHighSDK)
		if status != domain.StatusIncompatibleDeviceSdkAbi {
			t.Fatalf("expected StatusIncompatibleDeviceSdkAbi, got %s", status)
		}
		if !strings.Contains(res.ResolutionReason, "SDK") {
			t.Fatalf("expected reason to mention SDK, got %s", res.ResolutionReason)
		}
	})

	t.Run("Scenario 5: Incompatible ABI -> INCOMPATIBLE_DEVICE_SDK_ABI", func(t *testing.T) {
		candX86Only := baseCandidate
		candX86Only.NativeABIs = []string{"x86_64"} // Device is arm64

		status, res := domain.ResolveUpdate(deviceArm64, installedApp, candX86Only)
		if status != domain.StatusIncompatibleDeviceSdkAbi {
			t.Fatalf("expected StatusIncompatibleDeviceSdkAbi, got %s", status)
		}
		if !strings.Contains(res.ResolutionReason, "architecture") && !strings.Contains(res.ResolutionReason, "ABI") {
			t.Fatalf("expected reason to mention ABI/architecture, got %s", res.ResolutionReason)
		}
	})

	t.Run("Scenario 6: Missing Hardware Feature -> INCOMPATIBLE_DEVICE_SDK_ABI", func(t *testing.T) {
		candSensor := baseCandidate
		candSensor.RequiredFeatures = []string{"android.hardware.sensor.barometer"} // Device lacks barometer

		status, res := domain.ResolveUpdate(deviceArm64, installedApp, candSensor)
		if status != domain.StatusIncompatibleDeviceSdkAbi {
			t.Fatalf("expected StatusIncompatibleDeviceSdkAbi, got %s", status)
		}
		if !strings.Contains(res.ResolutionReason, "barometer") {
			t.Fatalf("expected reason to mention missing feature, got %s", res.ResolutionReason)
		}
	})

	t.Run("Scenario 7: Batch Multi-App Resolution", func(t *testing.T) {
		installedList := []domain.InstalledAppSpec{
			installedApp,
			{
				PackageName: "com.example.nonexistent",
				VersionCode: 1,
			},
		}
		candidates := map[string]domain.CandidateRelease{
			installedApp.PackageName: baseCandidate,
		}

		results := domain.BatchResolveUpdates(deviceArm64, installedList, candidates)
		if len(results) != 2 {
			t.Fatalf("expected 2 batch results, got %d", len(results))
		}
		if results[0].Status != domain.StatusUpdateAvailableCompatible {
			t.Fatalf("expected first result to be compatible update, got %s", results[0].Status)
		}
		if results[1].Status != domain.StatusNoUpdate {
			t.Fatalf("expected second result to be NO_UPDATE, got %s", results[1].Status)
		}
	})
}

// TC-OPS-003: Immutable CDN Path Overwrite Rejection & Expired Signed URL
func TestTCOPS003ImmutableCDNAndSignedURLExpiry(t *testing.T) {
	ctx := context.Background()
	storage := NewInMemoryObjectStorage()
	quarantineBucket := "bn-quarantine-bucket"
	publicBucket := "bn-public-cdn-bucket"
	cdnDomain := "cdn.appmarket.bookingnetwork.internal"

	publisher := NewCDNPublisher(storage, quarantineBucket, publicBucket, cdnDomain)
	tokenService := NewDownloadTokenService("test-secret-key-32-bytes-long!!", cdnDomain, 15*time.Minute)

	// Setup verified artifact in quarantine bucket
	artifactBytes := []byte("VALID_APK_PAYLOAD_CONTENT_FOR_RELEASE_100")
	shaHex := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	// Actually compute real sha256
	h := domain.ComputeSHA256(artifactBytes)
	shaHex = h

	quarantineKey := "quarantine/org-1/app-1/upload-100.tmp"
	_ = storage.PutObject(ctx, quarantineBucket, &StorageObject{
		Key:     quarantineKey,
		Content: artifactBytes,
		SHA256:  shaHex,
	})

	artifact := &domain.ReleaseArtifact{
		ID:              "art-100",
		ReleaseID:       "rel-100",
		PackageName:     "com.example.app",
		VersionCode:     100,
		SHA256Hash:      shaHex,
		S3QuarantineKey: quarantineKey,
	}

	t.Run("Sub-test 1: Successful Promotion to Immutable Public CDN", func(t *testing.T) {
		cdnURL, err := publisher.PromoteToPublicCDN(ctx, artifact)
		if err != nil {
			t.Fatalf("expected successful promotion, got %v", err)
		}
		expectedKey := FormatCDNKey(artifact.PackageName, artifact.ReleaseID, artifact.SHA256Hash)
		expectedURL := fmt.Sprintf("https://%s/%s", cdnDomain, expectedKey)
		if cdnURL != expectedURL {
			t.Fatalf("expected %s, got %s", expectedURL, cdnURL)
		}

		// Verify object exists in public bucket with immutable cache-control
		obj, err := storage.GetObject(ctx, publicBucket, expectedKey)
		if err != nil {
			t.Fatalf("expected object in public bucket: %v", err)
		}
		if obj.CacheControl != "public, max-age=31536000, immutable" {
			t.Fatalf("expected immutable cache control, got %s", obj.CacheControl)
		}
	})

	t.Run("Sub-test 2: Overwrite Rejection on Content Mismatch (Immutable Integrity)", func(t *testing.T) {
		// Attempt to overwrite same CDN key with different binary
		conflictingQuarantineKey := "quarantine/org-1/app-1/tampered-100.tmp"
		differentBytes := []byte("TAMPERED_OR_DIFFERENT_APK_BINARY")
		diffSHA := domain.ComputeSHA256(differentBytes)

		_ = storage.PutObject(ctx, quarantineBucket, &StorageObject{
			Key:     conflictingQuarantineKey,
			Content: differentBytes,
			SHA256:  diffSHA,
		})

		tamperedArtifact := &domain.ReleaseArtifact{
			ID:              "art-tampered",
			ReleaseID:       "rel-100", // Same release
			PackageName:     "com.example.app",
			VersionCode:     100,
			SHA256Hash:      diffSHA,
			S3QuarantineKey: conflictingQuarantineKey,
		}

		// Inject conflicting existing object with same key but different SHA
		conflictKey := FormatCDNKey(tamperedArtifact.PackageName, tamperedArtifact.ReleaseID, tamperedArtifact.SHA256Hash)
		_ = storage.PutObject(ctx, publicBucket, &StorageObject{
			Key:    conflictKey,
			SHA256: "DIFFERENT_HASH_ALREADY_STORED",
		})

		_, err := publisher.PromoteToPublicCDN(ctx, tamperedArtifact)
		if err != ErrCDNImmutablePathConflict {
			t.Fatalf("expected ErrCDNImmutablePathConflict, got %v", err)
		}
	})

	t.Run("Sub-test 3: Generate 15-Minute Signed Download URL", func(t *testing.T) {
		req := TokenRequest{
			AppID:          "com.example.app",
			ReleaseID:      "rel-100",
			SHA256Checksum: shaHex,
			DeviceOrUserID: "dev-hash-abc",
			CountryCode:    "KR",
			RightsLevel:    "TIER_1_DIRECT_MIRROR",
			ReleaseStatus:  domain.ReleasePublishedFull,
		}

		signedURL, err := tokenService.GenerateSignedURL(req)
		if err != nil {
			t.Fatalf("expected signed URL generation success, got %v", err)
		}

		// Validate token
		valid, err := tokenService.ValidateSignedURL(ctx, signedURL.URL)
		if err != nil || !valid {
			t.Fatalf("expected token validation to pass, got err=%v", err)
		}
	})

	t.Run("Sub-test 4: Expired Signed URL Rejection", func(t *testing.T) {
		// Token service with negative TTL (already expired)
		expiredService := NewDownloadTokenService("test-secret-key-32-bytes-long!!", cdnDomain, -1*time.Minute)
		req := TokenRequest{
			AppID:          "com.example.app",
			ReleaseID:      "rel-100",
			SHA256Checksum: shaHex,
			DeviceOrUserID: "dev-hash-abc",
			CountryCode:    "KR",
			RightsLevel:    "TIER_1_DIRECT_MIRROR",
			ReleaseStatus:  domain.ReleasePublishedFull,
		}

		signedURL, err := expiredService.GenerateSignedURL(req)
		if err != nil {
			t.Fatalf("generation failed: %v", err)
		}

		valid, err := expiredService.ValidateSignedURL(ctx, signedURL.URL)
		if valid || err != ErrDownloadTokenExpired {
			t.Fatalf("expected ErrDownloadTokenExpired, got valid=%v, err=%v", valid, err)
		}
	})

	t.Run("Sub-test 5: Kill Switch Revocation Rejection", func(t *testing.T) {
		req := TokenRequest{
			AppID:          "com.example.app",
			ReleaseID:      "rel-revoked",
			SHA256Checksum: shaHex,
			DeviceOrUserID: "dev-hash-abc",
			CountryCode:    "KR",
			RightsLevel:    "TIER_1_DIRECT_MIRROR",
			ReleaseStatus:  domain.ReleasePublishedFull,
		}

		signedURL, err := tokenService.GenerateSignedURL(req)
		if err != nil {
			t.Fatalf("expected URL generation to succeed before revoke: %v", err)
		}

		// Trigger Kill Switch Revocation
		tokenService.RevokeRelease("rel-revoked")

		// Subsequent validation must fail
		valid, err := tokenService.ValidateSignedURL(ctx, signedURL.URL)
		if valid || err != ErrDownloadTokenRevoked {
			t.Fatalf("expected ErrDownloadTokenRevoked, got valid=%v, err=%v", valid, err)
		}

		// Subsequent generation must also fail
		_, err = tokenService.GenerateSignedURL(req)
		if err != ErrDownloadTokenRevoked {
			t.Fatalf("expected ErrDownloadTokenRevoked on generation, got %v", err)
		}
	})

	t.Run("Sub-test 6: EXTERNAL_LINK_ONLY Apps Cannot Generate Download Token", func(t *testing.T) {
		req := TokenRequest{
			AppID:          "com.example.external",
			ReleaseID:      "rel-ext",
			SHA256Checksum: shaHex,
			DeviceOrUserID: "dev-hash-abc",
			CountryCode:    "KR",
			RightsLevel:    "TIER_5_EXTERNAL_LINK_ONLY",
			ReleaseStatus:  domain.ReleasePublishedFull,
		}

		_, err := tokenService.GenerateSignedURL(req)
		if err != ErrExternalLinkOnlyDownloadForbidden {
			t.Fatalf("expected ErrExternalLinkOnlyDownloadForbidden, got %v", err)
		}
	})
}
