package presentation

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
)

// DeveloperHandler serves developer endpoints.
type DeveloperHandler struct {
	orgs  *application.OrgService
	members *application.MemberService
	keys  *application.APIKeyService
}

// NewDeveloperHandler constructs handler.
func NewDeveloperHandler(orgs *application.OrgService, members *application.MemberService, keys *application.APIKeyService) *DeveloperHandler {
	return &DeveloperHandler{orgs: orgs, members: members, keys: keys}
}

// Routes returns chi router.
func (h *DeveloperHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/orgs", h.handleCreateOrg)
	r.Post("/orgs/{orgId}/verification", h.handleSubmitVerification)
	r.Post("/orgs/{orgId}/members", h.handleAddMember)
	r.Post("/orgs/{orgId}/api-keys", h.handleIssueAPIKey)
	return r
}

type createOrgReq struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func (h *DeveloperHandler) handleCreateOrg(w http.ResponseWriter, r *http.Request) {
	var req createOrgReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	org, err := h.orgs.CreateOrg(r.Context(), req.Slug, req.Name)
	if err != nil {
		if err == domain.ErrDuplicateOrgSlug {
			errors.WriteProblem(w, http.StatusConflict, "ERR_ORG_SLUG_EXISTS", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_ORG_CREATE_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(org)
}

type submitVerificationReq struct {
	DocType    string `json:"doc_type"`
	DocURL     string `json:"doc_url"`
	Domain     string `json:"domain"`
}

func (h *DeveloperHandler) handleSubmitVerification(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	var req submitVerificationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	v, err := h.orgs.SubmitVerification(r.Context(), orgID, req.DocType, req.DocURL, req.Domain)
	if err != nil {
		if err == domain.ErrOrgNotFound {
			errors.WriteProblem(w, http.StatusNotFound, "ERR_ORG_NOT_FOUND", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_VERIFICATION_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(v)
}

type addMemberReq struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (h *DeveloperHandler) handleAddMember(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	var req addMemberReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	m, err := h.members.AddMember(r.Context(), orgID, req.UserID, domain.DevRole(req.Role))
	if err != nil {
		if err == domain.ErrOrgNotFound {
			errors.WriteProblem(w, http.StatusNotFound, "ERR_ORG_NOT_FOUND", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_MEMBER_ADD_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(m)
}

type issueKeyReq struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
	TTLDays int     `json:"ttl_days"`
	AppID  *string  `json:"app_id,omitempty"`
}

func (h *DeveloperHandler) handleIssueAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	var req issueKeyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	raw, k, err := h.keys.IssueAPIKey(r.Context(), orgID, req.AppID, req.Name, req.Scopes, req.TTLDays)
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_APIKEY_ISSUE_FAILED", err.Error(), r.URL.Path)
		return
	}
	resp := map[string]any{
		"api_key":   raw,
		"record":    k,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
