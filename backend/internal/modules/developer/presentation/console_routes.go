package presentation

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/console"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
)

// ConsoleDeps carries optional console aggregations built by the composition
// root. Concrete platform/console types are used because console results
// carry module-domain payloads that this package must not name (see
// architecture_test.go). Nil deps leave the matching routes unregistered.
type ConsoleDeps struct {
	Resolver    *console.OrgAppResolver
	Uploads     *console.UploadService
	Rollouts    *console.RolloutService
	Catalog     *console.CatalogViewService
	Metadata    *console.MetadataService
	Ads         *console.AdsViewService
	Commerce    *console.CommerceViewService
	Lineage     *console.LineageService
}

// AttachConsole stores console aggregations for route registration.
func (h *DeveloperHandler) AttachConsole(deps ConsoleDeps) {
	h.console = &deps
}

func (h *DeveloperHandler) consoleRoutes(r chi.Router) {
	if h.console == nil {
		return
	}
	auth := h.requirePermission("")
	r.With(auth).Get("/orgs", h.handleListOrgs)
	r.With(auth).Get("/orgs/{orgId}/verification", h.handleGetVerification)
	r.With(auth).Get("/orgs/{orgId}/members", h.handleListMembers)
	r.With(auth).Get("/orgs/{orgId}/api-keys", h.handleListAPIKeys)
	r.With(auth).Delete("/orgs/{orgId}/api-keys/{keyId}", h.handleRevokeAPIKey)
	if h.console.Catalog != nil {
		r.With(auth).Get("/orgs/{orgId}/apps", h.handleListOrgApps)
		r.With(auth).Get("/apps/{appId}/metadata", h.handleListMetadata)
		r.With(auth).Post("/apps/{appId}/submit", h.handleSubmitForReview)
	}
	if h.console.Metadata != nil {
		r.With(auth).Put("/apps/{appId}/metadata", h.handleSaveMetadata)
		r.With(auth).Get("/apps/{appId}/media", h.handleListMedia)
		r.With(auth).Post("/apps/{appId}/media", h.handleUploadMedia)
		r.With(auth).Get("/apps/{appId}/media/{assetId}/content", h.handleMediaContent)
		r.With(auth).Delete("/apps/{appId}/media/{assetId}", h.handleDeleteMedia)
	}
	if h.console.Uploads != nil {
		r.With(auth).Post("/apps/{appId}/uploads", h.handleInitUpload)
		r.With(auth).Post("/apps/{appId}/uploads/{uploadId}/complete", h.handleCompleteUpload)
	}
	if h.console.Rollouts != nil {
		r.With(auth).Get("/apps/{appId}/releases", h.handleListReleases)
		r.With(auth).Get("/apps/{appId}/releases/current", h.handleGetCurrentRelease)
		r.With(auth).Patch("/apps/{appId}/releases/current/rollout", h.handleUpdateRollout)
		r.With(auth).Patch("/releases/current/rollout", h.handleUpdateRolloutLegacy)
	}
	if h.console.Lineage != nil {
		r.With(auth).Get("/apps/{appId}/lineage", h.handleGetLineage)
	}
	if h.console.Ads != nil {
		r.With(auth).Get("/apps/{appId}/ads/campaigns", h.handleListCampaigns)
		r.With(auth).Post("/apps/{appId}/ads/campaigns", h.handleCreateCampaign)
		r.With(auth).Post("/apps/{appId}/ads/campaigns/{campaignId}/pause", h.handlePauseCampaign)
	}
	if h.console.Commerce != nil {
		r.With(auth).Get("/apps/{appId}/promotions", h.handleListPromotions)
		r.With(auth).Post("/apps/{appId}/promotions", h.handleCreatePromotion)
		r.With(auth).Get("/apps/{appId}/pricing", h.handleListPricing)
		r.With(auth).Post("/apps/{appId}/pricing", h.handleSetPricing)
		r.With(auth).Get("/orgs/{orgId}/settlements", h.handleOrgSettlement)
	}
}

