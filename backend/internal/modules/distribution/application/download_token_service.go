package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
)

var (
	ErrDownloadTokenExpired                = errors.New("ERR_DOWNLOAD_TOKEN_EXPIRED")
	ErrDownloadTokenInvalidSignature       = errors.New("ERR_DOWNLOAD_TOKEN_INVALID_SIGNATURE")
	ErrDownloadTokenRevoked                = errors.New("ERR_DOWNLOAD_TOKEN_REVOKED_BY_KILL_SWITCH")
	ErrExternalLinkOnlyDownloadForbidden   = errors.New("ERR_EXTERNAL_LINK_ONLY_CANNOT_GENERATE_TOKEN")
	ErrCountryGeoBlocked                   = errors.New("ERR_COUNTRY_GEO_BLOCKED")
	ErrReleaseNotPublished                 = errors.New("ERR_RELEASE_NOT_PUBLISHED")
)

type DownloadTokenService struct {
	hmacSecret        []byte
	cdnDomain         string
	tokenTTL          time.Duration
	revokedReleases   map[string]bool
	mu                sync.RWMutex
}

func NewDownloadTokenService(secretKey string, cdnDomain string, tokenTTL time.Duration) *DownloadTokenService {
	if tokenTTL == 0 {
		tokenTTL = 15 * time.Minute // 15-minute standard expiry
	}
	return &DownloadTokenService{
		hmacSecret:      []byte(secretKey),
		cdnDomain:       cdnDomain,
		tokenTTL:        tokenTTL,
		revokedReleases: make(map[string]bool),
	}
}

type TokenRequest struct {
	AppID           string
	ReleaseID       string
	SHA256Checksum  string
	DeviceOrUserID  string
	CountryCode     string
	BlockedCountries []string
	RightsLevel     string // e.g. "TIER_1_DIRECT_MIRROR", "TIER_5_EXTERNAL_LINK_ONLY"
	ReleaseStatus   domain.ReleaseStatus
}

type SignedDownloadURL struct {
	URL            string    `json:"url"`
	ExpiresAt      time.Time `json:"expires_at"`
	SHA256Checksum string    `json:"sha256_checksum"`
}

// GenerateSignedURL creates a 15-minute expiring HMAC-SHA256 signed URL for CDN access
func (s *DownloadTokenService) GenerateSignedURL(req TokenRequest) (*SignedDownloadURL, error) {
	// 1. External link only cannot generate download token
	if strings.EqualFold(req.RightsLevel, "EXTERNAL_LINK_ONLY") || strings.EqualFold(req.RightsLevel, "TIER_5_EXTERNAL_LINK_ONLY") {
		return nil, ErrExternalLinkOnlyDownloadForbidden
	}

	// 2. Status verification
	if req.ReleaseStatus != domain.ReleasePublishedFull && req.ReleaseStatus != domain.ReleasePublishedStaged {
		return nil, ErrReleaseNotPublished
	}

	// 3. Kill Switch check
	s.mu.RLock()
	if s.revokedReleases[req.ReleaseID] || req.ReleaseStatus == domain.ReleaseRevokedByAdmin {
		s.mu.RUnlock()
		return nil, ErrDownloadTokenRevoked
	}
	s.mu.RUnlock()

	// 4. Country check
	reqCountry := strings.ToUpper(strings.TrimSpace(req.CountryCode))
	for _, blocked := range req.BlockedCountries {
		if strings.ToUpper(strings.TrimSpace(blocked)) == reqCountry {
			return nil, ErrCountryGeoBlocked
		}
	}

	expiresAt := time.Now().UTC().Add(s.tokenTTL)
	expUnix := expiresAt.Unix()

	path := fmt.Sprintf("/packages/%s/%s/%s.apk", req.AppID, req.ReleaseID, req.SHA256Checksum)
	sig := s.computeSignature(path, expUnix, req.DeviceOrUserID)

	signedURL := fmt.Sprintf("https://%s%s?exp=%d&uid=%s&sig=%s",
		s.cdnDomain, path, expUnix, url.QueryEscape(req.DeviceOrUserID), sig)

	return &SignedDownloadURL{
		URL:            signedURL,
		ExpiresAt:      expiresAt,
		SHA256Checksum: req.SHA256Checksum,
	}, nil
}

// ValidateSignedURL verifies the cryptographic signature, expiration time, and kill switch revocation
func (s *DownloadTokenService) ValidateSignedURL(ctx context.Context, rawURL string) (bool, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false, errors.New("invalid URL")
	}

	query := parsedURL.Query()
	expStr := query.Get("exp")
	uid := query.Get("uid")
	sig := query.Get("sig")

	if expStr == "" || sig == "" {
		return false, ErrDownloadTokenInvalidSignature
	}

	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return false, ErrDownloadTokenInvalidSignature
	}

	// Check Expiration (15 minutes)
	if time.Now().UTC().Unix() > expUnix {
		return false, ErrDownloadTokenExpired
	}

	// Check Signature
	expectedSig := s.computeSignature(parsedURL.Path, expUnix, uid)
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return false, ErrDownloadTokenInvalidSignature
	}

	// Extract releaseID from path: /packages/{app_id}/{release_id}/{sha256}.apk
	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) >= 3 {
		releaseID := parts[2]
		s.mu.RLock()
		if s.revokedReleases[releaseID] {
			s.mu.RUnlock()
			return false, ErrDownloadTokenRevoked
		}
		s.mu.RUnlock()
	}

	return true, nil
}

// RevokeRelease registers a release into the revocation blacklist for instant kill-switch
func (s *DownloadTokenService) RevokeRelease(releaseID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revokedReleases[releaseID] = true
}

func (s *DownloadTokenService) computeSignature(path string, exp int64, uid string) string {
	mac := hmac.New(sha256.New, s.hmacSecret)
	message := fmt.Sprintf("%s|%d|%s", path, exp, uid)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
