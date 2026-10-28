package audit_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/platform/audit"
)

// TC-API-005: Immutable Hash Chain Audit Logger and Tamper Detection
func TestAuditHashChainTamperDetection(t *testing.T) {
	repo := audit.NewInMemoryAuditRepository()
	logger := audit.NewHashChainLogger(repo)

	// 1. Record a sequence of 5 audit events
	for i := 1; i <= 5; i++ {
		actorID := fmt.Sprintf("actor-%d", i)
		action := fmt.Sprintf("APP_STATUS_CHANGE_%d", i)
		targetID := fmt.Sprintf("app-%d", i)
		diff := fmt.Sprintf(`{"status":"APPROVED","round":%d}`, i)

		_, err := logger.Log(actorID, action, "APP", targetID, diff, "127.0.0.1", "test-agent")
		if err != nil {
			t.Fatalf("failed to record audit entry %d: %v", i, err)
		}
	}

	records, err := repo.GetAllRecords()
	if err != nil {
		t.Fatalf("failed to get records: %v", err)
	}
	if len(records) != 5 {
		t.Fatalf("expected 5 records, got %d", len(records))
	}

	// 2. Verify legitimate untampered chain
	valid, tamperedIdx, err := audit.VerifyAuditChain(records)
	if !valid || err != nil || tamperedIdx != -1 {
		t.Fatalf("expected valid chain, got valid=%v, idx=%d, err=%v", valid, tamperedIdx, err)
	}

	// 3. TAMPER TEST A: Attacker alters payload of record #2
	tamperedRecordsA := make([]audit.AuditLogRecord, len(records))
	copy(tamperedRecordsA, records)
	tamperedRecordsA[2].DiffJSON = `{"status":"MALICIOUS_TAMPERED"}`

	validA, tamperedIdxA, errA := audit.VerifyAuditChain(tamperedRecordsA)
	if validA || errA == nil {
		t.Fatalf("expected tamper detection for payload mutation, but passed")
	}
	if tamperedIdxA != 2 {
		t.Fatalf("expected tamper detected at index 2, got %d", tamperedIdxA)
	}
	if !strings.Contains(errA.Error(), "audit record tampered at index 2") {
		t.Fatalf("unexpected error message: %v", errA)
	}

	// 4. TAMPER TEST B: Attacker deletes or replaces record #1, breaking the prev_hash link of record #2
	tamperedRecordsB := make([]audit.AuditLogRecord, len(records))
	copy(tamperedRecordsB, records)
	tamperedRecordsB[2].PrevHash = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

	validB, tamperedIdxB, errB := audit.VerifyAuditChain(tamperedRecordsB)
	if validB || errB == nil {
		t.Fatalf("expected chain break detection, but passed")
	}
	if tamperedIdxB != 2 {
		t.Fatalf("expected broken chain detected at index 2, got %d", tamperedIdxB)
	}
	if !strings.Contains(errB.Error(), "audit chain broken at index 2") {
		t.Fatalf("unexpected error message: %v", errB)
	}
}