func (h *DeveloperHandler) callerID(r *http.Request) string {
	if h.identity == nil {
		return ""
	}
	userID, _ := h.identity(r)
	return userID
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *DeveloperHandler) handleListOrgs(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.orgs.ListOrgsForUser(r.Context(), h.callerID(r))
	if err != nil {
		errors.WriteProblem(w, http.StatusInternalServerError, "ERR_ORG_LIST_FAILED", err.Error(), r.URL.Path)
		return
	}
	if orgs == nil {
		orgs = []domain.DeveloperOrg{}
	}
	writeJSON(w, http.StatusOK, orgs)
}

func (h *DeveloperHandler) handleGetVerification(w http.ResponseWriter, r *http.Request) {
	userID := h.callerID(r)
	orgID := chi.URLParam(r, "orgId")
	if err := h.members.CheckPermission(r.Context(), orgID, userID, "org:manage"); err != nil {
		errors.WriteProblem(w, http.StatusForbidden, "ERR_FORBIDDEN_TENANT_ACCESS", "Forbidden tenant access", r.URL.Path)
		return
	}
	v, err := h.orgs.GetVerification(r.Context(), orgID)
	if err != nil || v == nil {
		errors.WriteProblem(w, http.StatusNotFound, "ERR_VERIFICATION_NOT_FOUND", "No verification submitted", r.URL.Path)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (h *DeveloperHandler) handleListMembers(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	members, err := h.members.ListMembers(r.Context(), orgID)
	if err != nil {
		errors.WriteProblem(w, http.StatusInternalServerError, "ERR_MEMBER_LIST_FAILED", err.Error(), r.URL.Path)
		return
	}
	if members == nil {
		members = []domain.OrgMember{}
	}
	writeJSON(w, http.StatusOK, members)
}

func (h *DeveloperHandler) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	keys, err := h.keys.ListAPIKeys(r.Context(), orgID)
	if err != nil {
		errors.WriteProblem(w, http.StatusInternalServerError, "ERR_APIKEY_LIST_FAILED", err.Error(), r.URL.Path)
		return
	}
	if keys == nil {
		keys = []domain.APIKey{}
	}
	writeJSON(w, http.StatusOK, keys)
}

func (h *DeveloperHandler) handleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	if err := h.keys.RevokeAPIKey(r.Context(), chi.URLParam(r, "keyId")); err != nil {
		errors.WriteProblem(w, http.StatusNotFound, "ERR_APIKEY_NOT_FOUND", err.Error(), r.URL.Path)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "REVOKED"})
}

func (h *DeveloperHandler) handleListOrgApps(w http.ResponseWriter, r *http.Request) {
	apps, err := h.console.Catalog.ListOrgApps(r.Context(), h.callerID(r), chi.URLParam(r, "orgId"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	if apps == nil {
		apps = []console.AppDTO{}
	}
	writeJSON(w, http.StatusOK, apps)
}

func (h *DeveloperHandler) handleSubmitForReview(w http.ResponseWriter, r *http.Request) {
	if err := h.console.Catalog.SubmitForReview(r.Context(), h.callerID(r), r.URL.Query().Get("org"), chi.URLParam(r, "appId")); err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "PENDING_INITIAL_REVIEW"})
}

