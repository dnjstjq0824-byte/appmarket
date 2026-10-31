package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
)

// PostgresDeveloperRepository implements DeveloperRepository backed by PostgreSQL 16.
type PostgresDeveloperRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDeveloperRepository(pool *pgxpool.Pool) *PostgresDeveloperRepository {
	return &PostgresDeveloperRepository{pool: pool}
}

func (r *PostgresDeveloperRepository) CreateOrg(ctx context.Context, org *domain.DeveloperOrg) error {
	query := `
		INSERT INTO developer_orgs (
			id, org_slug, legal_name, display_name, org_type, country_code,
			support_email, verification_status, owner_user_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, 'CORPORATION', 'KR', 'support@' || $2 || '.com', $5, '00000000-0000-0000-0000-000000000001', $6, $7)
	`
	_, err := r.pool.Exec(
		ctx, query,
		org.ID, org.OrgSlug, org.OrgName, org.OrgName, org.VerificationStatus, org.CreatedAt, org.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create developer org: %w", err)
	}
	return nil
}

func (r *PostgresDeveloperRepository) GetOrgByID(ctx context.Context, orgID string) (*domain.DeveloperOrg, error) {
	query := `
		SELECT id, org_slug, display_name, verification_status, created_at, updated_at
		FROM developer_orgs
		WHERE id = $1
	`
	var org domain.DeveloperOrg
	err := r.pool.QueryRow(ctx, query, orgID).Scan(
		&org.ID, &org.OrgSlug, &org.OrgName, &org.VerificationStatus, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrOrgNotFound
		}
		return nil, fmt.Errorf("failed to get org by id: %w", err)
	}
	return &org, nil
}

func (r *PostgresDeveloperRepository) GetOrgBySlug(ctx context.Context, slug string) (*domain.DeveloperOrg, error) {
	query := `
		SELECT id, org_slug, display_name, verification_status, created_at, updated_at
		FROM developer_orgs
		WHERE org_slug = $1
	`
	var org domain.DeveloperOrg
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&org.ID, &org.OrgSlug, &org.OrgName, &org.VerificationStatus, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrOrgNotFound
		}
		return nil, fmt.Errorf("failed to get org by slug: %w", err)
	}
	return &org, nil
}

func (r *PostgresDeveloperRepository) UpdateOrgStatus(ctx context.Context, orgID string, status domain.OrgVerificationStatus) error {
	query := `
		UPDATE developer_orgs
		SET verification_status = $2, updated_at = NOW()
		WHERE id = $1
	`
	cmd, err := r.pool.Exec(ctx, query, orgID, status)
	if err != nil {
		return fmt.Errorf("failed to update org status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrOrgNotFound
	}
	return nil
}

func (r *PostgresDeveloperRepository) CreateVerification(ctx context.Context, v *domain.DeveloperVerification) error {
	query := `
		INSERT INTO developer_verifications (
			id, org_id, submitted_by, document_s3_vault_key, dns_txt_token, status, created_at
		) VALUES ($1, $2, '00000000-0000-0000-0000-000000000001', $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, v.ID, v.OrgID, v.DocURL, v.Domain, v.Status, v.SubmittedAt)
	if err != nil {
		return fmt.Errorf("failed to create verification: %w", err)
	}
	return nil
}

func (r *PostgresDeveloperRepository) GetLatestVerification(ctx context.Context, orgID string) (*domain.DeveloperVerification, error) {
	query := `
		SELECT id, org_id, 'BUSINESS_REGISTRATION', document_s3_vault_key, dns_txt_token, status, created_at
		FROM developer_verifications
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var v domain.DeveloperVerification
	err := r.pool.QueryRow(ctx, query, orgID).Scan(
		&v.ID, &v.OrgID, &v.DocType, &v.DocURL, &v.Domain, &v.Status, &v.SubmittedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest verification: %w", err)
	}
	return &v, nil
}

func (r *PostgresDeveloperRepository) AddMember(ctx context.Context, m *domain.OrgMember) error {
	query := `
		INSERT INTO developer_org_members (org_id, user_id, org_role, joined_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (org_id, user_id) DO UPDATE SET org_role = EXCLUDED.org_role
	`
	_, err := r.pool.Exec(ctx, query, m.OrgID, m.UserID, m.Role)
	if err != nil {
		return fmt.Errorf("failed to add org member: %w", err)
	}
	return nil
}

func (r *PostgresDeveloperRepository) GetMember(ctx context.Context, orgID, userID string) (*domain.OrgMember, error) {
	query := `
		SELECT org_id, user_id, org_role
		FROM developer_org_members
		WHERE org_id = $1 AND user_id = $2
	`
	var m domain.OrgMember
	err := r.pool.QueryRow(ctx, query, orgID, userID).Scan(&m.OrgID, &m.UserID, &m.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrMemberNotFound
		}
		return nil, fmt.Errorf("failed to get org member: %w", err)
	}
	return &m, nil
}

func (r *PostgresDeveloperRepository) ListMembers(ctx context.Context, orgID string) ([]domain.OrgMember, error) {
	query := `
		SELECT org_id, user_id, org_role
		FROM developer_org_members
		WHERE org_id = $1
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list org members: %w", err)
	}
	defer rows.Close()

	var members []domain.OrgMember
	for rows.Next() {
		var m domain.OrgMember
		if err := rows.Scan(&m.OrgID, &m.UserID, &m.Role); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *PostgresDeveloperRepository) CreateAPIKey(ctx context.Context, k *domain.APIKey) error {
	prefix := ""
	if len(k.KeyHash) >= 12 {
		prefix = k.KeyHash[:12]
	} else {
		prefix = "bn_live_test"
	}
	query := `
		INSERT INTO developer_api_keys (
			id, org_id, key_name, key_prefix, key_sha256, scopes, expires_at, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, '00000000-0000-0000-0000-000000000001', $8)
	`
	_, err := r.pool.Exec(
		ctx, query,
		k.ID, k.OrgID, k.Name, prefix, k.KeyHash, k.Scopes, k.ExpiresAt, k.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create api key: %w", err)
	}
	return nil
}

func (r *PostgresDeveloperRepository) GetAPIKeyByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	query := `
		SELECT id, org_id, key_name, scopes, expires_at, revoked_at IS NOT NULL, created_at
		FROM developer_api_keys
		WHERE key_sha256 = $1
	`
	var k domain.APIKey
	err := r.pool.QueryRow(ctx, query, keyHash).Scan(
		&k.ID, &k.OrgID, &k.Name, &k.Scopes, &k.ExpiresAt, &k.Revoked, &k.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrAPIKeyNotFound
		}
		return nil, fmt.Errorf("failed to get api key by hash: %w", err)
	}
	k.KeyHash = keyHash
	return &k, nil
}

func (r *PostgresDeveloperRepository) RevokeAPIKey(ctx context.Context, keyID string) error {
	query := `
		UPDATE developer_api_keys
		SET revoked_at = $2
		WHERE id = $1
	`
	cmd, err := r.pool.Exec(ctx, query, keyID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to revoke api key: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}

func (r *PostgresDeveloperRepository) ListAPIKeys(ctx context.Context, orgID string) ([]domain.APIKey, error) {
	query := `
		SELECT id, org_id, key_name, scopes, expires_at, revoked_at IS NOT NULL, created_at
		FROM developer_api_keys
		WHERE org_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}
	defer rows.Close()

	var keys []domain.APIKey
	for rows.Next() {
		var k domain.APIKey
		if err := rows.Scan(&k.ID, &k.OrgID, &k.Name, &k.Scopes, &k.ExpiresAt, &k.Revoked, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

