package application_test

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/infrastructure"
)

// TC-FUNC-007: Staged Rollout monotonicity and hash distribution across 10,000 devices
func TestStagedRolloutMonotonicityAndDistribution(t *testing.T) {
	releaseID := "release-rollout-test-uuid-1"
	totalDevices := 10000

	var eligibleAt20 []string
	var eligibleAt50 []string

	for i := 1; i <= totalDevices; i++ {
		deviceID := fmt.Sprintf("android-device-id-%06d", i)

		if domain.IsDeviceEligibleForRollout(releaseID, deviceID, 20) {
			eligibleAt20 = append(eligibleAt20, deviceID)
		}
		if domain.IsDeviceEligibleForRollout(releaseID, deviceID, 50) {
			eligibleAt50 = append(eligibleAt50, deviceID)
		}
	}

	count20 := len(eligibleAt20)
	count50 := len(eligibleAt50)

	// Verify ~20% (±1.5% tolerance -> 1850 ~ 2150)
	if count20 < 1850 || count20 > 2150 {
		t.Fatalf("expected ~20%% devices eligible at 20%% rollout, got %d (%.2f%%)", count20, float64(count20)/100.0)
	}

	// Verify ~50% (±1.5% tolerance -> 4850 ~ 5150)
	if count50 < 4850 || count50 > 5150 {
		t.Fatalf("expected ~50%% devices eligible at 50%% rollout, got %d (%.2f%%)", count50, float64(count50)/100.0)
	}

	// MONOTONICITY GUARANTEE: Every device eligible at 20% MUST be eligible at 50%!
	set50 := make(map[string]bool)
	for _, dev := range eligibleAt50 {
		set50[dev] = true
	}

	for _, dev := range eligibleAt20 {
		if !set50[dev] {
			t.Fatalf("monotonicity violation: device %s eligible at 20%% but NOT at 50%%", dev)
		}
	}
}

