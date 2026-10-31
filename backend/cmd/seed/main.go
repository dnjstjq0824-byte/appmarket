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
	"golang.org/x/crypto/bcrypt"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://bn_app:bn_secure_password_2026@127.0.0.1:5432/bn_appmarket?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("connecting to postgres for seeding", "url", strings.Split(databaseURL, "@")[len(strings.Split(databaseURL, "@"))-1])
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

	tx, err := pool.Begin(ctx)
	if err != nil {
		logger.Error("failed to start tx", "error", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	// 1. Seed Tenant
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	_, _ = tx.Exec(ctx, `
		INSERT INTO tenants (id, slug, name, is_active)
		VALUES ($1, 'global', 'BookingNetwork Global', true)
		ON CONFLICT (id) DO NOTHING;
	`, tenantID)

	// 2. Seed Super Admin User
	adminPassHash, _ := bcrypt.GenerateFromPassword([]byte("AdminSecurePassword2026!"), bcrypt.DefaultCost)
	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	_, _ = tx.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, password_hash, role_code, is_active, email_verified)
		VALUES ($1, $2, 'admin@bookingnetwork.com', $3, 'ADMIN_SUPER', true, true)
		ON CONFLICT (email) DO NOTHING;
	`, adminID, tenantID, string(adminPassHash))

	// 3. Seed Verified Developer Org
	devOrgID := uuid.MustParse("00000000-0000-0000-0000-000000000020")
	_, _ = tx.Exec(ctx, `
		INSERT INTO developer_orgs (id, tenant_id, legal_name, duns_number, kyb_status, country_code)
		VALUES ($1, $2, 'Acme Technologies Inc.', '12-345-6789', 'VERIFIED_KYB', 'KR')
		ON CONFLICT (id) DO NOTHING;
	`, devOrgID, tenantID)

	// 4. Seed Developer User
	devPassHash, _ := bcrypt.GenerateFromPassword([]byte("DevSecurePassword2026!"), bcrypt.DefaultCost)
	devUserID := uuid.MustParse("00000000-0000-0000-0000-000000000021")
	_, _ = tx.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, password_hash, role_code, is_active, email_verified)
		VALUES ($1, $2, 'dev@acmecorp.com', $3, 'DEVELOPER_ADMIN', true, true)
		ON CONFLICT (email) DO NOTHING;
	`, devUserID, tenantID, string(devPassHash))

	// Link Developer to Org
	_, _ = tx.Exec(ctx, `
		INSERT INTO developer_org_members (org_id, user_id, org_role)
		VALUES ($1, $2, 'OWNER')
		ON CONFLICT (org_id, user_id) DO NOTHING;
	`, devOrgID, devUserID)

	// 5. Seed Ledger Accounts
	accounts := []struct {
		code, name, accType string
	}{
		{"1010", "Toss / PortOne PG Settlement Receivable", "ASSET"},
		{"2010", "Developer Merchant Payable (Net 85%)", "LIABILITY"},
		{"4010", "Platform Commission Revenue (15%)", "REVENUE"},
	}
	for _, acc := range accounts {
		_, _ = tx.Exec(ctx, `
			INSERT INTO ledger_accounts (account_code, account_name, account_type, is_active)
			VALUES ($1, $2, $3, true)
			ON CONFLICT (account_code) DO NOTHING;
		`, acc.code, acc.name, acc.accType)
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error("failed to commit seeds", "error", err)
		os.Exit(1)
	}

	logger.Info("database successfully seeded with production reference records")
	fmt.Println("Database seeded successfully.")
}

