package rbac

import (
	"context"
	"net/http"
	"strings"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
)

type contextKey string

const (
	ClaimsContextKey contextKey = "claims"
)

// RolePermissions maps platform roles to their allowed permissions according to AUTHORIZATION_MATRIX.md.
var RolePermissions = map[string][]string{
	"USER": {
		"catalog:read",
		"review:write",
		"order:create",
		"user:profile:manage",
	},
	"DEV_OWNER": {
		"catalog:read",
		"dev:app:manage",
		"dev:release:publish",
		"dev:finance:manage",
		"dev:team:manage",
	},
	"DEV_RELEASE_MGR": {
		"catalog:read",
		"dev:app:manage",
		"dev:release:publish",
	},
	"DEV_EDITOR": {
		"catalog:read",
		"dev:app:manage",
	},
	"DEV_FINANCE": {
		"catalog:read",
		"dev:finance:manage",
	},
	"ADMIN_SUPER": {
		"*",
	},
	"ADMIN_KYB_REVIEWER": {
		"admin:kyb:read",
		"admin:kyb:decide",
	},
	"ADMIN_APP_REVIEWER": {
		"admin:review:read",
		"admin:review:decide",
	},
	"ADMIN_SECURITY": {
		"admin:security:read",
		"admin:security:decide",
		"admin:killswitch:execute",
	},
	"ADMIN_MODERATOR": {
		"admin:moderation:read",
		"admin:moderation:decide",
	},
	"ADMIN_SEARCH_OPS": {
		"admin:search:manage",
		"admin:connector:manage",
	},
	"ADMIN_FINANCE": {
		"admin:finance:read",
		"admin:refund:decide",
		"admin:settlement:approve",
	},
	"ADMIN_AUDITOR": {
		"admin:audit:read",
	},
}

type MatrixEnforcer struct {
	jwtManager *application.JWTManager
}

func NewMatrixEnforcer(jwtManager *application.JWTManager) *MatrixEnforcer {
	return &MatrixEnforcer{jwtManager: jwtManager}
}

// HasPermission checks if the set of roles contains the required permission (or wildcard "*").
func HasPermission(roles []string, requiredPermission string) bool {
	for _, role := range roles {
		perms, ok := RolePermissions[role]
		if !ok {
			continue
		}
		for _, p := range perms {
			if p == "*" || p == requiredPermission {
				return true
			}
		}
	}
	return false
}

// AuthenticateMiddleware validates the Bearer JWT token and places claims in the request context.
func (e *MatrixEnforcer) AuthenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Missing or invalid Authorization header", r.URL.Path)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := e.jwtManager.ValidateAccessToken(tokenStr)
		if err != nil {
			errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Invalid or expired access token", r.URL.Path)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermissionMiddleware enforces that the authenticated user possesses the specific permission.
func (e *MatrixEnforcer) RequirePermissionMiddleware(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsContextKey).(*application.AccessClaims)
			if !ok || claims == nil {
				errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Authentication required", r.URL.Path)
				return
			}

			if !HasPermission(claims.Roles, permission) {
				errors.WriteProblem(w, http.StatusForbidden, "ERR_FORBIDDEN", "Insufficient permissions for this operation", r.URL.Path)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireMFAMiddleware enforces that the user has verified MFA.
func (e *MatrixEnforcer) RequireMFAMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(ClaimsContextKey).(*application.AccessClaims)
		if !ok || claims == nil {
			errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Authentication required", r.URL.Path)
			return
		}

		if !claims.MFAVerified {
			errors.WriteProblem(w, http.StatusForbidden, "ERR_MFA_REQUIRED", domain.ErrMFARequired.Error(), r.URL.Path)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetClaims retrieves claims from the context.
func GetClaims(ctx context.Context) (*application.AccessClaims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*application.AccessClaims)
	return claims, ok
}

