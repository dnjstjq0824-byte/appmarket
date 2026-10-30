package connector

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/connector/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/connector/fdroid"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/connector/partner"
)

// TC-FUNC-004: 5-Tier Data Provenance & Rights Guard Enforcement
func TestTCFUNC004ProvenanceAndRightsGuard(t *testing.T) {
	guard := domain.NewRightsGuard()

	t.Run("Sub-test 1: Direct Download permissions by Rights Level", func(t *testing.T) {
		// Level 1 & 2 must allow download
		if err := guard.ValidateDownloadAllowed(domain.Level1DirectHosted); err != nil {
			t.Fatalf("expected Level 1 to allow direct download, got %v", err)
		}
		if err := guard.ValidateDownloadAllowed(domain.Level2LicensedFederated); err != nil {
			t.Fatalf("expected Level 2 to allow direct download, got %v", err)
		}

		// Level 3, 4, 5 must strictly forbid direct download
		levels := []domain.RightsLevel{
			domain.Level3SearchAndInfoOnly,
			domain.Level4OfficialLinkOnly,
			domain.Level5RestrictedProhibited,
		}

		for _, lvl := range levels {
			err := guard.ValidateDownloadAllowed(lvl)
			if err != domain.ErrDownloadForbiddenRightsLevel {
				t.Fatalf("expected ErrDownloadForbiddenRightsLevel for %s, got %v", lvl, err)
			}
		}
	})

	t.Run("Sub-test 2: Mandatory Provenance 4 Fields formatting", func(t *testing.T) {
		now := time.Now().UTC()
		meta, err := guard.FormatProvenance("Google Play Store", "https://play.google.com/store/apps/details?id=com.example", now, domain.Level4OfficialLinkOnly)
		if err != nil {
			t.Fatalf("expected format success, got %v", err)
		}

		if meta.ProvenanceLabel != "Google Play Store" {
			t.Fatalf("expected Google Play Store, got %s", meta.ProvenanceLabel)
		}
		if meta.CanDirectInstall != false {
			t.Fatalf("expected CanDirectInstall=false for Level 4")
		}
		if meta.ActionText != "공식 배포처로 이동 ↗" {
			t.Fatalf("expected official link text, got %s", meta.ActionText)
		}
	})

	t.Run("Sub-test 3: Partner Feed URL Domain Validation", func(t *testing.T) {
		pSyncer := partner.NewPartnerSyncer()

		// Valid official Google Play URL
		meta, err := pSyncer.IngestPartnerApp("Google Play", "https://play.google.com/store/apps/details?id=com.bank.app", "com.bank.app")
		if err != nil || meta.CanDirectInstall {
			t.Fatalf("expected valid ingestion with CanDirectInstall=false, got meta=%v, err=%v", meta, err)
		}

		// Invalid unapproved domain
		_, err = pSyncer.IngestPartnerApp("Rogue Store", "https://malicious-apk-host.com/download.apk", "com.bank.app")
		if err != partner.ErrUnapprovedStoreDomain {
			t.Fatalf("expected ErrUnapprovedStoreDomain, got %v", err)
		}
	})
}

// TC-SEC-003: External Index Signature Validation & Tamper Rejection
func TestTCSEC003ExternalIndexSignatureTamperRejection(t *testing.T) {
	pinnedFP := "43238D512C1E5EB2D6569F4A3AFBF5523418B82E0A3ED1552770ABB9A9C1CC25"
	syncer := fdroid.NewFDroidSyncer(pinnedFP)

	validPayload := []byte(`{
		"repo": {
			"name": "F-Droid",
			"url": "https://f-droid.org/repo",
			"fingerprint": "43238D512C1E5EB2D6569F4A3AFBF5523418B82E0A3ED1552770ABB9A9C1CC25",
			"timestamp": 1730000000
		},
		"packages": {
			"org.fdroid.fdroid": {
				"packageName": "org.fdroid.fdroid",
				"versionCode": 1020000,
				"versionName": "1.20.0",
				"minSdkVersion": 21,
				"apkName": "org.fdroid.fdroid_1020000.apk",
				"sha256": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				"size": 10500000,
				"signer": "43238D512C1E5EB2D6569F4A3AFBF5523418B82E0A3ED1552770ABB9A9C1CC25"
			}
		}
	}`)

	hasher := sha256.New()
	hasher.Write(validPayload)
	validDigest := hex.EncodeToString(hasher.Sum(nil))

	t.Run("Sub-test 1: Valid feed and signature matches successfully", func(t *testing.T) {
		results, err := syncer.SyncFeed(validPayload, validDigest)
		if err != nil {
			t.Fatalf("expected feed sync success, got %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].RightsLevel != domain.Level2LicensedFederated {
			t.Fatalf("expected Level 2 rights level, got %s", results[0].RightsLevel)
		}
	})

	t.Run("Sub-test 2: Tampered feed payload is rejected with ABORTED_SIGNATURE_INVALID", func(t *testing.T) {
		tamperedPayload := []byte(`{"repo": {"name": "Tampered F-Droid"}}`)

		// Attempt sync with old digest against tampered payload
		_, err := syncer.SyncFeed(tamperedPayload, validDigest)
		if err == nil || !errors.Is(err, domain.ErrAbortedSignatureInvalid) {
			t.Fatalf("expected ErrAbortedSignatureInvalid, got %v", err)
		}
	})

	t.Run("Sub-test 3: Mismatched repo certificate fingerprint is rejected", func(t *testing.T) {
		spoofedPayload := []byte(`{
			"repo": {
				"name": "Spoofed Repo",
				"url": "https://spoofed.org/repo",
				"fingerprint": "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
				"timestamp": 1730000000
			},
			"packages": {}
		}`)

		hasher := sha256.New()
		hasher.Write(spoofedPayload)
		spoofedDigest := hex.EncodeToString(hasher.Sum(nil))

		_, err := syncer.SyncFeed(spoofedPayload, spoofedDigest)
		if err == nil || !errors.Is(err, domain.ErrAbortedSignatureInvalid) {
			t.Fatalf("expected ErrAbortedSignatureInvalid on spoofed cert, got %v", err)
		}
	})
}

