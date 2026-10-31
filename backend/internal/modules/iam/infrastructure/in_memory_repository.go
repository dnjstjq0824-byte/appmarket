package infrastructure

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
)

// InMemoryUserRepository provides thread-safe in-memory storage for users and roles.
type InMemoryUserRepository struct {
	mu          sync.RWMutex
	users       map[string]*domain.User
	roles       map[string][]string
	orgRoles    map[string]map[string]string
	blindIdxMap map[string]string // blindIndex -> userID
}

// NewInMemoryUserRepository creates and seeds the in-memory user repository.
func NewInMemoryUserRepository() *InMemoryUserRepository {
	r := &InMemoryUserRepository{
		users:       make(map[string]*domain.User),
		roles:       make(map[string][]string),
		orgRoles:    make(map[string]map[string]string),
		blindIdxMap: make(map[string]string),
	}
	r.seedUsers()
	return r
}

func (r *InMemoryUserRepository) seedUsers() {
	now := time.Now().UTC()

	// 1. Admin Super User
	adminID := "usr-admin-0000-0000-000000000001"
	adminBlind := ComputeBlindIndex("admin@appmarket.local", "seed-hmac-salt")
	admin := &domain.User{
		ID:                  adminID,
		Email:               "admin@appmarket.local",
		EmailBlindIndex:     adminBlind,
		PasswordHash:        "$argon2id$v=19$m=65536,t=3,p=4$seedAdminSalt$eZz8v...",
		DisplayName:         "Platform Administrator",
		Status:              domain.StatusActive,
		FailedLoginAttempts: 0,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	r.users[adminID] = admin
	r.blindIdxMap[adminBlind] = adminID
	r.roles[adminID] = []string{"ADMIN_SUPER"}

	// 2. Verified Developer
	devID := "usr-dev-0000-0000-000000000002"
	devBlind := ComputeBlindIndex("dev@omniwork.com", "seed-hmac-salt")
	dev := &domain.User{
		ID:                  devID,
		Email:               "dev@omniwork.com",
		EmailBlindIndex:     devBlind,
		PasswordHash:        "$argon2id$v=19$m=65536,t=3,p=4$seedDevSalt$eZz8v...",
		DisplayName:         "OmniWork Dev Lead",
		Status:              domain.StatusActive,
		FailedLoginAttempts: 0,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	r.users[devID] = dev
	r.blindIdxMap[devBlind] = devID
	r.roles[devID] = []string{"DEV_OWNER"}
	r.orgRoles[devID] = map[string]string{"org-omni-01": "DEV_OWNER"}

	// 3. Consumer End-User
	consumerID := "usr-user-0000-0000-000000000003"
	consumerBlind := ComputeBlindIndex("user@test.local", "seed-hmac-salt")
	consumer := &domain.User{
		ID:                  consumerID,
		Email:               "user@test.local",
		EmailBlindIndex:     consumerBlind,
		PasswordHash:        "$argon2id$v=19$m=65536,t=3,p=4$seedUserSalt$eZz8v...",
		DisplayName:         "Test Consumer",
		Status:              domain.StatusActive,
		FailedLoginAttempts: 0,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	r.users[consumerID] = consumer
	r.blindIdxMap[consumerBlind] = consumerID
	r.roles[consumerID] = []string{"USER"}
}

// ComputeBlindIndex produces consistent sha256 blind indices.
func ComputeBlindIndex(email, salt string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%s", salt, email)))
	return hex.EncodeToString(h.Sum(nil))
}

func (r *InMemoryUserRepository) CreateUser(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.blindIdxMap[u.EmailBlindIndex]; exists {
		return domain.ErrUserAlreadyExists
	}
	r.users[u.ID] = u
	r.blindIdxMap[u.EmailBlindIndex] = u.ID
	if _, ok := r.roles[u.ID]; !ok {
		r.roles[u.ID] = []string{"USER"}
	}
	return nil
}

func (r *InMemoryUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	copyU := *u
	return &copyU, nil
}

func (r *InMemoryUserRepository) GetUserByEmailBlindIndex(ctx context.Context, blindIndex string) (*domain.User, error) {
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

func (r *InMemoryUserRepository) UpdateUserStatus(ctx context.Context, id string, status domain.UserStatus, failedAttempts int, lockedUntil *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.users[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.Status = status
	u.FailedLoginAttempts = failedAttempts
	u.LockedUntil = lockedUntil
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *InMemoryUserRepository) GetUserRoles(ctx context.Context, userID string) ([]string, map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roles := r.roles[userID]
	if roles == nil {
		roles = []string{"USER"}
	}
	orgRoles := r.orgRoles[userID]
	if orgRoles == nil {
		orgRoles = make(map[string]string)
	}
	return roles, orgRoles, nil
}

// InMemorySessionRepository provides thread-safe session tracking.
type InMemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*domain.AuthSession
	byHash   map[string]string
	byFamily map[string][]string
}

func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions: make(map[string]*domain.AuthSession),
		byHash:   make(map[string]string),
		byFamily: make(map[string][]string),
	}
}

func (r *InMemorySessionRepository) CreateSession(ctx context.Context, session *domain.AuthSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[session.ID] = session
	r.byHash[session.RefreshTokenHash] = session.ID
	r.byFamily[session.TokenFamilyID] = append(r.byFamily[session.TokenFamilyID], session.ID)
	return nil
}

func (r *InMemorySessionRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byHash[tokenHash]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	s := r.sessions[id]
	copyS := *s
	return &copyS, nil
}

func (r *InMemorySessionRepository) RotateSession(ctx context.Context, oldSessionID string, newSession *domain.AuthSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	old, ok := r.sessions[oldSessionID]
	if !ok {
		return domain.ErrSessionNotFound
	}
	now := time.Now().UTC()
	old.RotatedAt = &now
	old.ReplacedByHash = newSession.RefreshTokenHash

	r.sessions[newSession.ID] = newSession
	r.byHash[newSession.RefreshTokenHash] = newSession.ID
	r.byFamily[newSession.TokenFamilyID] = append(r.byFamily[newSession.TokenFamilyID], newSession.ID)
	return nil
}

func (r *InMemorySessionRepository) RevokeFamily(ctx context.Context, tokenFamilyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids := r.byFamily[tokenFamilyID]
	for _, id := range ids {
		if s, ok := r.sessions[id]; ok {
			s.IsRevoked = true
		}
	}
	return nil
}

// InMemoryMFARepository provides thread-safe in-memory MFA storage.
type InMemoryMFARepository struct {
	mu    sync.RWMutex
	creds map[string]*domain.MFACredential
}

func NewInMemoryMFARepository() *InMemoryMFARepository {
	return &InMemoryMFARepository{
		creds: make(map[string]*domain.MFACredential),
	}
}

func (r *InMemoryMFARepository) GetMFACredential(ctx context.Context, userID string) (*domain.MFACredential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.creds[userID]
	if !ok {
		return nil, nil
	}
	copyC := *c
	return &copyC, nil
}

func (r *InMemoryMFARepository) SaveMFACredential(ctx context.Context, cred *domain.MFACredential) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.creds[cred.UserID] = cred
	return nil
}

