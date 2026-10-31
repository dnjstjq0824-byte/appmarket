package domain

import (
	"errors"
	"time"
)

type UserStatus string

const (
	StatusPendingVerification UserStatus = "PENDING_EMAIL_VERIFICATION"
	StatusActive              UserStatus = "ACTIVE"
	StatusLockedBruteForce    UserStatus = "LOCKED_BRUTE_FORCE"
	StatusSuspended           UserStatus = "SUSPENDED_POLICY_VIOLATION"
	StatusDeletedAnonymized   UserStatus = "DELETED_ANONYMIZED"
)

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrUserAlreadyExists          = errors.New("user with this email already exists")
	ErrInvalidCredentials        = errors.New("invalid email or password")
	ErrAccountLocked             = errors.New("ERR_ACCOUNT_LOCKED: account locked due to consecutive failed logins")
	ErrTokenReuseDetected        = errors.New("ERR_TOKEN_REUSE_DETECTED: refresh token reuse detected, family revoked")
	ErrSessionNotFound           = errors.New("session not found")
	ErrSessionExpired            = errors.New("ERR_SESSION_EXPIRED: session has expired")
	ErrSessionRevoked            = errors.New("ERR_SESSION_REVOKED: session has been revoked")
	ErrMFARequired               = errors.New("ERR_MFA_REQUIRED: multi-factor authentication is required")
	ErrInvalidMFACode            = errors.New("ERR_INVALID_MFA_CODE: invalid TOTP code or recovery code")
	ErrDualControlSelfApproval   = errors.New("ERR_DUAL_CONTROL_SELF_APPROVAL_FORBIDDEN: initiator cannot approve own request")
	ErrDualApprovalAlreadyClosed = errors.New("ERR_DUAL_APPROVAL_ALREADY_CLOSED: request has already been decided")
	ErrUnauthorized              = errors.New("ERR_UNAUTHORIZED: authentication required")
	ErrForbidden                 = errors.New("ERR_FORBIDDEN: insufficient permissions")
)

type User struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	EmailBlindIndex     string     `json:"email_blind_index"`
	PasswordHash        string     `json:"-"`
	DisplayName         string     `json:"display_name"`
	Status              UserStatus `json:"status"`
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type AuthSession struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	TokenFamilyID     string     `json:"token_family_id"`
	RefreshTokenHash  string     `json:"refresh_token_hash"`
	DeviceFingerprint string     `json:"device_fingerprint"`
	UserAgent         string     `json:"user_agent"`
	IPAddress         string     `json:"ip_address"`
	IsRevoked         bool       `json:"is_revoked"`
	ReplacedByHash    string     `json:"replaced_by_hash,omitempty"`
	ExpiresAt         time.Time  `json:"expires_at"`
	CreatedAt         time.Time  `json:"created_at"`
	RotatedAt         *time.Time `json:"rotated_at,omitempty"`
}

type MFACredential struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	Type              string    `json:"type"` // "TOTP"
	SecretEncrypted   string    `json:"-"`
	BackupCodesHashed []string  `json:"-"`
	IsVerified        bool      `json:"is_verified"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Role struct {
	RoleCode    string   `json:"role_code"`
	RoleScope   string   `json:"role_scope"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "PENDING"
	ApprovalApproved ApprovalStatus = "APPROVED"
	ApprovalRejected ApprovalStatus = "REJECTED"
)

type DualApprovalRequest struct {
	ID               string         `json:"id"`
	OperationType    string         `json:"operation_type"` // e.g. "SETTLEMENT_PAYOUT", "PERMANENT_APP_BAN", "GRANT_ADMIN_ROLE"
	InitiatorID      string         `json:"initiator_id"`
	ApproverID       *string        `json:"approver_id,omitempty"`
	TargetResourceID string         `json:"target_resource_id"`
	PayloadJSON      string         `json:"payload_json"`
	Status           ApprovalStatus `json:"status"`
	Reason           string         `json:"reason"`
	CreatedAt        time.Time      `json:"created_at"`
	DecidedAt        *time.Time     `json:"decided_at,omitempty"`
}