// TC-FUNC-008: 12-state release state machine legal and illegal transitions
func TestReleaseStateMachineTransitions(t *testing.T) {
	ctx := context.Background()
	repo := application.NewInMemoryDistributionRepo()
	svc := application.NewArtifactService(repo, nil)

	releaseID := "release-state-machine-001"
	rel := &domain.AppRelease{
		ID:                releaseID,
		DistributionID:    "dist-1",
		AppID:             "app-1",
		ReleaseTrack:      domain.TrackProduction,
		VersionCode:       100,
		VersionName:       "1.0.0",
		ReleaseStatus:     domain.ReleaseDraft,
		RolloutPercentage: 100,
		CreatedBy:         "user-admin-1",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	_ = repo.CreateRelease(ctx, rel)

	// 1. ILLEGAL TRANSITION: Try jumping from DRAFT directly to PUBLISHED_FULL
	_, err := svc.TransitionReleaseStatus(ctx, releaseID, domain.ReleasePublishedFull)
	if err == nil {
		t.Fatalf("expected illegal transition from DRAFT to PUBLISHED_FULL to fail")
	}

	// 2. ILLEGAL TRANSITION: Try jumping from SECURITY_SCANNING directly to PUBLISHED_FULL
	rel.ReleaseStatus = domain.ReleaseSecurityScanning
	_ = repo.UpdateRelease(ctx, rel)
	_, err = svc.TransitionReleaseStatus(ctx, releaseID, domain.ReleasePublishedFull)
	if err == nil {
		t.Fatalf("expected illegal transition from SECURITY_SCANNING to PUBLISHED_FULL to fail")
	}

	// 3. LEGAL SUCCESSION: DRAFT -> ARTIFACT_VERIFYING -> SECURITY_SCANNING -> READY_FOR_SUBMISSION -> SUBMITTED_FOR_REVIEW -> IN_REVIEW -> APPROVED_PENDING_PUBLISH -> PUBLISHED_FULL
	rel.ReleaseStatus = domain.ReleaseDraft
	_ = repo.UpdateRelease(ctx, rel)

	legalPath := []domain.ReleaseStatus{
		domain.ReleaseArtifactVerifying,
		domain.ReleaseSecurityScanning,
		domain.ReleaseReadyForSubmission,
		domain.ReleaseSubmittedForReview,
		domain.ReleaseInReview,
		domain.ReleaseApprovedPendingPub,
		domain.ReleasePublishedFull,
	}

	for _, step := range legalPath {
		updated, err := svc.TransitionReleaseStatus(ctx, releaseID, step)
		if err != nil {
			t.Fatalf("expected legal transition to %s to succeed, got %v", step, err)
		}
		if updated.ReleaseStatus != step {
			t.Fatalf("expected status %s, got %s", step, updated.ReleaseStatus)
		}
	}
}

// Helper to create a test APK buffer with optional APK signing block and files
func createTestAPK(hasV2SigBlock bool, v1Only bool, manifestContent string, files map[string][]byte) ([]byte, *x509.Certificate, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Write AndroidManifest.xml
	if manifestContent != "" {
		mfWriter, err := zipWriter.Create("AndroidManifest.xml")
		if err != nil {
			return nil, nil, err
		}
		if _, err := mfWriter.Write([]byte(manifestContent)); err != nil {
			return nil, nil, err
		}
	}

	// Write extra files (e.g. lib/arm64-v8a/libfoo.so)
	for name, data := range files {
		fw, err := zipWriter.Create(name)
		if err != nil {
			return nil, nil, err
		}
		if _, err := fw.Write(data); err != nil {
			return nil, nil, err
		}
	}

	if v1Only {
		// Add META-INF/CERT.RSA
		fw, err := zipWriter.Create("META-INF/CERT.RSA")
		if err != nil {
			return nil, nil, err
		}
		_, _ = fw.Write([]byte("dummy v1 PKCS7 signature"))
	}

	if err := zipWriter.Close(); err != nil {
		return nil, nil, err
	}

	zipBytes := buf.Bytes()
	if !hasV2SigBlock {
		return zipBytes, nil, nil
	}

	// Generate a self-signed RSA test certificate (2048-bit)
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(123456789),
		Subject: pkix.Name{
			CommonName:   "Test Official Release",
			Organization: []string{"Acme Corp"},
			Country:      []string{"KR"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, nil, err
	}
	parsedCert, _ := x509.ParseCertificate(certDER)

	// Build an APK Signing Block containing v2 signature
	// Format:
	// - uint64: size_of_block (excluding this field)
	// - ID-value pairs:
	//     - uint64: pairLen
	//     - uint32: 0x7109871a (v2)
	//     - certDER bytes wrapped with sequence tag
	// - uint64: size_of_block (same)
	// - 16 bytes: "APK Sig Block 42"
	v2ValueBuf := new(bytes.Buffer)
	v2ValueBuf.Write(certDER)

	pairBuf := new(bytes.Buffer)
	pairLen := uint64(4 + v2ValueBuf.Len())
	_ = binary.Write(pairBuf, binary.LittleEndian, pairLen)
	_ = binary.Write(pairBuf, binary.LittleEndian, uint32(0x7109871a))
	pairBuf.Write(v2ValueBuf.Bytes())

	blockSize := uint64(pairBuf.Len() + 8 + 16)

	sigBlockBuf := new(bytes.Buffer)
	_ = binary.Write(sigBlockBuf, binary.LittleEndian, blockSize)
	sigBlockBuf.Write(pairBuf.Bytes())
	_ = binary.Write(sigBlockBuf, binary.LittleEndian, blockSize)
	sigBlockBuf.WriteString("APK Sig Block 42")

	// Inject signing block right before Central Directory
	// Find EOCD
	eocdIdx := bytes.LastIndex(zipBytes, []byte{0x50, 0x4b, 0x05, 0x06})
	if eocdIdx == -1 {
		return nil, nil, fmt.Errorf("EOCD not found")
	}
	cdOffset := binary.LittleEndian.Uint32(zipBytes[eocdIdx+16 : eocdIdx+20])

	newApk := new(bytes.Buffer)
	newApk.Write(zipBytes[:cdOffset])
	newApk.Write(sigBlockBuf.Bytes())

	// Adjust Central Directory offset in EOCD
	newCDOffset := cdOffset + uint32(sigBlockBuf.Len())
	newEOCD := make([]byte, len(zipBytes)-eocdIdx)
	copy(newEOCD, zipBytes[eocdIdx:])
	binary.LittleEndian.PutUint32(newEOCD[16:20], newCDOffset)

	newApk.Write(zipBytes[cdOffset:eocdIdx])
	newApk.Write(newEOCD)

	return newApk.Bytes(), parsedCert, nil
}

// TC-FUNC-009: APK AXML manifest parsing and v1-only APK rejection
func TestAPKParserAndV1Rejection(t *testing.T) {
	manifestXML := `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.example.superapp"
    android:versionCode="42"
    android:versionName="1.4.2">
    <uses-sdk android:minSdkVersion="26" android:targetSdkVersion="34" />
    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.CAMERA" />
</manifest>`

	files := map[string][]byte{
		"lib/arm64-v8a/libnative.so": []byte("ELF binary stub"),
	}

	// 1. V1-ONLY APK -> Must be rejected with ErrApkV2SignatureRequired
	v1ApkBytes, _, err := createTestAPK(false, true, manifestXML, files)
	if err != nil {
		t.Fatalf("failed to create v1 apk: %v", err)
	}

	_, err = infrastructure.VerifyAPKSignature(bytes.NewReader(v1ApkBytes), int64(len(v1ApkBytes)), nil)
	if err == nil {
		t.Fatalf("expected v1-only APK to be rejected")
	}
	if err != domain.ErrApkV2SignatureRequired {
		t.Fatalf("expected ErrApkV2SignatureRequired, got %v", err)
	}

	// 2. V2 SIGNED APK -> Must parse manifest and verify signature
	v2ApkBytes, cert, err := createTestAPK(true, false, manifestXML, files)
	if err != nil {
		t.Fatalf("failed to create v2 apk: %v", err)
	}

	parsedManifest, err := infrastructure.InspectAPK(bytes.NewReader(v2ApkBytes), int64(len(v2ApkBytes)), "")
	if err != nil {
		t.Fatalf("failed to inspect APK: %v", err)
	}

	if parsedManifest.PackageName != "com.example.superapp" {
		t.Fatalf("expected package name com.example.superapp, got %s", parsedManifest.PackageName)
	}
	if parsedManifest.VersionCode != 42 {
		t.Fatalf("expected versionCode 42, got %d", parsedManifest.VersionCode)
	}
	if parsedManifest.MinSDKVersion != 26 || parsedManifest.TargetSDKVersion != 34 {
		t.Fatalf("expected SDK versions 26/34, got %d/%d", parsedManifest.MinSDKVersion, parsedManifest.TargetSDKVersion)
	}
	if len(parsedManifest.Permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(parsedManifest.Permissions))
	}
	if len(parsedManifest.NativeABIs) != 1 || parsedManifest.NativeABIs[0] != "arm64-v8a" {
		t.Fatalf("expected arm64-v8a ABI, got %v", parsedManifest.NativeABIs)
	}

	verifiedSig, err := infrastructure.VerifyAPKSignature(bytes.NewReader(v2ApkBytes), int64(len(v2ApkBytes)), nil)
	if err != nil {
		t.Fatalf("failed to verify v2 signature: %v", err)
	}
	if verifiedSig.SignatureSchemeVersion != 2 {
		t.Fatalf("expected v2 signature scheme, got %d", verifiedSig.SignatureSchemeVersion)
	}
	expectedFP := fmt.Sprintf("%x", sha256.Sum256(cert.Raw))
	if verifiedSig.CertSHA256Fingerprint != expectedFP {
		t.Fatalf("fingerprint mismatch: expected %s, got %s", expectedFP, verifiedSig.CertSHA256Fingerprint)
	}
}

// TC-SEC-007: Signature mismatch block against previous release
func TestSignatureMismatchRejection(t *testing.T) {
	ctx := context.Background()
	repo := application.NewInMemoryDistributionRepo()
	svc := application.NewArtifactService(repo, nil)

	manifestXML := `<manifest package="com.acme.app" versionCode="2" versionName="1.0.1"></manifest>`

	// Create Release 1 with Key A
	apkBytesA, certA, err := createTestAPK(true, false, manifestXML, nil)
	if err != nil {
		t.Fatalf("failed to create apk A: %v", err)
	}
	shaA := fmt.Sprintf("%x", sha256.Sum256(apkBytesA))

	releaseID1 := "rel-001"
	rel1 := &domain.AppRelease{
		ID:             releaseID1,
		DistributionID: "dist-001",
		AppID:          "app-001",
		ReleaseStatus:  domain.ReleaseDraft,
	}
	_ = repo.CreateRelease(ctx, rel1)

	session1, _ := svc.InitiateUploadSession(ctx, "org-1", "app-1", domain.ArtifactTypeAPK, "app-v1.apk", int64(len(apkBytesA)), shaA)
	_, _, err = svc.ProcessUploadedAPK(ctx, session1.ID, releaseID1, bytes.NewReader(apkBytesA), int64(len(apkBytesA)))
	if err != nil {
		t.Fatalf("failed to process initial APK: %v", err)
	}

	// Create Release 2 with Key B (different signing key!)
	apkBytesB, _, err := createTestAPK(true, false, manifestXML, nil)
	if err != nil {
		t.Fatalf("failed to create apk B: %v", err)
	}
	shaB := fmt.Sprintf("%x", sha256.Sum256(apkBytesB))

	releaseID2 := "rel-002"
	rel2 := &domain.AppRelease{
		ID:             releaseID2,
		DistributionID: "dist-001", // same distribution!
		AppID:          "app-001",
		ReleaseStatus:  domain.ReleaseDraft,
	}
	_ = repo.CreateRelease(ctx, rel2)

	session2, _ := svc.InitiateUploadSession(ctx, "org-1", "app-1", domain.ArtifactTypeAPK, "app-v2-mismatch.apk", int64(len(apkBytesB)), shaB)
	_, _, err = svc.ProcessUploadedAPK(ctx, session2.ID, releaseID2, bytes.NewReader(apkBytesB), int64(len(apkBytesB)))

	// Must fail with ErrSignatureMismatch!
	if err == nil {
		t.Fatalf("expected signature mismatch rejection, but succeeded")
	}
	if err != domain.ErrSignatureMismatch {
		t.Fatalf("expected ErrSignatureMismatch, got %v", err)
	}

	_ = certA
}

// Zip Bomb and Zip Slip security validation tests
func TestZipBombAndZipSlipRejection(t *testing.T) {
	// 1. ZIP BOMB TEST: Compression ratio > 100:1
	bombBuf := new(bytes.Buffer)
	zw := zip.NewWriter(bombBuf)
	zw.RegisterCompressor(zip.Deflate, nil)
	fw, err := zw.Create("huge_zeroes.bin")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	zeroes := make([]byte, 2*1024*1024) // 2MB
	_, _ = fw.Write(zeroes)
	_ = zw.Close()

	bombBytes := bombBuf.Bytes()
	_, err = infrastructure.InspectAPK(bytes.NewReader(bombBytes), int64(len(bombBytes)), "")
	if err == nil {
		t.Fatalf("expected zip bomb to be detected and rejected")
	}
	if err != domain.ErrZipBombDetected {
		t.Fatalf("expected ErrZipBombDetected, got %v", err)
	}

	// 2. ZIP SLIP TEST: Relative path '../malicious.so'
	slipBuf := new(bytes.Buffer)
	zwSlip := zip.NewWriter(slipBuf)
	_, _ = zwSlip.Create("../etc/malicious.so")
	_ = zwSlip.Close()
	slipBytes := slipBuf.Bytes()
	_, err = infrastructure.InspectAPK(bytes.NewReader(slipBytes), int64(len(slipBytes)), "")
	if err == nil {
		t.Fatalf("expected zip slip to be detected and rejected")
	}
	if err != domain.ErrZipSlipDetected {
		t.Fatalf("expected ErrZipSlipDetected, got %v", err)
	}
}
