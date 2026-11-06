package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	iamApp "github.com/bookingnetwork/appmarket/backend/internal/modules/iam/application"
)

func requiredEnv(name string) string {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		slog.Error("required environment variable missing", "name", name)
		os.Exit(2)
	}
	return v
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	databaseURL := requiredEnv("DATABASE_URL")
	hmacSecret := requiredEnv("IAM_HMAC_SECRET")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if parts := strings.Split(databaseURL, "@"); len(parts) > 1 {
		logger.Info("connecting to postgres for seeding", "host", parts[len(parts)-1])
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		logger.Error("failed to create pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("database ping failed", "error", err)
		os.Exit(1)
	}

	hasher := iamApp.NewPasswordHasher()

	tx, err := pool.Begin(ctx)
	if err != nil {
		logger.Error("failed to start tx", "error", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	exec := func(query string, args ...any) {
		if _, err := tx.Exec(ctx, query, args...); err != nil {
			logger.Error("seed statement failed", "error", err)
			os.Exit(1)
		}
	}

	// 1. Seed Super Admin User (schema: users + user_roles, no tenants table)
	adminPassHash, err := hasher.HashPassword("AdminSecurePassword2026!")
	if err != nil {
		logger.Error("failed to hash admin password", "error", err)
		os.Exit(1)
	}
	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	adminEmail := "admin@bookingnetwork.com"
	exec(`
		INSERT INTO users (id, email, email_blind_index, password_hash, display_name, country_code, preferred_locale, preferred_currency, status)
		VALUES ($1, $2, $3, $4, 'Platform Administrator', 'KR', 'ko-KR', 'KRW', 'ACTIVE')
		ON CONFLICT (email) DO NOTHING;
	`, adminID, adminEmail, iamApp.ComputeBlindIndex(adminEmail, hmacSecret), adminPassHash)
	exec(`
		INSERT INTO user_roles (user_id, role_code)
		SELECT id, 'ADMIN_SUPER' FROM users WHERE email = $1
		ON CONFLICT (user_id, role_code) DO NOTHING;
	`, adminEmail)

	// 2. Seed Developer User (owner of the demo org below)
	devPassHash, err := hasher.HashPassword("DevSecurePassword2026!")
	if err != nil {
		logger.Error("failed to hash developer password", "error", err)
		os.Exit(1)
	}
	devUserID := uuid.MustParse("00000000-0000-0000-0000-000000000021")
	devEmail := "dev@acmecorp.com"
	exec(`
		INSERT INTO users (id, email, email_blind_index, password_hash, display_name, country_code, preferred_locale, preferred_currency, status)
		VALUES ($1, $2, $3, $4, 'Acme Developer', 'KR', 'ko-KR', 'KRW', 'ACTIVE')
		ON CONFLICT (email) DO NOTHING;
	`, devUserID, devEmail, iamApp.ComputeBlindIndex(devEmail, hmacSecret), devPassHash)
	exec(`
		INSERT INTO user_roles (user_id, role_code)
		SELECT id, 'USER' FROM users WHERE email = $1
		ON CONFLICT (user_id, role_code) DO NOTHING;
	`, devEmail)

	// 3. Seed Verified Developer Org
	devOrgID := uuid.MustParse("00000000-0000-0000-0000-000000000020")
	exec(`
		INSERT INTO developer_orgs (id, org_slug, legal_name, display_name, org_type, country_code, support_email, verification_status, commission_tier_code, owner_user_id)
		VALUES ($1, 'acme-technologies', 'Acme Technologies Inc.', 'Acme', 'CORPORATION', 'KR', $2, 'VERIFIED', 'STANDARD_TIER', $3)
		ON CONFLICT (id) DO NOTHING;
	`, devOrgID, devEmail, devUserID)

	// 4. Link Developer to Org as owner
	exec(`
		INSERT INTO developer_org_members (org_id, user_id, org_role)
		VALUES ($1, $2, 'DEV_OWNER')
		ON CONFLICT (org_id, user_id) DO NOTHING;
	`, devOrgID, devUserID)

	// 5. Seed Ledger Accounts (codes match commerce double-entry postings)
	accounts := []struct {
		code, accType, currency string
	}{
		{"ASSET:PG_CLEARING", "ASSET", "KRW"},
		{"LIAB:TAX_PAYABLE", "LIABILITY", "KRW"},
		{"LIAB:DEV_ESCROW", "LIABILITY", "KRW"},
		{"LIAB:DEV_SETTLEABLE", "LIABILITY", "KRW"},
		{"REV:PLATFORM_COMMISSION", "REVENUE", "KRW"},
	}
	for _, acc := range accounts {
		exec(`
			INSERT INTO ledger_accounts (account_code, account_type, currency)
			VALUES ($1, $2, $3)
			ON CONFLICT (account_code) DO NOTHING;
		`, acc.code, acc.accType, acc.currency)
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error("failed to commit seeds", "error", err)
		os.Exit(1)
	}

	logger.Info("database successfully seeded with production reference records")
	fmt.Println("Database seeded successfully.")
}
