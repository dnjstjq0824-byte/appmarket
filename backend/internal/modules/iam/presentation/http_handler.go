package presentation

import (
	"encoding/json"
	"net/http"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/rbac"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	authService  *application.AuthService
	jwtManager   *application.JWTManager
	dualControl  *rbac.DualControlEngine
	enforcer     *rbac.MatrixEnforcer
	oauthService *application.OAuthService
	baseURL      string
}

func NewAuthHandler(
	authService *application.AuthService,
	jwtManager *application.JWTManager,
	dualControl *rbac.DualControlEngine,
	enforcer *rbac.MatrixEnforcer,
	baseURL string,
	oauthService *application.OAuthService,
) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		jwtManager:   jwtManager,
		dualControl:  dualControl,
		enforcer:     enforcer,
		oauthService: oauthService,
		baseURL:      baseURL,
	}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Public Auth endpoints
	r.Post("/register", h.handleRegister)
	r.Post("/login", h.handleLogin)
	r.Post("/token/refresh", h.handleRefreshToken)
	r.Post("/mfa/verify", h.handleMFAVerify)
	r.Post("/mfa/enroll", h.handleMFAEnroll)
	r.Get("/.well-known/openid-configuration", h.handleOpenIDConfig)
	r.Get("/jwks.json", h.handleJWKS)
	if h.oauthService != nil {
		r.Get("/oauth/providers", h.handleOAuthProviders)
		r.Get("/oauth/{provider}/login", h.handleOAuthLogin)
		r.Get("/oauth/{provider}/callback", h.handleOAuthCallback)
	}

	// Protected endpoints
	r.Group(func(protected chi.Router) {
		protected.Use(h.enforcer.AuthenticateMiddleware)
		protected.Post("/logout", h.handleLogout)
		protected.Get("/users/me", h.handleGetMe)

		// Admin dual-control approval endpoints
		guard := func(permission string) func(http.Handler) http.Handler {
			return h.enforcer.RequirePermissionMiddleware(permission)
		}
		mfa := h.enforcer.RequireMFAMiddleware
		protected.With(guard("admin:approval:read")).Get("/admin/approvals/pending", h.handleListPendingApprovals)
		protected.With(guard("admin:approval:decide")).Post("/admin/approvals", h.handleCreateApproval)
		protected.With(guard("admin:approval:decide"), mfa).Post("/admin/approvals/{requestId}/approve", h.handleApproveRequest)
		protected.With(guard("admin:approval:decide"), mfa).Post("/admin/approvals/{requestId}/reject", h.handleRejectRequest)
	})

	return r
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var input application.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", "Invalid JSON request body", r.URL.Path)
		return
	}

	user, err := h.authService.Register(r.Context(), input)
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			errors.WriteProblem(w, http.StatusConflict, "ERR_USER_EXISTS", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_REGISTRATION_FAILED", err.Error(), r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) handleOAuthProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.oauthService.Providers()
	if providers == nil {
		providers = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"providers": providers})
}

