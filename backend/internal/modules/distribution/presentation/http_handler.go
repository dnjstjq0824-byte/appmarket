package presentation

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
)

// CandidateProvider provides candidate releases for requested package names.
type CandidateProvider interface {
	GetCandidatesForPackages(ctx context.Context, packageNames []string) (map[string]domain.CandidateRelease, error)
}

// DistributionHandler serves distribution, token, and device update endpoints.
type DistributionHandler struct {
	tokenService      *application.DownloadTokenService
	candidateProvider CandidateProvider
}

// NewDistributionHandler creates a new DistributionHandler instance.
func NewDistributionHandler(
	tokenService *application.DownloadTokenService,
	candidateProvider CandidateProvider,
) *DistributionHandler {
	return &DistributionHandler{
		tokenService:      tokenService,
		candidateProvider: candidateProvider,
	}
}

// Routes mounts the distribution HTTP routes on Chi router.
func (h *DistributionHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/apps/{appId}/releases/{releaseId}/download-token", h.handleGenerateToken)
	r.Post("/device/check-updates", h.handleCheckUpdates)
	return r
}

type tokenReqBody struct {
	SHA256Checksum   string   `json:"sha256_checksum"`
	DeviceOrUserID   string   `json:"device_or_user_id"`
	CountryCode      string   `json:"country_code"`
	BlockedCountries []string `json:"blocked_countries"`
	RightsLevel      string   `json:"rights_level"`
}

func (h *DistributionHandler) handleGenerateToken(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	releaseID := chi.URLParam(r, "releaseId")

	var body tokenReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", "Invalid JSON body", r.URL.Path)
		return
	}

	req := application.TokenRequest{
		AppID:            appID,
		ReleaseID:        releaseID,
		SHA256Checksum:   body.SHA256Checksum,
		DeviceOrUserID:   body.DeviceOrUserID,
		CountryCode:      body.CountryCode,
		BlockedCountries: body.BlockedCountries,
		RightsLevel:      body.RightsLevel,
		ReleaseStatus:    domain.ReleasePublishedFull,
	}

	signedURL, err := h.tokenService.GenerateSignedURL(req)
	if err != nil {
		switch err {
		case application.ErrExternalLinkOnlyDownloadForbidden:
			errors.WriteProblem(w, http.StatusForbidden, "ERR_EXTERNAL_LINK_ONLY_CANNOT_GENERATE_TOKEN", err.Error(), r.URL.Path)
		case application.ErrCountryGeoBlocked:
			errors.WriteProblem(w, http.StatusUnavailableForLegalReasons, "ERR_COUNTRY_GEO_BLOCKED", err.Error(), r.URL.Path)
		case application.ErrDownloadTokenRevoked:
			errors.WriteProblem(w, http.StatusGone, "ERR_DOWNLOAD_TOKEN_REVOKED", err.Error(), r.URL.Path)
		default:
			errors.WriteProblem(w, http.StatusBadRequest, "ERR_TOKEN_GENERATION_FAILED", err.Error(), r.URL.Path)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(signedURL)
}

type checkUpdatesReqBody struct {
	Device   domain.DeviceSpec         `json:"device"`
	Packages []domain.InstalledAppSpec `json:"packages"`
}

func (h *DistributionHandler) handleCheckUpdates(w http.ResponseWriter, r *http.Request) {
	var body checkUpdatesReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", "Invalid JSON body", r.URL.Path)
		return
	}

	pkgNames := make([]string, 0, len(body.Packages))
	for _, p := range body.Packages {
		pkgNames = append(pkgNames, p.PackageName)
	}

	var candidates map[string]domain.CandidateRelease
	if h.candidateProvider != nil {
		var err error
		candidates, err = h.candidateProvider.GetCandidatesForPackages(r.Context(), pkgNames)
		if err != nil {
			errors.WriteProblem(w, http.StatusInternalServerError, "ERR_UPDATE_RESOLVE_FAILED", err.Error(), r.URL.Path)
			return
		}
	} else {
		candidates = make(map[string]domain.CandidateRelease)
	}

	results := domain.BatchResolveUpdates(body.Device, body.Packages, candidates)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(results)
}
