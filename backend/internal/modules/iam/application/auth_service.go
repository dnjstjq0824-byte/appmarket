package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmailBlindIndex(ctx context.Context, blindIndex string) (*domain.User, error)
	UpdateUserStatus(ctx context.Context, id string, status domain.UserStatus, failedAttempts int, lockedUntil *time.Time) error
	GetUserRoles(ctx context.Context, userID string) ([]string, map[string]string, error)
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session *domain.AuthSession) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthSession, error)
	RotateSession(ctx context.Context, oldSessionID string, newSession *domain.AuthSession) error
	RevokeFamily(ctx context.Context, tokenFamilyID string) error
}

type MFARepository interface {
	GetMFACredential(ctx context.Context, userID string) (*domain.MFACredential, error)
	SaveMFACredential(ctx context.Context, cred *domain.MFACredential) error
}

type AuthServiceConfig struct {
	HMACSecret     string
	MaxFailedLogin int
	LockoutDuration time.Duration
}

type AuthService struct {
	userRepo    UserRepository
	sessionRepo SessionRepository
	mfaRepo     MFARepository
	hasher      *PasswordHasher
	jwtManager  *JWTManager
	mfaService  *MFAService
	cfg         AuthServiceConfig
	mu          sync.RWMutex
}

func NewAuthService(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	mfaRepo MFARepository,
	hasher *PasswordHasher,
	jwtManager *JWTManager,
	mfaService *MFAService,
	cfg AuthServiceConfig,
) *AuthService {
	if cfg.MaxFailedLogin <= 0 {
		cfg.MaxFailedLogin = 5
	}
	if cfg.LockoutDuration <= 0 {
		cfg.LockoutDuration = 15 * time.Minute
	}
	if cfg.HMACSecret == "" {
		cfg.HMACSecret = "default-development-hmac-secret-32-chars-long"
	}
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		mfaRepo:     mfaRepo,
		hasher:      hasher,
		jwtManager:  jwtManager,
		mfaService:  mfaService,
		cfg:         cfg,
	}
}

type RegisterInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type AuthResult struct {
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	ExpiresIn    int64        `json:"expires_in,omitempty"`
	TokenType    string       `json:"token_type,omitempty"`
	MFARequired  bool         `json:"mfa_required"`
	User         *domain.User `json:"user,omitempty"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("invalid email address")
	}
	if len(input.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	blindIndex := ComputeBlindIndex(email, s.cfg.HMACSecret)
	existing, _ := s.userRepo.GetUserByEmailBlindIndex(ctx, blindIndex)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	passwordHash, err := s.hasher.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:                  uuid.New().String(),
		Email:               email,
		EmailBlindIndex:     blindIndex,
		PasswordHash:        passwordHash,
		DisplayName:         input.DisplayName,
		Status:              domain.StatusActive,
		FailedLoginAttempts: 0,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password, ipAddress, userAgent string) (*AuthResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	email = strings.ToLower(strings.TrimSpace(email))
	blindIndex := ComputeBlindIndex(email, s.cfg.HMACSecret)

	user, err := s.userRepo.GetUserByEmailBlindIndex(ctx, blindIndex)
	if err != nil || user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	now := time.Now().UTC()
	if user.LockedUntil != nil && user.LockedUntil.After(now) {
		return nil, domain.ErrAccountLocked
	}

	match, err := s.hasher.VerifyPassword(password, user.PasswordHash)
	if err != nil || !match {
		newFailed := user.FailedLoginAttempts + 1
		var lockedUntil *time.Time
		status := user.Status

		if newFailed >= s.cfg.MaxFailedLogin {
			t := now.Add(s.cfg.LockoutDuration)
			lockedUntil = &t
			status = domain.StatusLockedBruteForce
		}

		_ = s.userRepo.UpdateUserStatus(ctx, user.ID, status, newFailed, lockedUntil)
		if newFailed >= s.cfg.MaxFailedLogin {
			return nil, domain.ErrAccountLocked
		}
		return nil, domain.ErrInvalidCredentials
	}

	// Reset failed attempts on successful password check
	if user.FailedLoginAttempts > 0 || user.LockedUntil != nil {
		_ = s.userRepo.UpdateUserStatus(ctx, user.ID, domain.StatusActive, 0, nil)
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
	}

	roles, orgRoles, _ := s.userRepo.GetUserRoles(ctx, user.ID)
	if len(roles) == 0 {
		roles = []string{"USER"}
	}

	// Check MFA
	mfaCred, _ := s.mfaRepo.GetMFACredential(ctx, user.ID)
	isMFAEnforced := isMFAMandatoryForRoles(roles)
	if (mfaCred != nil && mfaCred.IsVerified) || isMFAEnforced {
		return &AuthResult{
			MFARequired: true,
			User:        user,
		}, nil
	}

	return s.createSessionTokens(ctx, user, roles, orgRoles, false, ipAddress, userAgent)
}

func (s *AuthService) CompleteLoginWithMFA(ctx context.Context, userID, mfaCode, ipAddress, userAgent string) (*AuthResult, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, domain.ErrUserNotFound
	}

	mfaCred, err := s.mfaRepo.GetMFACredential(ctx, user.ID)
	if err != nil || mfaCred == nil {
		return nil, domain.ErrMFARequired
	}

	valid, err := s.mfaService.ValidateMFA(mfaCred, mfaCode, time.Now().UTC())
	if err != nil || !valid {
		return nil, domain.ErrInvalidMFACode
	}

	// If a recovery code was consumed, save updated hashed codes
	_ = s.mfaRepo.SaveMFACredential(ctx, mfaCred)

	roles, orgRoles, _ := s.userRepo.GetUserRoles(ctx, user.ID)
	if len(roles) == 0 {
		roles = []string{"USER"}
	}

	return s.createSessionTokens(ctx, user, roles, orgRoles, true, ipAddress, userAgent)
}

// RefreshToken handles Refresh Token Rotation and Reuse Detection (TC-SEC-005).
func (s *AuthService) RefreshToken(ctx context.Context, rawRefreshToken, ipAddress, userAgent string) (*AuthResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokenHash := hashToken(rawRefreshToken)
	session, err := s.sessionRepo.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil || session == nil {
		return nil, domain.ErrSessionExpired
	}

	// REUSE DETECTION: If token was already revoked, revoke the ENTIRE token family!
	if session.IsRevoked {
		_ = s.sessionRepo.RevokeFamily(ctx, session.TokenFamilyID)
		return nil, domain.ErrTokenReuseDetected
	}

	now := time.Now().UTC()
	if session.ExpiresAt.Before(now) {
		return nil, domain.ErrSessionExpired
	}

	user, err := s.userRepo.GetUserByID(ctx, session.UserID)
	if err != nil || user == nil || user.Status != domain.StatusActive {
		return nil, domain.ErrUnauthorized
	}

	// Generate new refresh token
	newRawRefreshToken, err := generateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	newHash := hashToken(newRawRefreshToken)

	newSession := &domain.AuthSession{
		ID:                uuid.New().String(),
		UserID:            user.ID,
		TokenFamilyID:     session.TokenFamilyID,
		RefreshTokenHash:  newHash,
		DeviceFingerprint: session.DeviceFingerprint,
		UserAgent:         userAgent,
		IPAddress:         ipAddress,
		IsRevoked:         false,
		ExpiresAt:         now.Add(14 * 24 * time.Hour),
		CreatedAt:         now,
	}

	// Atomically rotate: mark old session as revoked with replaced_by_hash and create new session
	session.IsRevoked = true
	session.ReplacedByHash = newHash
	session.RotatedAt = &now

	if err := s.sessionRepo.RotateSession(ctx, session.ID, newSession); err != nil {
		return nil, fmt.Errorf("failed to rotate session: %w", err)
	}

	roles, orgRoles, _ := s.userRepo.GetUserRoles(ctx, user.ID)
	if len(roles) == 0 {
		roles = []string{"USER"}
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, roles, orgRoles, true)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: newRawRefreshToken,
		ExpiresIn:    900,
		TokenType:    "Bearer",
		MFARequired:  false,
		User:         user,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	tokenHash := hashToken(rawRefreshToken)
	session, err := s.sessionRepo.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil || session == nil {
		return nil
	}
	return s.sessionRepo.RevokeFamily(ctx, session.TokenFamilyID)
}

func (s *AuthService) createSessionTokens(
	ctx context.Context,
	user *domain.User,
	roles []string,
	orgRoles map[string]string,
	mfaVerified bool,
	ipAddress, userAgent string,
) (*AuthResult, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, roles, orgRoles, mfaVerified)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	rawRefreshToken, err := generateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	now := time.Now().UTC()
	familyID := uuid.New().String()
	tokenHash := hashToken(rawRefreshToken)

	session := &domain.AuthSession{
		ID:                uuid.New().String(),
		UserID:            user.ID,
		TokenFamilyID:     familyID,
		RefreshTokenHash:  tokenHash,
		DeviceFingerprint: "",
		UserAgent:         userAgent,
		IPAddress:         ipAddress,
		IsRevoked:         false,
		ExpiresAt:         now.Add(14 * 24 * time.Hour),
		CreatedAt:         now,
	}

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to persist auth session: %w", err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    900,
		TokenType:    "Bearer",
		MFARequired:  false,
		User:         user,
	}, nil
}

func isMFAMandatoryForRoles(roles []string) bool {
	for _, r := range roles {
		if strings.HasPrefix(r, "ADMIN_") || r == "DEV_OWNER" || r == "DEV_RELEASE_MGR" || r == "DEV_FINANCE" {
			return true
		}
	}
	return false
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func generateRandomToken(bytesLen int) (string, error) {
	b := make([]byte, bytesLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