func (h *AuthHandler) handleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	state, err := h.oauthService.GenerateState()
	if err != nil {
		errors.WriteProblem(w, http.StatusInternalServerError, "ERR_OAUTH_STATE_FAILED", err.Error(), r.URL.Path)
		return
	}
	target, err := h.oauthService.LoginURL(provider, state)
	if err != nil {
		errors.WriteProblem(w, http.StatusNotFound, "ERR_OAUTH_PROVIDER_UNKNOWN", err.Error(), r.URL.Path)
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *AuthHandler) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_OAUTH_PROVIDER_DENIED", "Provider refused authorization: "+errParam, r.URL.Path)
		return
	}
	result, err := h.oauthService.HandleCallback(r.Context(), provider, r.URL.Query().Get("code"), r.URL.Query().Get("state"), r.RemoteAddr, r.UserAgent())
	if err != nil {
		status := http.StatusBadGateway
		code := "ERR_OAUTH_FAILED"
		switch err {
		case domain.ErrOAuthProviderUnknown:
			status, code = http.StatusNotFound, "ERR_OAUTH_PROVIDER_UNKNOWN"
		case domain.ErrOAuthStateInvalid:
			status, code = http.StatusBadRequest, "ERR_OAUTH_STATE_INVALID"
		case domain.ErrAccountLocked:
			status, code = http.StatusForbidden, "ERR_ACCOUNT_LOCKED"
		case domain.ErrOAuthEmailMissing:
			status, code = http.StatusBadRequest, "ERR_OAUTH_EMAIL_MISSING"
		}
		errors.WriteProblem(w, status, code, err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", "Invalid JSON request body", r.URL.Path)
		return
	}

	result, err := h.authService.Login(r.Context(), req.Email, req.Password, r.RemoteAddr, r.UserAgent())
	if err != nil {
		if err == domain.ErrAccountLocked {
			errors.WriteProblem(w, http.StatusForbidden, "ERR_ACCOUNT_LOCKED", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusUnauthorized, "ERR_INVALID_CREDENTIALS", err.Error(), r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", "Invalid JSON request body", r.URL.Path)
		return
	}

	result, err := h.authService.RefreshToken(r.Context(), req.RefreshToken, r.RemoteAddr, r.UserAgent())
	if err != nil {
		if err == domain.ErrTokenReuseDetected {
			// TC-SEC-005: 401 ERR_TOKEN_REUSE_DETECTED
			errors.WriteProblem(w, http.StatusUnauthorized, "ERR_TOKEN_REUSE_DETECTED", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusUnauthorized, "ERR_SESSION_INVALID", err.Error(), r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

type MFAVerifyRequest struct {
	UserID  string `json:"user_id"`
	MFACode string `json:"mfa_code"`
}

func (h *AuthHandler) handleMFAVerify(w http.ResponseWriter, r *http.Request) {
	var req MFAVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", "Invalid JSON request body", r.URL.Path)
		return
	}

	result, err := h.authService.CompleteLoginWithMFA(r.Context(), req.UserID, req.MFACode, r.RemoteAddr, r.UserAgent())
	if err != nil {
		errors.WriteProblem(w, http.StatusUnauthorized, "ERR_INVALID_MFA_CODE", err.Error(), r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

type MFAEnrollRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) handleMFAEnroll(w http.ResponseWriter, r *http.Request) {
	var req MFAEnrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", "Invalid JSON request body", r.URL.Path)
		return
	}

	enrolled, err := h.authService.EnrollMFA(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			errors.WriteProblem(w, http.StatusUnauthorized, "ERR_INVALID_CREDENTIALS", err.Error(), r.URL.Path)
		case domain.ErrMFAAlreadyEnrolled:
			errors.WriteProblem(w, http.StatusConflict, "ERR_MFA_ALREADY_ENROLLED", err.Error(), r.URL.Path)
		default:
			errors.WriteProblem(w, http.StatusInternalServerError, "ERR_MFA_ENROLL_FAILED", err.Error(), r.URL.Path)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(enrolled)
}

func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	_ = h.authService.Logout(r.Context(), req.RefreshToken)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := rbac.GetClaims(r.Context())
	if !ok || claims == nil {
		errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Unauthorized", r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(claims)
}

func (h *AuthHandler) handleJWKS(w http.ResponseWriter, r *http.Request) {
	jwks := h.jwtManager.GetJWKS()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(jwks)
}

func (h *AuthHandler) handleOpenIDConfig(w http.ResponseWriter, r *http.Request) {
	config := h.jwtManager.GetOpenIDConfiguration(h.baseURL)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config)
}

func (h *AuthHandler) handleListPendingApprovals(w http.ResponseWriter, r *http.Request) {
	list, err := h.dualControl.ListPending(r.Context())
	if err != nil {
		errors.WriteProblem(w, http.StatusInternalServerError, "ERR_APPROVAL_LIST_FAILED", err.Error(), r.URL.Path)
		return
	}
	if list == nil {
		list = []*domain.DualApprovalRequest{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (h *AuthHandler) handleCreateApproval(w http.ResponseWriter, r *http.Request) {
	claims, ok := rbac.GetClaims(r.Context())
	if !ok || claims == nil {
		errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Unauthorized", r.URL.Path)
		return
	}
	var req struct {
		OperationType    string `json:"operation_type"`
		TargetResourceID string `json:"target_resource_id"`
		PayloadJSON      string `json:"payload_json"`
		Reason           string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", "Invalid JSON request body", r.URL.Path)
		return
	}
	if req.OperationType == "" || req.Reason == "" {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", "operation_type and reason are required", r.URL.Path)
		return
	}
	created, err := h.dualControl.CreateRequest(r.Context(), req.OperationType, claims.UserID, req.TargetResourceID, req.PayloadJSON, req.Reason)
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_APPROVAL_CREATE_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

// handleApproveRequest processes admin dual-control approval (TC-SEC-006).
func (h *AuthHandler) handleApproveRequest(w http.ResponseWriter, r *http.Request) {
	claims, ok := rbac.GetClaims(r.Context())
	if !ok || claims == nil {
		errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Unauthorized", r.URL.Path)
		return
	}

	requestID := chi.URLParam(r, "requestId")
	approvedReq, err := h.dualControl.Approve(r.Context(), requestID, claims.UserID)
	if err != nil {
		if err == domain.ErrDualControlSelfApproval {
			// TC-SEC-006: 403 ERR_DUAL_CONTROL_SELF_APPROVAL_FORBIDDEN
			errors.WriteProblem(w, http.StatusForbidden, "ERR_DUAL_CONTROL_SELF_APPROVAL_FORBIDDEN", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_APPROVAL_FAILED", err.Error(), r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(approvedReq)
}

func (h *AuthHandler) handleRejectRequest(w http.ResponseWriter, r *http.Request) {
	claims, ok := rbac.GetClaims(r.Context())
	if !ok || claims == nil {
		errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Unauthorized", r.URL.Path)
		return
	}

	requestID := chi.URLParam(r, "requestId")
	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	rejectedReq, err := h.dualControl.Reject(r.Context(), requestID, claims.UserID, req.Reason)
	if err != nil {
		if err == domain.ErrDualControlSelfApproval {
			errors.WriteProblem(w, http.StatusForbidden, "ERR_DUAL_CONTROL_SELF_APPROVAL_FORBIDDEN", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_REJECTION_FAILED", err.Error(), r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rejectedReq)
}

