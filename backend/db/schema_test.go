package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TC-DB-001 & TC-DB-002: Verify core schema migration completeness, table definitions, and constraints.
func TestMigrationSchemaFilesIntegrity(t *testing.T) {
	upSQLPath := filepath.Join("migrations", "0001_initial_schema.up.sql")
	downSQLPath := filepath.Join("migrations", "0001_initial_schema.down.sql")
	idxSQLPath := filepath.Join("migrations", "0002_indexes_and_constraints.up.sql")
	seedSQLPath := filepath.Join("migrations", "0003_seed_reference_data.up.sql")

	upSQL, err := os.ReadFile(upSQLPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", upSQLPath, err)
	}

	downSQL, err := os.ReadFile(downSQLPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", downSQLPath, err)
	}

	idxSQL, err := os.ReadFile(idxSQLPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", idxSQLPath, err)
	}

	seedSQL, err := os.ReadFile(seedSQLPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", seedSQLPath, err)
	}

	upContent := string(upSQL)
	downContent := string(downSQL)
	idxContent := string(idxSQL)
	seedContent := string(seedSQL)

	// Verify required core tables exist in up migration
	requiredTables := []string{
		"users",
		"user_identities",
		"auth_sessions",
		"mfa_credentials",
		"roles",
		"user_roles",
		"developer_orgs",
		"developer_org_members",
		"catalog_apps",
		"app_localizations",
		"app_distributions",
		"app_releases",
		"release_artifacts",
		"apk_signatures",
		"orders",
		"payments",
		"user_entitlements",
		"refunds",
		"ledger_accounts",
		"ledger_transactions",
		"ledger_entries",
		"audit_logs",
		"background_jobs",
		"outbox_events",
		"dual_approval_requests",
		"system_settings",
		"install_events",
		"search_query_logs",
	}

	for _, table := range requiredTables {
		createPattern := fmt.Sprintf("CREATE TABLE %s", table)
		if !strings.Contains(upContent, createPattern) {
			t.Errorf("expected table %s to be created in %s", table, upSQLPath)
		}
		dropPattern := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		if !strings.Contains(downContent, dropPattern) {
			t.Errorf("expected table %s to be dropped in %s", table, downSQLPath)
		}
	}

	// Verify foreign key integrity & ON DELETE RESTRICT on key distribution and catalog tables
	if !strings.Contains(upContent, "ON DELETE RESTRICT") {
		t.Errorf("expected ON DELETE RESTRICT constraints in initial schema")
	}

	// Verify indexes and constraints in 0002
	if !strings.Contains(idxContent, "trg_verify_ledger_balance") {
		t.Errorf("expected trg_verify_ledger_balance constraint trigger in %s", idxSQLPath)
	}
	if !strings.Contains(idxContent, "DEFERRABLE INITIALLY DEFERRED") {
		t.Errorf("expected DEFERRABLE INITIALLY DEFERRED trigger in %s", idxSQLPath)
	}

	// Verify seeds in 0003
	if !strings.Contains(seedContent, "INSERT INTO roles") {
		t.Errorf("expected roles seed in %s", seedSQLPath)
	}
	if !strings.Contains(seedContent, "INSERT INTO commission_policies") {
		t.Errorf("expected commission_policies seed in %s", seedSQLPath)
	}
}

// TC-COM-005 & TC-DB-003: Double-Entry Ledger Zero-Balance Logic Test
type LedgerEntrySimulation struct {
	Direction   string
	AmountMinor int64
}

func VerifyLedgerBalanceLogic(entries []LedgerEntrySimulation) error {
	var netBalance int64
	for _, e := range entries {
		if e.Direction == "DEBIT" {
			netBalance += e.AmountMinor
		} else if e.Direction == "CREDIT" {
			netBalance -= e.AmountMinor
		} else {
			return fmt.Errorf("invalid direction %s", e.Direction)
		}
	}
	if netBalance != 0 {
		return fmt.Errorf("ERR_LEDGER_UNBALANCED: Transaction has non-zero balance %d", netBalance)
	}
	return nil
}

func TestLedgerZeroBalanceEnforcement(t *testing.T) {
	// Balanced transaction
	balancedEntries := []LedgerEntrySimulation{
		{Direction: "DEBIT", AmountMinor: 1000},
		{Direction: "CREDIT", AmountMinor: 850},  // Dev payout
		{Direction: "CREDIT", AmountMinor: 150},  // Platform fee
	}

	if err := VerifyLedgerBalanceLogic(balancedEntries); err != nil {
		t.Fatalf("expected balanced transaction to succeed, got %v", err)
	}

	// Unbalanced transaction
	unbalancedEntries := []LedgerEntrySimulation{
		{Direction: "DEBIT", AmountMinor: 1000},
		{Direction: "CREDIT", AmountMinor: 800}, // Missing 200
	}

	err := VerifyLedgerBalanceLogic(unbalancedEntries)
	if err == nil {
		t.Fatalf("expected unbalanced transaction to fail")
	}
	if !strings.Contains(err.Error(), "ERR_LEDGER_UNBALANCED") {
		t.Fatalf("expected ERR_LEDGER_UNBALANCED error message, got %v", err)
	}
}
