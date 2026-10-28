package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
)

// APIKeyService handles CI/CD keys.
type APIKeyService struct {
	repo DeveloperRepository
}

func NewAPIKeyService(repo DeveloperRepository) *APIKeyService {
	return &APIKeyService{repo: repo}
}

// IssueAPIKey creates a new key; returns raw key (shown once) and record.
func (s *APIKeyService) IssueAPIKey(ctx context.Context, orgID string, appID *string, name string, scopes []string, ttlDays int) (string, *domain.APIKey, error) {
	if ttlDays <= 0 {
		ttlDays = 90
	}
	raw, err := generateRandomKey(32)
	if err != nil {
		return "", nil, err
	}
	hash := sha256.Sum256([]byte(raw))
	keyHash := hex.EncodeToString(hash[:])
	now := time.Now().UTC()
	k := &domain.APIKey{
		ID:        uuid.New().String(),
		OrgID:     orgID,
		AppID:     appID,
		Name:      name,
		KeyHash:   keyHash,
		Scopes:    scopes,
		ExpiresAt: now.Add(time.Duration(ttlDays) * 24 * time.Hour),
		Revoked:   false,
		CreatedAt: now,
	}
	if err := s.repo.CreateAPIKey(ctx, k); err != nil {
		return "", nil, err
	}
	return raw, k, nil
}

// ValidateAPIKey validates a raw key and returns the record.
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, raw string) (*domain.APIKey, error) {
	hash := sha256.Sum256([]byte(raw))
	keyHash := hex.EncodeToString(hash[:])
	k, err := s.repo.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, domain.ErrAPIKeyNotFound
	}
	if !k.IsValid(time.Now().UTC()) {
		return nil, domain.ErrAPIKeyNotFound
	}
	return k, nil
}

// RevokeAPIKey revokes a key.
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, keyID string) error {
	return s.repo.RevokeAPIKey(ctx, keyID)
}

func generateRandomKey(bytesLen int) (string, error) {
	b := make([]byte, bytesLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
