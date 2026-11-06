package infrastructure

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
)

// InMemoryIdentityRepository stores federated identities in-process. It backs
// the Postgres repository when no database handle is configured.
type InMemoryIdentityRepository struct {
	mu         sync.RWMutex
	byProvider map[string]*domain.UserIdentity // key: provider + ":" + subject
}

func NewInMemoryIdentityRepository() *InMemoryIdentityRepository {
	return &InMemoryIdentityRepository{byProvider: make(map[string]*domain.UserIdentity)}
}

func identityKey(provider, subject string) string { return provider + ":" + subject }

func (r *InMemoryIdentityRepository) FindUserIDByProvider(ctx context.Context, provider, subject string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if id, ok := r.byProvider[identityKey(provider, subject)]; ok {
		return id.UserID, nil
	}
	return "", domain.ErrUserNotFound
}

func (r *InMemoryIdentityRepository) LinkIdentity(ctx context.Context, identity *domain.UserIdentity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if identity.ID == "" {
		identity.ID = uuid.New().String()
	}
	if identity.CreatedAt.IsZero() {
		identity.CreatedAt = time.Now().UTC()
	}
	r.byProvider[identityKey(identity.Provider, identity.ProviderSubject)] = identity
	return nil
}

// PostgresIdentityRepository persists user_identities rows in PostgreSQL 16.
type PostgresIdentityRepository struct {
	db       *pgxpool.Pool
	fallback *InMemoryIdentityRepository
}

func NewPostgresIdentityRepository(db *pgxpool.Pool) *PostgresIdentityRepository {
	return &PostgresIdentityRepository{db: db, fallback: NewInMemoryIdentityRepository()}
}

func (r *PostgresIdentityRepository) FindUserIDByProvider(ctx context.Context, provider, subject string) (string, error) {
	if r.db == nil {
		return r.fallback.FindUserIDByProvider(ctx, provider, subject)
	}
	var userID string
	err := r.db.QueryRow(ctx,
		`SELECT user_id FROM user_identities WHERE provider = $1 AND provider_subject = $2`,
		provider, subject).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrUserNotFound
		}
		return "", err
	}
	return userID, nil
}

func (r *PostgresIdentityRepository) LinkIdentity(ctx context.Context, identity *domain.UserIdentity) error {
	if r.db == nil {
		return r.fallback.LinkIdentity(ctx, identity)
	}
	if identity.ID == "" {
		identity.ID = uuid.New().String()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_identities (id, user_id, provider, provider_subject, provider_email)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider, provider_subject) DO UPDATE SET provider_email = EXCLUDED.provider_email`,
		identity.ID, identity.UserID, identity.Provider, identity.ProviderSubject, nullString(identity.ProviderEmail),
	)
	return err
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