func (h *DeveloperHandler) handleListMetadata(w http.ResponseWriter, r *http.Request) {
	locs, err := h.console.Catalog.ListLocalizations(r.Context(), h.callerID(r), r.URL.Query().Get("org"), chi.URLParam(r, "appId"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	if locs == nil {
		locs = []console.LocalizationDTO{}
	}
	writeJSON(w, http.StatusOK, locs)
}

type saveMetadataReq struct {
	OrgID            string `json:"orgId"`
	Locale           string `json:"locale"`
	Title            string `json:"title"`
	ShortDescription string `json:"shortDescription"`
	FullDescription  string `json:"fullDescription"`
	WhatsNew         string `json:"whatsNew"`
}

func (h *DeveloperHandler) handleSaveMetadata(w http.ResponseWriter, r *http.Request) {
	var req saveMetadataReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	saved, err := h.console.Metadata.SaveMetadata(r.Context(), h.callerID(r), req.OrgID, chi.URLParam(r, "appId"),
		req.Locale, req.Title, req.ShortDescription, req.FullDescription, req.WhatsNew)
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (h *DeveloperHandler) handleListMedia(w http.ResponseWriter, r *http.Request) {
	list, err := h.console.Metadata.ListMedia(r.Context(), h.callerID(r), r.URL.Query().Get("org"), chi.URLParam(r, "appId"), r.URL.Query().Get("locale"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	if list == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *DeveloperHandler) handleUploadMedia(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(21 << 20); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_MEDIA_INVALID", "Malformed multipart body: "+err.Error(), r.URL.Path)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_MEDIA_INVALID", "image file is required", r.URL.Path)
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, 21<<20))
	_ = file.Close()
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_MEDIA_INVALID", "Failed to read image: "+err.Error(), r.URL.Path)
		return
	}
	asset, err := h.console.Metadata.SaveMediaImage(r.Context(), h.callerID(r), r.FormValue("org"), chi.URLParam(r, "appId"), r.FormValue("locale"), r.FormValue("formFactor"), r.FormValue("assetType"), header.Filename, data)
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, asset)
}

