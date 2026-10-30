package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
)

type LedgerService struct {
	mu           sync.RWMutex
	transactions map[string]domain.LedgerTransaction
	idempotency  map[string]string // IdempotencyKey -> TransactionID
	balances     map[string]int64  // Account -> Signed balance
}

func NewLedgerService() *LedgerService {
	return &LedgerService{
		transactions: make(map[string]domain.LedgerTransaction),
		idempotency:  make(map[string]string),
		balances:     make(map[string]int64),
	}
}

// RecordTransaction posts a balanced multi-entry transaction
func (s *LedgerService) RecordTransaction(ctx context.Context, tx domain.LedgerTransaction, idempotencyKey string) (*domain.LedgerTransaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Idempotency Check
	if idempotencyKey != "" {
		if existingID, exists := s.idempotency[idempotencyKey]; exists {
			existingTx := s.transactions[existingID]
			return &existingTx, nil
		}
	}

	// 2. Strict Double-Entry Balance Verification: sum(Debits) == sum(Credits)
	var totalDebits int64 = 0
	var totalCredits int64 = 0

	for _, entry := range tx.Entries {
		if entry.Amount < 0 {
			return nil, fmt.Errorf("negative amount forbidden in ledger entry: %s", entry.Account)
		}
		if entry.Type == domain.EntryDebit {
			totalDebits += entry.Amount
		} else if entry.Type == domain.EntryCredit {
			totalCredits += entry.Amount
		}
	}

	if totalDebits != totalCredits {
		return nil, fmt.Errorf("%w: Debits (%d) != Credits (%d)", domain.ErrLedgerUnbalanced, totalDebits, totalCredits)
	}

	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now().UTC()
	}

	// 3. Post to balances
	for _, entry := range tx.Entries {
		if entry.Type == domain.EntryDebit {
			s.balances[entry.Account] += entry.Amount
		} else {
			s.balances[entry.Account] -= entry.Amount
		}
	}

	s.transactions[tx.ID] = tx
	if idempotencyKey != "" {
		s.idempotency[idempotencyKey] = tx.ID
	}

	return &tx, nil
}

func (s *LedgerService) GetBalance(account string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.balances[account]
}

func (s *LedgerService) GetTransactionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.transactions)
}

// VerifyZeroSumInvariant audits all recorded ledger transactions and verifies sum(Debits) == sum(Credits)
func (s *LedgerService) VerifyZeroSumInvariant(ctx context.Context) (bool, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalDebits int64 = 0
	var totalCredits int64 = 0

	for _, tx := range s.transactions {
		for _, entry := range tx.Entries {
			if entry.Type == domain.EntryDebit {
				totalDebits += entry.Amount
			} else {
				totalCredits += entry.Amount
			}
		}
	}

	diff := totalDebits - totalCredits
	if diff != 0 {
		return false, diff, domain.ErrLedgerUnbalanced
	}
	return true, 0, nil
}
