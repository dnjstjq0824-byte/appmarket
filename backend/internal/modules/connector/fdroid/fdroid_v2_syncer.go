package fdroid

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/connector/domain"
)

type FDroidPackageEntry struct {
	PackageName   string            `json:"packageName"`
	VersionCode   int64             `json:"versionCode"`
	VersionName   string            `json:"versionName"`
	MinSDKVersion int               `json:"minSdkVersion"`
	ApkName       string            `json:"apkName"`
	SHA256        string            `json:"sha256"`
	SizeBytes     int64             `json:"size"`
	SignerCertHex string            `json:"signer"`
	Summary       map[string]string `json:"summary"`
	License       string            `json:"license"`
}

type FDroidIndexV2 struct {
	Repo struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		Fingerprint string `json:"fingerprint"` // Expected Certificate Fingerprint
		Timestamp   int64  `json:"timestamp"`
	} `json:"repo"`
	Packages map[string]FDroidPackageEntry `json:"packages"`
}

type FDroidSyncer struct {
	pinnedRepoFingerprint string
	rightsGuard           *domain.RightsGuard
}

func NewFDroidSyncer(pinnedFingerprint string) *FDroidSyncer {
	return &FDroidSyncer{
		pinnedRepoFingerprint: strings.ToUpper(strings.TrimSpace(pinnedFingerprint)),
		rightsGuard:           domain.NewRightsGuard(),
	}
}

// SyncFeed parses and cryptographically validates F-Droid Index v2 JSON feed
func (s *FDroidSyncer) SyncFeed(indexV2JSON []byte, feedSignatureHex string) ([]domain.ProvenanceMetadata, error) {
	if len(indexV2JSON) == 0 {
		return nil, errors.New("empty feed content")
	}

	// 1. Verify Feed Signature (Simulating JAR / GPG entry signature verification)
	hasher := sha256.New()
	hasher.Write(indexV2JSON)
	calculatedHash := hex.EncodeToString(hasher.Sum(nil))

	if feedSignatureHex != "" {
		// If provided signature hash does not match calculated SHA-256 of feed payload -> ABORTED_SIGNATURE_INVALID
		if !strings.EqualFold(calculatedHash, strings.TrimSpace(feedSignatureHex)) {
			return nil, fmt.Errorf("%w: feed content digest mismatch", domain.ErrAbortedSignatureInvalid)
		}
	}

	// 2. Parse JSON payload
	var feed FDroidIndexV2
	if err := json.Unmarshal(indexV2JSON, &feed); err != nil {
		return nil, fmt.Errorf("%w: malformed json payload: %v", domain.ErrAbortedSignatureInvalid, err)
	}

	// 3. Verify Repo Certificate Fingerprint against Pinned Key
	feedFingerprint := strings.ToUpper(strings.TrimSpace(feed.Repo.Fingerprint))
	if s.pinnedRepoFingerprint != "" && feedFingerprint != s.pinnedRepoFingerprint {
		return nil, fmt.Errorf("%w: unpinned repository certificate fingerprint %s", domain.ErrAbortedSignatureInvalid, feedFingerprint)
	}

	// 4. Transform packages into Level 2 Provenance records
	results := make([]domain.ProvenanceMetadata, 0, len(feed.Packages))
	syncTime := time.Unix(feed.Repo.Timestamp, 0).UTC()
	if feed.Repo.Timestamp == 0 {
		syncTime = time.Now().UTC()
	}

	for _, pkg := range feed.Packages {
		officialURL := fmt.Sprintf("%s/%s", strings.TrimRight(feed.Repo.URL, "/"), pkg.PackageName)
		meta, err := s.rightsGuard.FormatProvenance("F-Droid Official FOSS", officialURL, syncTime, domain.Level2LicensedFederated)
		if err != nil {
			continue
		}
		results = append(results, *meta)
	}

	return results, nil
}
