package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRepositoryQuerySchemaConformance guards hand-written repository SQL
// against drift from backend/db/migrations/0001_initial_schema.up.sql.
// CI runs without a live database, so this file-level contract is the gate
// that fails when a query references a table or column the schema never
// created (previously masked by in-memory fallbacks).
func TestRepositoryQuerySchemaConformance(t *testing.T) {
	repoRoot := ".."
	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(repoRoot, rel))
		if err != nil {
			t.Fatalf("failed to read %s: %v", rel, err)
		}
		return string(b)
	}

	catalog := read(filepath.Join("internal", "modules", "catalog", "infrastructure", "postgres_repository.go"))
	distribution := read(filepath.Join("internal", "modules", "distribution", "infrastructure", "candidate_provider.go"))
	iam := read(filepath.Join("internal", "modules", "iam", "infrastructure", "postgres_repository.go"))
	schema := read(filepath.Join("db", "migrations", "0001_initial_schema.up.sql"))

	t.Run("catalog latest release uses published statuses and artifact columns", func(t *testing.T) {
		for _, stale := range []string{
			"release_status = 'PUBLISHED'",
			"r.min_sdk",
			"r.target_sdk",
			"r.release_notes,",
			"r.release_notes ",
		} {
			if strings.Contains(catalog, stale) {
				t.Errorf("catalog repository still contains stale SQL %q", stale)
			}
		}
		for _, want := range []string{
			"PUBLISHED_STAGED",
			"PUBLISHED_FULL",
			"release_artifacts",
			"release_notes_i18n",
		} {
			if !strings.Contains(catalog, want) {
				t.Errorf("catalog repository missing schema-aligned SQL %q", want)
			}
		}
		if !strings.Contains(schema, "release_status VARCHAR(36)") {
			t.Errorf("schema guard broken: app_releases.release_status definition not found")
		}
	})

	t.Run("distribution candidates query real catalog tables", func(t *testing.T) {
		for _, stale := range []string{
			"FROM releases",
			"JOIN apps ",
			"r.status",
			"r.supported_abis",
			"r.universal_binary_url",
			"r.universal_sha256",
			"r.file_size_bytes",
			"r.signing_cert_sha256",
		} {
			if strings.Contains(distribution, stale) {
				t.Errorf("candidate provider still contains stale SQL %q", stale)
			}
		}
		for _, want := range []string{
			"FROM app_releases",
			"catalog_apps",
			"release_artifacts",
			"app_distributions",
			"PUBLISHED_STAGED",
		} {
			if !strings.Contains(distribution, want) {
				t.Errorf("candidate provider missing schema-aligned SQL %q", want)
			}
		}
	})

	t.Run("iam member roles use org_role column", func(t *testing.T) {
		if strings.Contains(iam, "SELECT org_id, role FROM developer_org_members") {
			t.Errorf("iam repository selects nonexistent developer_org_members.role column")
		}
		if !strings.Contains(iam, "org_role") {
			t.Errorf("iam repository must select developer_org_members.org_role")
		}
		if !strings.Contains(schema, "org_role VARCHAR(32)") {
			t.Errorf("schema guard broken: developer_org_members.org_role definition not found")
		}
	})
}

// TestSeedMatchesSchema guards backend/cmd/seed/main.go against drift from the
// migration schema (no tenants table, users/developer_orgs/ledger_accounts
// columns must exist in 0001).
func TestSeedMatchesSchema(t *testing.T) {
	seed, err := os.ReadFile(filepath.Join("..", "cmd", "seed", "main.go"))
	if err != nil {
		t.Fatalf("failed to read seed: %v", err)
	}
	src := string(seed)

	for _, stale := range []string{
		"INSERT INTO tenants",
		"tenant_id",
		"role_code, is_active",
		"duns_number",
		"kyb_status",
		"account_name",
		"VALUES ($1, $2, 'OWNER')",
	} {
		if strings.Contains(src, stale) {
			t.Errorf("seed still references schema-unknown %q", stale)
		}
	}
	for _, want := range []string{
		"email_blind_index",
		"display_name",
		"verification_status",
		"org_role",
		"DEV_OWNER",
		"user_roles",
		"account_code",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("seed missing schema-aligned insert %q", want)
		}
	}
}
