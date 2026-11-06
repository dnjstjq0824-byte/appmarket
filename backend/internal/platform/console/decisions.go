package console

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	catalogApp "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/audit"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/sanitize"
)

var ErrDecisionInvalid = errors.New("ERR_DECISION_INVALID")

// ReviewDecision records an admin approve/reject verdict on a submission.
type ReviewDecision struct {
	AppID      string         `json:"appId"`
	Decision   string         `json:"decision"`
	Notes      string         `json:"notes"`
	Checklist  map[string]bool `json:"checklist"`
	DecidedBy  string         `json:"decidedBy"`
	DecidedAt  time.Time      `json:"decidedAt"`
}

// DecisionService stores review verdicts and mirrors them to the hash-chained
// audit log.
type DecisionService struct {
	mu        sync.RWMutex
	decisions map[string]*ReviewDecision
	audit     *audit.HashChainLogger
	catalog   *catalogApp.CatalogService
}

func NewDecisionService(logger *audit.HashChainLogger) *DecisionService {
	return &DecisionService{decisions: make(map[string]*ReviewDecision), audit: logger}
}

// WithCatalog enables catalog status flips on decisions. Without it Decide
// records the verdict only (unit-test mode).
func (s *DecisionService) WithCatalog(catalog *catalogApp.CatalogService) *DecisionService {
	s.catalog = catalog
	return s
}

// Decide validates and records an APPROVED/REJECTED verdict.
func (s *DecisionService) Decide(ctx context.Context, appID, decision, notes string, checklist map[string]bool, adminID, ip, userAgent string) (*ReviewDecision, error) {
	verdict := strings.ToUpper(strings.TrimSpace(decision))
	if verdict != "APPROVED" && verdict != "REJECTED" {
		return nil, fmt.Errorf("%w: decision must be APPROVED or REJECTED", ErrDecisionInvalid)
	}
	if verdict == "REJECTED" && strings.TrimSpace(notes) == "" {
		return nil, fmt.Errorf("%w: rejection requires reviewer notes", ErrDecisionInvalid)
	}
	if s.catalog != nil {
		if err := s.catalog.ApplyReviewDecision(ctx, appID, verdict == "APPROVED"); err != nil {
			return nil, err
		}
	}
	d := &ReviewDecision{
		AppID:     appID,
		Decision:  verdict,
		Notes:     sanitize.Strict(notes),
		Checklist: checklist,
		DecidedBy: adminID,
		DecidedAt: time.Now().UTC(),
	}
	s.mu.Lock()
	s.decisions[appID] = d
	s.mu.Unlock()
	if s.audit != nil {
		diff, _ := json.Marshal(map[string]any{"decision": verdict, "notes": d.Notes})
		_, _ = s.audit.Log(adminID, "admin.review.decision", "app", appID, string(diff), ip, userAgent)
	}
	cp := *d
	return &cp, nil
}

func (s *DecisionService) GetDecision(_ context.Context, appID string) (*ReviewDecision, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.decisions[appID]
	if !ok {
		return nil, fmt.Errorf("%w: no decision for app", ErrDecisionInvalid)
	}
	cp := *d
	return &cp, nil
}

func (s *DecisionService) ListDecisions(_ context.Context) []ReviewDecision {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ReviewDecision, 0, len(s.decisions))
	for _, d := range s.decisions {
		out = append(out, *d)
	}
	return out
}
