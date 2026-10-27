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
	authService *application.AuthService
	jwtManager  *application.JWTManager
	dualControl *rbac.DualControlEngine
	enforcer    *rbac.MatrixEnforcer
	baseURL     string
}

func NewAuthHandler(
	authService *application.AuthService,
	jwtManager *application.JWTManager,
	dualControl *rbac.DualControlEngine,
	enforcer *rbac.MatrixEnforcer,
	baseURL string,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtManager:  jwtManager,
		dualControl: dualControl,
		enforcer:    enforcer,
		baseURL:     baseURL,
	}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Public Auth endpoints
	r.Post("/register", h.handleRegister)
	r.Post("/login", h.handleLogin)
	r.Post("/token/refresh", h.handleRefreshToken)
	r.Post("/mfa/verify", h.handleMFAVerify)
	r.Get("/.well-known/openid-configuration", h.handleOpenIDConfig)
	r.Get("/jwks.json", h.handleJWKS)

	// Protected endpoints
	r.Group(func(protected chi.Router) {
		protected.Use(h.enforcer.AuthenticateMiddleware)
		protected.Post("/logout", h.handleLogout)
		protected.Get("/users/me", h.handleGetMe)

		// Admin dual-control approval endpoints
		protected.Post("/admin/approvals/{requestId}/approve", h.handleApproveRequest)
		protected.Post("/admin/approvals/{requestId}/reject", h.handleRejectRequest)
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
