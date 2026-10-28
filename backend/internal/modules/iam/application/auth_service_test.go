package application_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
)

type mockUserRepo struct {
	mu          sync.RWMutex
	users       map[string]*domain.User
	roles       map[string][]string
	orgRoles    map[string]map[string]string
	blindIdxMap map[string]string // blindIndex -> userID
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:       make(map[string]*domain.User),
		roles:       make(map[string][]string),
		orgRoles:    make(map[string]map[string]string),
		blindIdxMap: make(map[string]string),
	}
}

func (r *mockUserRepo) CreateUser(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	r.blindIdxMap[u.EmailBlindIndex] = u.ID
	return nil
}

func (r *mockUserRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	copyU := *u
	return &copyU, nil
}

func (r *mockUserRepo) GetUserByEmailBlindIndex(ctx context.Context, blindIndex string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.blindIdxMap[blindIndex]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	u := r.users[id]
	copyU := *u
	return &copyU, nil
}

func (r *mockUserRepo) UpdateUserStatus(ctx context.Context, id string, status domain.UserStatus, failedAttempts int, lockedUntil *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.Status = status
	u.FailedLoginAttempts = failedAttempts
	u.LockedUntil = lockedUntil
	return nil
}

func (r *mockUserRepo) GetUserRoles(ctx context.Context, userID string) ([]string, map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	roles := r.roles[userID]
	orgRoles := r.orgRoles[userID]
	if roles == nil {
		roles = []string{"USER"}
	}
	if orgRoles == nil {
		orgRoles = make(map[string]string)
	}
	return roles, orgRoles, nil
}

type mockSessionRepo struct {
	mu       sync.RWMutex
	sessions map[string]*domain.AuthSession // sessionID -> session
	byHash   map[string]string              // tokenHash -> sessionID
	byFamily map[string][]string            // familyID -> []sessionID
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions: make(map[string]*domain.AuthSession),
		byHash:   make(map[string]string),
		byFamily: make(map[string][]string),
	}
}

func (r *mockSessionRepo) CreateSession(ctx context.Context, s *domain.AuthSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = s
	r.byHash[s.RefreshTokenHash] = s.ID
	r.byFamily[s.TokenFamilyID] = append(r.byFamily[s.TokenFamilyID], s.ID)
	return nil
}

func (r *mockSessionRepo) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byHash[tokenHash]
	if !ok {
		return nil, domain.ErrSessionExpired
	}
	s := r.sessions[id]
	copyS := *s
	return &copyS, nil
}

func (r *mockSessionRepo) RotateSession(ctx context.Context, oldSessionID string, newSession *domain.AuthSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.sessions[oldSessionID]
	if !ok {
		return domain.ErrSessionExpired
	}
	now := time.Now().UTC()
	old.IsRevoked = true
	old.ReplacedByHash = newSession.RefreshTokenHash
	old.RotatedAt = &now

	r.sessions[newSession.ID] = newSession
	r.byHash[newSession.RefreshTokenHash] = newSession.ID
	r.byFamily[newSession.TokenFamilyID] = append(r.byFamily[newSession.TokenFamilyID], newSession.ID)
	return nil
}

func (r *mockSessionRepo) RevokeFamily(ctx context.Context, tokenFamilyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range r.byFamily[tokenFamilyID] {
		if s, ok := r.sessions[id]; ok {
			s.IsRevoked = true
		}
	}
	return nil
}

type mockMFARepo struct {
	mu    sync.RWMutex
	creds map[string]*domain.MFACredential
}

func newMockMFARepo() *mockMFARepo {
	return &mockMFARepo{creds: make(map[string]*domain.MFACredential)}
}

func (r *mockMFARepo) GetMFACredential(ctx context.Context, userID string) (*domain.MFACredential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.creds[userID]
	if !ok {
		return nil, nil
	}
	copyC := *c
	return &copyC, nil
}

func (r *mockMFARepo) SaveMFACredential(ctx context.Context, cred *domain.MFACredential) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.creds[cred.UserID] = cred
	return nil
}

func setupTestAuthService(t *testing.T) (*application.AuthService, *mockUserRepo, *mockSessionRepo, *mockMFARepo, *application.JWTManager) {
	uRepo := newMockUserRepo()
	sRepo := newMockSessionRepo()
	mRepo := newMockMFARepo()

	// Use smaller argon2 parameters in test for speed
	hasher := application.NewPasswordHasher(application.Argon2idParams{
		Memory:      1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})

	jwtMgr, err := application.NewJWTManager(nil, "https://api.appmarket.bookingnetwork.internal")
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}

	mfaSvc := application.NewMFAService("BN App Market")

	svc := application.NewAuthService(uRepo, sRepo, mRepo, hasher, jwtMgr, mfaSvc, application.AuthServiceConfig{
		HMACSecret:      "test-secret-key-32-chars-long!",
		MaxFailedLogin:  5,
		LockoutDuration: 15 * time.Minute,
	})

	return svc, uRepo, sRepo, mRepo, jwtMgr
}