func (h *DeveloperHandler) handleMediaContent(w http.ResponseWriter, r *http.Request) {
	obj, err := h.console.Metadata.MediaContent(r.Context(), h.callerID(r), r.URL.Query().Get("org"), chi.URLParam(r, "appId"), chi.URLParam(r, "assetId"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Cache-Control", obj.CacheControl)
	w.Header().Set("ETag", `"`+obj.SHA256+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(obj.Content)
}

func (h *DeveloperHandler) handleDeleteMedia(w http.ResponseWriter, r *http.Request) {
	if err := h.console.Metadata.DeleteMedia(r.Context(), h.callerID(r), r.URL.Query().Get("org"), chi.URLParam(r, "appId"), chi.URLParam(r, "assetId")); err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "DELETED"})
}

type initUploadReq struct {
	OrgID          string `json:"orgId"`
	FileName       string `json:"fileName"`
	FileSize       int64  `json:"fileSize"`
	ExpectedSHA256 string `json:"expectedSha256"`
}

func (h *DeveloperHandler) handleInitUpload(w http.ResponseWriter, r *http.Request) {
	var req initUploadReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	orgID, appID, err := h.console.Resolver.ResolveOrgApp(r.Context(), h.callerID(r), req.OrgID, chi.URLParam(r, "appId"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	session, err := h.console.Uploads.InitUpload(r.Context(), orgID, appID, req.FileName, req.FileSize, req.ExpectedSHA256)
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, console.ToUploadSessionDTO(session))
}

func (h *DeveloperHandler) handleCompleteUpload(w http.ResponseWriter, r *http.Request) {
	uploadID := chi.URLParam(r, "uploadId")
	var sha string
	var apkBytes []byte
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(2 << 30); err != nil {
			errors.WriteProblem(w, http.StatusBadRequest, "ERR_UPLOAD_INVALID", "Malformed multipart body: "+err.Error(), r.URL.Path)
			return
		}
		sha = r.FormValue("sha256")
		if file, _, err := r.FormFile("apk"); err == nil {
			apkBytes, _ = io.ReadAll(io.LimitReader(file, 2<<30))
			_ = file.Close()
		}
	} else {
		var req struct {
			SHA256 string `json:"sha256"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		sha = req.SHA256
	}
	res, err := h.console.Uploads.CompleteUpload(r.Context(), uploadID, sha, apkBytes)
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, console.ToCompleteUploadDTO(res))
}

func (h *DeveloperHandler) handleListReleases(w http.ResponseWriter, r *http.Request) {
	lineage, err := h.console.Lineage.AppLineage(r.Context(), h.callerID(r), r.URL.Query().Get("org"), chi.URLParam(r, "appId"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, lineage.Releases)
}

func (h *DeveloperHandler) handleGetCurrentRelease(w http.ResponseWriter, r *http.Request) {
	_, appID, err := h.console.Resolver.ResolveOrgApp(r.Context(), h.callerID(r), r.URL.Query().Get("org"), chi.URLParam(r, "appId"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	current, err := h.console.Rollouts.CurrentRelease(r.Context(), appID)
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, console.ToReleaseDTO(current))
}

type rolloutReq struct {
	RolloutPercentage *int   `json:"rolloutPercentage"`
	Action            string `json:"action"`
	Reason            string `json:"reason"`
}

func (h *DeveloperHandler) handleUpdateRollout(w http.ResponseWriter, r *http.Request) {
	h.updateRollout(w, r, chi.URLParam(r, "appId"))
}

func (h *DeveloperHandler) handleUpdateRolloutLegacy(w http.ResponseWriter, r *http.Request) {
	h.updateRollout(w, r, "current")
}

func (h *DeveloperHandler) updateRollout(w http.ResponseWriter, r *http.Request, appHandle string) {
	var req rolloutReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	_, appID, err := h.console.Resolver.ResolveOrgApp(r.Context(), h.callerID(r), r.URL.Query().Get("org"), appHandle)
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	switch strings.ToUpper(strings.TrimSpace(req.Action)) {
	case "HALT":
		rel, err := h.console.Rollouts.Halt(r.Context(), appID, req.Reason)
		if err != nil {
			writeConsoleError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, console.ToReleaseDTO(rel))
	case "RESUME":
		rel, err := h.console.Rollouts.Resume(r.Context(), appID)
		if err != nil {
			writeConsoleError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, console.ToReleaseDTO(rel))
	case "":
		if req.RolloutPercentage == nil {
			errors.WriteProblem(w, http.StatusBadRequest, "ERR_ROLLOUT_REJECTED", "rolloutPercentage or action is required", r.URL.Path)
			return
		}
		rel, err := h.console.Rollouts.UpdateRollout(r.Context(), appID, *req.RolloutPercentage)
		if err != nil {
			writeConsoleError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, console.ToReleaseDTO(rel))
	default:
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_ROLLOUT_REJECTED", "Unknown action. Use HALT or RESUME.", r.URL.Path)
	}
}

func (h *DeveloperHandler) handleGetLineage(w http.ResponseWriter, r *http.Request) {
	lineage, err := h.console.Lineage.AppLineage(r.Context(), h.callerID(r), r.URL.Query().Get("org"), chi.URLParam(r, "appId"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, lineage)
}

type createCampaignReq struct {
	OrgID          string   `json:"orgId"`
	AppName        string   `json:"appName"`
	TargetKeywords []string `json:"targetKeywords"`
	MaxBidCPC      int64    `json:"maxBidCpc"`
	DailyBudget    int64    `json:"dailyBudget"`
}

func (h *DeveloperHandler) handleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	var req createCampaignReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	c, err := h.console.Ads.CreateCampaign(r.Context(), h.callerID(r), req.OrgID, chi.URLParam(r, "appId"),
		req.AppName, req.TargetKeywords, req.MaxBidCPC, req.DailyBudget)
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *DeveloperHandler) handleListCampaigns(w http.ResponseWriter, r *http.Request) {
	list, err := h.console.Ads.ListCampaigns(r.Context(), h.callerID(r), r.URL.Query().Get("org"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	if list == nil {
		list = []console.CampaignDTO{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *DeveloperHandler) handlePauseCampaign(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Paused bool `json:"paused"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	c, err := h.console.Ads.SetPaused(r.Context(), h.callerID(r), r.URL.Query().Get("org"), chi.URLParam(r, "campaignId"), req.Paused)
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

type createPromoReq struct {
	OrgID         string `json:"orgId"`
	CampaignName  string `json:"campaignName"`
	DiscountType  string `json:"discountType"`
	DiscountValue int    `json:"discountValue"`
	StartsAt      string `json:"startsAt"`
	EndsAt        string `json:"endsAt"`
	MaxUses       int    `json:"maxUses"`
	CodeCount     int    `json:"codeCount"`
	CodePrefix    string `json:"codePrefix"`
}

func (h *DeveloperHandler) handleCreatePromotion(w http.ResponseWriter, r *http.Request) {
	var req createPromoReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	starts, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_PROMO_INVALID", "startsAt must be RFC3339", r.URL.Path)
		return
	}
	ends, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_PROMO_INVALID", "endsAt must be RFC3339", r.URL.Path)
		return
	}
	promo, codes, err := h.console.Commerce.CreatePromotion(r.Context(), h.callerID(r), req.OrgID, chi.URLParam(r, "appId"),
		console.CreatePromoArgs{
			CampaignName: req.CampaignName, DiscountType: req.DiscountType, DiscountValue: req.DiscountValue,
			StartsAt: starts, EndsAt: ends, MaxUses: req.MaxUses, CodeCount: req.CodeCount, CodePrefix: req.CodePrefix,
		})
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"promotion": promo, "codes": codes})
}

