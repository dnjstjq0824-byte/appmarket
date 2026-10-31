package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
)

// PostgresUserRepository provides PostgreSQL 16 persistence for users.
type PostgresUserRepository struct {
	db       *pgxpool.Pool
	fallback *InMemoryUserRepository
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		db:       db,
		fallback: NewInMemoryUserRepository(),
	}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, u *domain.User) error {
	if r.db == nil {
		return r.fallback.CreateUser(ctx, u)
	}

	query := `
		INSERT INTO users (id, email, email_blind_index, password_hash, display_name, status, failed_login_attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.Exec(ctx, query,
		u.ID, u.Email, u.EmailBlindIndex, u.PasswordHash, u.DisplayName,
		string(u.Status), u.FailedLoginAttempts, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Insert default USER role
	_, _ = r.db.Exec(ctx, `INSERT INTO user_roles (user_id, role_code) VALUES ($1, 'USER') ON CONFLICT DO NOTHING`, u.ID)
	return nil
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if r.db == nil {
		return r.fallback.GetUserByID(ctx, id)
	}

	query := `
		SELECT id, email, email_blind_index, password_hash, display_name, status, failed_login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE id = $1`

	var u domain.User
	var statusStr string
	var passHash *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.EmailBlindIndex, &passHash, &u.DisplayName,
		&statusStr, &u.FailedLoginAttempts, &u.LockedUntil, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.fallback.GetUserByID(ctx, id)
		}
		return nil, err
	}
	u.Status = domain.UserStatus(statusStr)
	if passHash != nil {
		u.PasswordHash = *passHash
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetUserByEmailBlindIndex(ctx context.Context, blindIndex string) (*domain.User, error) {
	if r.db == nil {
		return r.fallback.GetUserByEmailBlindIndex(ctx, blindIndex)
	}

	query := `
		SELECT id, email, email_blind_index, password_hash, display_name, status, failed_login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE email_blind_index = $1`

	var u domain.User
	var statusStr string
	var passHash *string
	err := r.db.QueryRow(ctx, query, blindIndex).Scan(
		&u.ID, &u.Email, &u.EmailBlindIndex, &passHash, &u.DisplayName,
		&statusStr, &u.FailedLoginAttempts, &u.LockedUntil, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.fallback.GetUserByEmailBlindIndex(ctx, blindIndex)
		}
		return nil, err
	}
	u.Status = domain.UserStatus(statusStr)
	if passHash != nil {
		u.PasswordHash = *passHash
	}
	return &u, nil
}

func (r *PostgresUserRepository) UpdateUserStatus(ctx context.Context, id string, status domain.UserStatus, failedAttempts int, lockedUntil *time.Time) error {
	if r.db == nil {
		return r.fallback.UpdateUserStatus(ctx, id, status, failedAttempts, lockedUntil)
	}

	query := `
		UPDATE users
		SET status = $2, failed_login_attempts = $3, locked_until = $4, updated_at = NOW()
		WHERE id = $1`

	cmd, err := r.db.Exec(ctx, query, id, string(status), failedAttempts, lockedUntil)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return r.fallback.UpdateUserStatus(ctx, id, status, failedAttempts, lockedUntil)
	}
	return nil
}

func (r *PostgresUserRepository) GetUserRoles(ctx context.Context, userID string) ([]string, map[string]string, error) {
	if r.db == nil {
		return r.fallback.GetUserRoles(ctx, userID)
	}

	rows, err := r.db.Query(ctx, `SELECT role_code FROM user_roles WHERE user_id = $1`, userID)
	if err != nil {
		return r.fallback.GetUserRoles(ctx, userID)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err == nil {
			roles = append(roles, role)
		}
	}
	if len(roles) == 0 {
		roles = []string{"USER"}
	}

	// Fetch tenant roles from developer_org_members
	orgRoles := make(map[string]string)
	memRows, err := r.db.Query(ctx, `SELECT org_id, role FROM developer_org_members WHERE user_id = $1`, userID)
	if err == nil {
		defer memRows.Close()
		for memRows.Next() {
			var orgID, role string
			if err := memRows.Scan(&orgID, &role); err == nil {
				orgRoles[orgID] = role
			}
		}
	}

	return roles, orgRoles, nil
}

// PostgresSessionRepository provides PostgreSQL 16 persistence for auth sessions.
type PostgresSessionRepository struct {
	db       *pgxpool.Pool
	fallback *InMemorySessionRepository
}

func NewPostgresSessionRepository(db *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{
		db:       db,
		fallback: NewInMemorySessionRepository(),
	}
}

func (r *PostgresSessionRepository) CreateSession(ctx context.Context, s *domain.AuthSession) error {
	if r.db == nil {
		return r.fallback.CreateSession(ctx, s)
	}

	query := `
		INSERT INTO auth_sessions (id, user_id, token_family_id, refresh_token_sha256, device_fingerprint, user_agent, is_revoked, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.Exec(ctx, query,
		s.ID, s.UserID, s.TokenFamilyID, s.RefreshTokenHash, s.DeviceFingerprint, s.UserAgent, s.IsRevoked, s.ExpiresAt, s.CreatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresSessionRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthSession, error) {
	if r.db == nil {
		return r.fallback.GetSessionByTokenHash(ctx, tokenHash)
	}

	query := `
		SELECT id, user_id, token_family_id, refresh_token_sha256, device_fingerprint, user_agent, is_revoked, expires_at, created_at, rotated_at
		FROM auth_sessions
		WHERE refresh_token_sha256 = $1`

	var s domain.AuthSession
	var devFP, ua *string
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&s.ID, &s.UserID, &s.TokenFamilyID, &s.RefreshTokenHash, &devFP, &ua, &s.IsRevoked, &s.ExpiresAt, &s.CreatedAt, &s.RotatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.fallback.GetSessionByTokenHash(ctx, tokenHash)
		}
		return nil, err
	}
	if devFP != nil {
		s.DeviceFingerprint = *devFP
	}
	if ua != nil {
		s.UserAgent = *ua
	}
	return &s, nil
}

