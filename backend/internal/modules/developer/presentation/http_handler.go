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
	orgs     *application.OrgService
	members  *application.MemberService
	keys     *application.APIKeyService
	payouts  *application.PayoutService
	identity IdentityResolver
	console  *ConsoleDeps
}

// AttachPayouts stores the payout service built by the composition root.
func (h *DeveloperHandler) AttachPayouts(payouts *application.PayoutService) {
	h.payouts = payouts
}

// IdentityResolver must return a user ID from verified authentication only.
// The composition root supplies the IAM adapter; request headers are not identity.
type IdentityResolver func(*http.Request) (string, error)

// NewDeveloperHandler constructs handler.
func NewDeveloperHandler(orgs *application.OrgService, members *application.MemberService, keys *application.APIKeyService, identity ...IdentityResolver) *DeveloperHandler {
	h := &DeveloperHandler{orgs: orgs, members: members, keys: keys}
	if len(identity) == 1 {
		h.identity = identity[0]
	}
	return h
}

// Routes returns chi router.
func (h *DeveloperHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.With(h.requirePermission("")).Post("/orgs", h.handleCreateOrg)
	r.With(h.requirePermission("org:manage")).Post("/orgs/{orgId}/verification", h.handleSubmitVerification)
	r.With(h.requirePermission("member:invite")).Post("/orgs/{orgId}/members", h.handleAddMember)
	r.With(h.requirePermission("member:remove")).Delete("/orgs/{orgId}/members/{userId}", h.handleRemoveMember)
	r.With(h.requirePermission("org:manage")).Post("/orgs/{orgId}/api-keys", h.handleIssueAPIKey)
	if h.payouts != nil {
		r.With(h.requirePermission("org:manage")).Get("/orgs/{orgId}/payout-account", h.handleGetPayoutAccount)
		r.With(h.requirePermission("org:manage")).Put("/orgs/{orgId}/payout-account", h.handleSavePayoutAccount)
	}
	h.consoleRoutes(r)
	return r
}

func (h *DeveloperHandler) requirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if h.identity == nil {
				errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Authentication required", r.URL.Path)
				return
			}
			userID, err := h.identity(r)
			if err != nil || userID == "" {
				errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Authentication required", r.URL.Path)
				return
			}
			if permission != "" {
				orgID := chi.URLParam(r, "orgId")
				if orgID == "" || h.members == nil || h.members.CheckPermission(r.Context(), orgID, userID, permission) != nil {
					errors.WriteProblem(w, http.StatusForbidden, "ERR_FORBIDDEN_TENANT_ACCESS", "Forbidden tenant access", r.URL.Path)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
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
	if userID, idErr := h.identity(r); idErr == nil && userID != "" {
		if _, err := h.members.AddMember(r.Context(), org.ID, userID, domain.RoleDevOwner); err != nil {
			errors.WriteProblem(w, http.StatusInternalServerError, "ERR_ORG_OWNER_LINK_FAILED", err.Error(), r.URL.Path)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(org)
}

type submitVerificationReq struct {
	DocType string `json:"doc_type"`
	DocURL  string `json:"doc_url"`
	Domain  string `json:"domain"`
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

func (h *DeveloperHandler) handleRemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	userID := chi.URLParam(r, "userId")
	if err := h.members.RemoveMember(r.Context(), orgID, userID); err != nil {
		if err == domain.ErrMemberNotFound {
			errors.WriteProblem(w, http.StatusNotFound, "ERR_MEMBER_NOT_FOUND", err.Error(), r.URL.Path)
			return
		}
		if err == domain.ErrLastOwnerRemoval {
			errors.WriteProblem(w, http.StatusConflict, "ERR_LAST_OWNER_REMOVAL", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_MEMBER_REMOVE_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "REMOVED"})
}

func (h *DeveloperHandler) handleGetPayoutAccount(w http.ResponseWriter, r *http.Request) {
	account, err := h.payouts.GetPayoutAccount(r.Context(), chi.URLParam(r, "orgId"))
	if err != nil {
		if err == domain.ErrPayoutAccountNotFound {
			errors.WriteProblem(w, http.StatusNotFound, "ERR_PAYOUT_ACCOUNT_NOT_FOUND", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusInternalServerError, "ERR_PAYOUT_ACCOUNT_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(account)
}

type savePayoutReq struct {
	Currency      string `json:"currency"`
	BankName      string `json:"bank_name"`
	HolderName    string `json:"holder_name"`
	AccountNumber string `json:"account_number"`
	RoutingCode   string `json:"routing_code"`
}

func (h *DeveloperHandler) handleSavePayoutAccount(w http.ResponseWriter, r *http.Request) {
	var req savePayoutReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	account, err := h.payouts.SavePayoutAccount(r.Context(), chi.URLParam(r, "orgId"), application.PayoutAccountInput{
		Currency: req.Currency, BankName: req.BankName, HolderName: req.HolderName,
		AccountNumber: req.AccountNumber, RoutingCode: req.RoutingCode,
	})
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_PAYOUT_ACCOUNT_INVALID", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(account)
}

type issueKeyReq struct {
	Name    string   `json:"name"`
	Scopes  []string `json:"scopes"`
	TTLDays int      `json:"ttl_days"`
	AppID   *string  `json:"app_id,omitempty"`
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
		"api_key": raw,
		"record":  k,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
