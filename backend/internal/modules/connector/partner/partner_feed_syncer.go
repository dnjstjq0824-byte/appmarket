package partner

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/connector/domain"
)

var (
	ErrInvalidPartnerURL     = errors.New("ERR_INVALID_PARTNER_OFFICIAL_URL")
	ErrUnapprovedStoreDomain = errors.New("ERR_UNAPPROVED_STORE_DOMAIN")
)

type PartnerSyncer struct {
	rightsGuard *domain.RightsGuard
}

func NewPartnerSyncer() *PartnerSyncer {
	return &PartnerSyncer{
		rightsGuard: domain.NewRightsGuard(),
	}
}

// IngestPartnerApp ingests verified external store metadata and enforces Level 4 official link only
func (s *PartnerSyncer) IngestPartnerApp(storeName, officialStoreURL, packageName string) (*domain.ProvenanceMetadata, error) {
	parsed, err := url.Parse(officialStoreURL)
	if err != nil {
		return nil, ErrInvalidPartnerURL
	}

	// Whitelisted schemes: https, samsungapps, market
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && scheme != "samsungapps" && scheme != "market" {
		return nil, ErrInvalidPartnerURL
	}

	// Verify recognized official domains
	if scheme == "https" {
		host := strings.ToLower(parsed.Host)
		if !strings.HasSuffix(host, "google.com") && !strings.HasSuffix(host, "samsung.com") && !strings.HasSuffix(host, "github.com") {
			return nil, ErrUnapprovedStoreDomain
		}
	}

	meta, err := s.rightsGuard.FormatProvenance(storeName, officialStoreURL, time.Now().UTC(), domain.Level4OfficialLinkOnly)
	if err != nil {
		return nil, err
	}

	return meta, nil
}
