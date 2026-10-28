package domain

import (
	"errors"
	"time"
)

// OrgVerificationStatus states.
type OrgVerificationStatus string

const (
	OrgUnverified     OrgVerificationStatus = "UNVERIFIED"
	OrgPendingReview  OrgVerificationStatus = "PENDING_REVIEW"
	OrgActionRequired OrgVerificationStatus = "ACTION_REQUIRED"
	OrgVerified       OrgVerificationStatus = "VERIFIED"
	OrgRejected       OrgVerificationStatus = "REJECTED"
	OrgSuspended      OrgVerificationStatus = "SUSPENDED"
)

// DevRole codes.
type DevRole string

const (
	RoleDevOwner       DevRole = "DEV_OWNER"
	RoleReleaseManager DevRole = "DEV_RELEASE_MGR"
	RoleListingEditor  DevRole = "DEV_EDITOR"
	RoleFinanceManager DevRole = "DEV_FINANCE"
	RoleSupportAgent   DevRole = "DEV_SUPPORT"
)

// Errors
var (
	ErrOrgNotFound          = errors.New("org not found")
	ErrOrgNotVerified       = errors.New("ERR_DEV_ORG_NOT_VERIFIED: organization is not verified")
	ErrForbiddenTenantAccess = errors.New("ERR_FORBIDDEN_TENANT_ACCESS: forbidden tenant access")
	ErrMemberNotFound       = errors.New("member not found")
	ErrAPIKeyNotFound       = errors.New("api key not found")
	ErrInvalidOrgSlug       = errors.New("invalid org slug")
	ErrDuplicateOrgSlug     = errors.New("org slug already exists")
)

// DeveloperOrg represents a developer organization.
type DeveloperOrg struct {
	ID                string                `json:"id"`
	OrgSlug           string                `json:"org_slug"`
	OrgName           string                `json:"org_name"`
	VerificationStatus OrgVerificationStatus `json:"verification_status"`
	MFARequired       bool                  `json:"mfa_required"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

// DeveloperVerification represents a KYB submission.
type DeveloperVerification struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"org_id"`
	DocType     string    `json:"doc_type"`
	DocURL      string    `json:"doc_url"`
	Domain      string    `json:"domain"`
	Status      string    `json:"status"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// OrgMember represents membership.
type OrgMember struct {
	OrgID  string  `json:"org_id"`
	UserID string  `json:"user_id"`
	Role   DevRole `json:"role"`
}

// APIKey represents a scoped CI/CD key (hash stored).
type APIKey struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	AppID     *string   `json:"app_id,omitempty"`
	Name      string    `json:"name"`
	KeyHash   string    `json:"-"`
	Scopes    []string  `json:"scopes"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

// HasScope checks if key has a scope.
func (k APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// IsValid reports whether key is usable.
func (k APIKey) IsValid(now time.Time) bool {
	if k.Revoked {
		return false
	}
	return now.Before(k.ExpiresAt)
}

// Permission matrix: role -> allowed scopes/actions.
var RolePermissions = map[DevRole]map[string]bool{
	RoleDevOwner: {
		"org:manage": true, "member:invite": true, "app:create": true,
		"release:upload": true, "release:submit": true, "listing:edit": true,
		"finance:view": true, "review:reply": true,
	},
	RoleReleaseManager: {
		"app:create": true, "release:upload": true, "release:submit": true, "listing:edit": true,
	},
	RoleListingEditor: {
		"listing:edit": true,
	},
	RoleFinanceManager: {
		"finance:view": true,
	},
	RoleSupportAgent: {
		"review:reply": true,
	},
}

// RoleHasPermission checks role permission.
func RoleHasPermission(role DevRole, perm string) bool {
	if perms, ok := RolePermissions[role]; ok {
		return perms[perm]
	}
	return false
}