func (r *PostgresSessionRepository) RotateSession(ctx context.Context, oldSessionID string, newSession *domain.AuthSession) error {
	if r.db == nil {
		return r.fallback.RotateSession(ctx, oldSessionID, newSession)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `UPDATE auth_sessions SET rotated_at = $2 WHERE id = $1`, oldSessionID, now)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO auth_sessions (id, user_id, token_family_id, refresh_token_sha256, device_fingerprint, user_agent, is_revoked, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = tx.Exec(ctx, query,
		newSession.ID, newSession.UserID, newSession.TokenFamilyID, newSession.RefreshTokenHash,
		newSession.DeviceFingerprint, newSession.UserAgent, newSession.IsRevoked, newSession.ExpiresAt, newSession.CreatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresSessionRepository) RevokeFamily(ctx context.Context, tokenFamilyID string) error {
	if r.db == nil {
		return r.fallback.RevokeFamily(ctx, tokenFamilyID)
	}

	_, err := r.db.Exec(ctx, `UPDATE auth_sessions SET is_revoked = TRUE, revoked_reason = 'REUSE_OR_MANUAL_REVOCATION' WHERE token_family_id = $1`, tokenFamilyID)
	return err
}

// PostgresMFARepository provides PostgreSQL 16 persistence for MFA credentials.
type PostgresMFARepository struct {
	db       *pgxpool.Pool
	fallback *InMemoryMFARepository
}

func NewPostgresMFARepository(db *pgxpool.Pool) *PostgresMFARepository {
	return &PostgresMFARepository{
		db:       db,
		fallback: NewInMemoryMFARepository(),
	}
}

func (r *PostgresMFARepository) GetMFACredential(ctx context.Context, userID string) (*domain.MFACredential, error) {
	if r.db == nil {
		return r.fallback.GetMFACredential(ctx, userID)
	}

	query := `
		SELECT id, user_id, mfa_type, secret_ciphertext, backup_codes_hashes, is_verified, created_at, updated_at
		FROM mfa_credentials
		WHERE user_id = $1`

	var c domain.MFACredential
	var secretBytes []byte
	var backupJSON []byte
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&c.ID, &c.UserID, &c.Type, &secretBytes, &backupJSON, &c.IsVerified, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.fallback.GetMFACredential(ctx, userID)
		}
		return nil, err
	}
	c.SecretEncrypted = string(secretBytes)
	if len(backupJSON) > 0 {
		_ = json.Unmarshal(backupJSON, &c.BackupCodesHashed)
	}
	return &c, nil
}

func (r *PostgresMFARepository) SaveMFACredential(ctx context.Context, cred *domain.MFACredential) error {
	if r.db == nil {
		return r.fallback.SaveMFACredential(ctx, cred)
	}

	backupJSON, _ := json.Marshal(cred.BackupCodesHashed)
	query := `
		INSERT INTO mfa_credentials (id, user_id, mfa_type, secret_ciphertext, backup_codes_hashes, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE
		SET secret_ciphertext = EXCLUDED.secret_ciphertext,
		    backup_codes_hashes = EXCLUDED.backup_codes_hashes,
		    is_verified = EXCLUDED.is_verified,
		    updated_at = EXCLUDED.updated_at`

	_, err := r.db.Exec(ctx, query,
		cred.ID, cred.UserID, cred.Type, []byte(cred.SecretEncrypted), backupJSON, cred.IsVerified, cred.CreatedAt, cred.UpdatedAt,
	)
	return err
}