// TC-SEC-005: Refresh Token Rotation & Reuse Detection
func TestRefreshTokenRotationAndReuseDetection(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _ := setupTestAuthService(t)

	// 1. Register and Login
	_, err := svc.Register(ctx, application.RegisterInput{
		Email:       "test@example.com",
		Password:    "SecureP@ssword123",
		DisplayName: "Test Developer",
	})
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	loginRes, err := svc.Login(ctx, "test@example.com", "SecureP@ssword123", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	initialRefreshToken := loginRes.RefreshToken
	if initialRefreshToken == "" {
		t.Fatalf("expected non-empty refresh token")
	}

	// 2. Legitimate Refresh Token Rotation (Token 1 -> Token 2)
	rotateRes1, err := svc.RefreshToken(ctx, initialRefreshToken, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("first token rotation failed: %v", err)
	}
	secondRefreshToken := rotateRes1.RefreshToken
	if secondRefreshToken == initialRefreshToken {
		t.Fatalf("rotated token must be distinct from original token")
	}

	// 3. Legitimate Refresh Token Rotation (Token 2 -> Token 3)
	rotateRes2, err := svc.RefreshToken(ctx, secondRefreshToken, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("second token rotation failed: %v", err)
	}
	thirdRefreshToken := rotateRes2.RefreshToken

	// 4. ATTEMPT REUSE: An attacker attempts to submit Token 1 (which was already rotated)
	reuseRes, err := svc.RefreshToken(ctx, initialRefreshToken, "192.168.1.100", "attacker-agent")
	if err == nil {
		t.Fatalf("expected token reuse to be detected and rejected, but got success: %v", reuseRes)
	}
	if err != domain.ErrTokenReuseDetected {
		t.Fatalf("expected ErrTokenReuseDetected, got %v", err)
	}

	// 5. ENTIRE FAMILY REVOKED: Even the latest valid token (Token 3) must now be rejected because the family was revoked!
	_, err = svc.RefreshToken(ctx, thirdRefreshToken, "127.0.0.1", "test-agent")
	if err == nil {
		t.Fatalf("expected Token 3 to be rejected following family revocation, but succeeded")
	}
}

// TC-API-001: Auth lifecycle, password hashing and 5-consecutive brute force lockout
func TestAuthLifecycleAndBruteForceLockout(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _ := setupTestAuthService(t)

	// 1. Register
	user, err := svc.Register(ctx, application.RegisterInput{
		Email:       "alice@example.com",
		Password:    "SuperStrongPass!99",
		DisplayName: "Alice Dev",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if user.Status != domain.StatusActive {
		t.Fatalf("expected active user, got %s", user.Status)
	}

	// 2. Duplicate registration rejection
	_, err = svc.Register(ctx, application.RegisterInput{
		Email:       "alice@example.com",
		Password:    "AnotherPass123!",
		DisplayName: "Alice Duplicate",
	})
	if err != domain.ErrUserAlreadyExists {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}

	// 3. Failed logins (4 times -> still invalid credentials)
	for i := 1; i <= 4; i++ {
		_, err = svc.Login(ctx, "alice@example.com", "WrongPassword!", "127.0.0.1", "agent")
		if err != domain.ErrInvalidCredentials {
			t.Fatalf("attempt %d: expected ErrInvalidCredentials, got %v", i, err)
		}
	}

	// 4. 5th failed login -> locks account for 15 minutes
	_, err = svc.Login(ctx, "alice@example.com", "WrongPassword!", "127.0.0.1", "agent")
	if err != domain.ErrAccountLocked {
		t.Fatalf("5th attempt: expected ErrAccountLocked, got %v", err)
	}

	// 5. Subsequent attempt even with correct password must be rejected while locked
	_, err = svc.Login(ctx, "alice@example.com", "SuperStrongPass!99", "127.0.0.1", "agent")
	if err != domain.ErrAccountLocked {
		t.Fatalf("while locked: expected ErrAccountLocked, got %v", err)
	}
}

// TOTP MFA verification test
func TestMFAServiceVerification(t *testing.T) {
	mfaSvc := application.NewMFAService("BN App Market")

	secret, qrURL, recoveryCodes, hashedCodes, err := mfaSvc.GenerateSecret("user@example.com")
	if err != nil {
		t.Fatalf("failed to generate secret: %v", err)
	}

	if !strings.Contains(qrURL, "otpauth://totp/") {
		t.Fatalf("invalid qr url: %s", qrURL)
	}
	if len(recoveryCodes) != 10 || len(hashedCodes) != 10 {
		t.Fatalf("expected 10 recovery codes")
	}

	// Compute TOTP code at current time
	now := time.Now().UTC()
	code, err := mfaSvc.ComputeTOTP(secret, now)
	if err != nil {
		t.Fatalf("failed to compute totp: %v", err)
	}

	// Verify TOTP code
	if !mfaSvc.VerifyTOTP(secret, code, now) {
		t.Fatalf("expected valid totp verification")
	}

	// Verify recovery code consumption
	valid, remaining := mfaSvc.VerifyAndConsumeRecoveryCode(recoveryCodes[0], hashedCodes)
	if !valid {
		t.Fatalf("expected recovery code to be valid")
	}
	if len(remaining) != 9 {
		t.Fatalf("expected 9 remaining recovery codes, got %d", len(remaining))
	}

	// Reusing same recovery code must fail
	validAgain, _ := mfaSvc.VerifyAndConsumeRecoveryCode(recoveryCodes[0], remaining)
	if validAgain {
		t.Fatalf("reusing already consumed recovery code must fail")
	}
}

