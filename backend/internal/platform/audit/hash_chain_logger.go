package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const GenesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

type AuditLogRecord struct {
	ID         string    `json:"id"`
	ActorID    string    `json:"actor_id"`
	Action     string    `json:"action"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	DiffJSON   string    `json:"diff_json"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
	PrevHash   string    `json:"prev_hash"`
	RecordHash string    `json:"record_hash"`
}

// ComputeRecordHash calculates the immutable SHA-256 hash for an audit record linked to its prev_hash.
func ComputeRecordHash(prevHash, actorID, action, targetType, targetID, diffJSON string, createdAt time.Time) string {
	payload := fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s|%d",
		prevHash,
		actorID,
		action,
		targetType,
		targetID,
		diffJSON,
		createdAt.UnixNano(),
	)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

type AuditRepository interface {
	GetLastRecord() (*AuditLogRecord, error)
	AppendRecord(record *AuditLogRecord) error
	GetAllRecords() ([]AuditLogRecord, error)
}

type InMemoryAuditRepository struct {
	mu      sync.RWMutex
	records []AuditLogRecord
}

func NewInMemoryAuditRepository() *InMemoryAuditRepository {
	return &InMemoryAuditRepository{
		records: make([]AuditLogRecord, 0),
	}
}

func (r *InMemoryAuditRepository) GetLastRecord() (*AuditLogRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.records) == 0 {
		return nil, nil
	}
	last := r.records[len(r.records)-1]
	return &last, nil
}

func (r *InMemoryAuditRepository) AppendRecord(record *AuditLogRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, *record)
	return nil
}

func (r *InMemoryAuditRepository) GetAllRecords() ([]AuditLogRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]AuditLogRecord, len(r.records))
	copy(res, r.records)
	return res, nil
}

type HashChainLogger struct {
	repo AuditRepository
	mu   sync.Mutex
}

func NewHashChainLogger(repo AuditRepository) *HashChainLogger {
	if repo == nil {
		repo = NewInMemoryAuditRepository()
	}
	return &HashChainLogger{repo: repo}
}

// Log records a new audit event into the immutable cryptographic chain.
func (l *HashChainLogger) Log(actorID, action, targetType, targetID, diffJSON, ipAddress, userAgent string) (*AuditLogRecord, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	last, err := l.repo.GetLastRecord()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve last audit record: %w", err)
	}

	prevHash := GenesisHash
	if last != nil && last.RecordHash != "" {
		prevHash = last.RecordHash
	}

	now := time.Now().UTC()
	recordHash := ComputeRecordHash(prevHash, actorID, action, targetType, targetID, diffJSON, now)

	record := &AuditLogRecord{
		ID:         uuid.New().String(),
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		DiffJSON:   diffJSON,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  now,
		PrevHash:   prevHash,
		RecordHash: recordHash,
	}

	if err := l.repo.AppendRecord(record); err != nil {
		return nil, fmt.Errorf("failed to persist audit log record: %w", err)
	}

	return record, nil
}

// VerifyAuditChain inspects the full chain to detect any tampering or broken link (TC-API-005).
func VerifyAuditChain(records []AuditLogRecord) (bool, int, error) {
	if len(records) == 0 {
		return true, -1, nil
	}

	expectedPrev := GenesisHash
	for i, rec := range records {
		// 1. Check prev_hash link
		if rec.PrevHash != expectedPrev {
			return false, i, fmt.Errorf("audit chain broken at index %d (ID %s): expected prev_hash %s, got %s",
				i, rec.ID, expectedPrev, rec.PrevHash)
		}

		// 2. Recompute record hash to detect internal payload tampering
		computedHash := ComputeRecordHash(
			rec.PrevHash,
			rec.ActorID,
			rec.Action,
			rec.TargetType,
			rec.TargetID,
			rec.DiffJSON,
			rec.CreatedAt,
		)

		if rec.RecordHash != computedHash {
			return false, i, fmt.Errorf("audit record tampered at index %d (ID %s): computed %s, recorded %s",
				i, rec.ID, computedHash, rec.RecordHash)
		}

		expectedPrev = rec.RecordHash
	}

	return true, -1, nil
}