func (h *DeveloperHandler) handleListPromotions(w http.ResponseWriter, r *http.Request) {
	list, err := h.console.Commerce.ListPromotions(r.Context(), h.callerID(r), r.URL.Query().Get("org"), chi.URLParam(r, "appId"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	if list == nil {
		list = []console.PromotionDTO{}
	}
	writeJSON(w, http.StatusOK, list)
}

type setPricingReq struct {
	OrgID        string `json:"orgId"`
	CountryCode  string `json:"countryCode"`
	Currency     string `json:"currency"`
	ListPriceMinor int64 `json:"listPriceMinor"`
	TaxInclusive bool   `json:"taxInclusive"`
}

func (h *DeveloperHandler) handleSetPricing(w http.ResponseWriter, r *http.Request) {
	var req setPricingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	tier, err := h.console.Commerce.SetPriceTier(r.Context(), h.callerID(r), req.OrgID, chi.URLParam(r, "appId"),
		req.CountryCode, req.Currency, req.ListPriceMinor, req.TaxInclusive)
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, tier)
}

func (h *DeveloperHandler) handleListPricing(w http.ResponseWriter, r *http.Request) {
	tiers, err := h.console.Commerce.ListPriceTiers(r.Context(), h.callerID(r), r.URL.Query().Get("org"), chi.URLParam(r, "appId"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	if tiers == nil {
		tiers = []console.PriceTierDTO{}
	}
	writeJSON(w, http.StatusOK, tiers)
}

func (h *DeveloperHandler) handleOrgSettlement(w http.ResponseWriter, r *http.Request) {
	settlement, err := h.console.Commerce.OrgSettlement(r.Context(), h.callerID(r), chi.URLParam(r, "orgId"))
	if err != nil {
		writeConsoleError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, settlement)
}

func writeConsoleError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusBadRequest
	code := "ERR_CONSOLE_REQUEST_FAILED"
	switch err {
	case console.ErrNoOrg, console.ErrNoApp:
		status, code = http.StatusNotFound, "ERR_CONSOLE_EMPTY"
	case console.ErrOrgAccessDenied, console.ErrAppAccessDenied:
		status, code = http.StatusForbidden, "ERR_FORBIDDEN_TENANT_ACCESS"
	case console.ErrUploadNotFound:
		status, code = http.StatusNotFound, "ERR_UPLOAD_NOT_FOUND"
	case console.ErrMediaNotFound:
		status, code = http.StatusNotFound, "ERR_MEDIA_NOT_FOUND"
	case console.ErrNoRelease:
		status, code = http.StatusNotFound, "ERR_NO_RELEASE"
	case console.ErrUploadExpired:
		status, code = http.StatusGone, "ERR_UPLOAD_EXPIRED"
	case console.ErrChecksumMismatch:
		status, code = http.StatusConflict, "ERR_UPLOAD_CHECKSUM_MISMATCH"
	case console.ErrRolloutRejected:
		status, code = http.StatusConflict, "ERR_ROLLOUT_REJECTED"
	}
	errors.WriteProblem(w, status, code, err.Error(), r.URL.Path)
}
