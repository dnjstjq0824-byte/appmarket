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
	distApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	distDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
	searchDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/search/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/audit"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/sanitize"
)

// SearchIndexer receives a search document whenever an app is approved. The
// production implementation is *search.Application.SearchService (OpenSearch
// primary, fallback store); indexing stays best-effort so a search outage
// never blocks a review decision.
type SearchIndexer interface {
	IndexDocument(ctx context.Context, doc searchDomain.SearchDoc) error
}

var ErrDecisionInvalid = errors.New("ERR_DECISION_INVALID")

// ReviewDecision records an admin approve/reject verdict on a submission.
type ReviewDecision struct {
	AppID     string          `json:"appId"`
	Decision  string          `json:"decision"`
	Notes     string          `json:"notes"`
	Checklist map[string]bool `json:"checklist"`
	DecidedBy string          `json:"decidedBy"`
	DecidedAt time.Time       `json:"decidedAt"`
}

// DecisionService stores review verdicts and mirrors them to the hash-chained
// audit log.
type DecisionService struct {
	mu        sync.RWMutex
	decisions map[string]*ReviewDecision
	audit     *audit.HashChainLogger
	catalog   *catalogApp.CatalogService
	search    SearchIndexer
	releases  *distApp.ArtifactService
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

// WithSearchIndexer enables best-effort search indexing on approval.
func (s *DecisionService) WithSearchIndexer(search SearchIndexer) *DecisionService {
	s.search = search
	return s
}

// WithReleases enables best-effort release advancement on decisions: the
// newest in-review release moves to APPROVED_PENDING_PUBLISH on approval and
// back to CHANGES_REQUESTED on rejection. Never fails the recorded verdict.
func (s *DecisionService) WithReleases(artifacts *distApp.ArtifactService) *DecisionService {
	s.releases = artifacts
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
		if verdict == "APPROVED" {
			s.indexApprovedApp(ctx, appID)
		}
		s.advanceReleaseForDecision(ctx, appID, verdict == "APPROVED")
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

// indexApprovedApp pushes the freshly published app to the search index.
// Every step is best-effort: lookup/index failures are swallowed so search
// availability never affects the recorded review verdict.
func (s *DecisionService) indexApprovedApp(ctx context.Context, appID string) {
	if s.search == nil {
		return
	}
	app, err := s.catalog.GetAppByID(ctx, appID)
	if err != nil || app == nil {
		return
	}
	doc := searchDomain.SearchDoc{
		AppID:           app.ID,
		PackageName:     app.PrimaryPackageName,
		TitleI18n:       map[string]string{},
		DescriptionI18n: map[string]string{},
		Category:        app.PrimaryCategoryID,
		BayesianRating:  app.BayesianRating,
		DownloadCount:   app.InstallCount,
		PublishedAt:     time.Now().UTC(),
	}
	if locs, err := s.catalog.ListLocalizations(ctx, appID); err == nil {
		for _, loc := range locs {
			if loc.Title != "" {
				doc.TitleI18n[loc.Locale] = loc.Title
			}
			if desc := firstNonEmpty(loc.FullDescription, loc.ShortDescription); desc != "" {
				doc.DescriptionI18n[loc.Locale] = desc
			}
		}
	}
	if len(doc.TitleI18n) == 0 {
		doc.TitleI18n[app.DefaultLocale] = app.CanonicalSlug
	}
	_ = s.search.IndexDocument(ctx, doc)
}

// advanceReleaseForDecision moves the newest in-review release along with the
// verdict: APPROVED_PENDING_PUBLISH on approval, CHANGES_REQUESTED (via
// IN_REVIEW) on rejection so the developer can fix and resubmit. Best-effort:
// releases are ephemeral working state, so a missing or mis-stated release
// never fails the recorded app verdict.
func (s *DecisionService) advanceReleaseForDecision(ctx context.Context, appID string, approved bool) {
	if s.releases == nil {
		return
	}
	releases, err := s.releases.ListReleasesByApp(ctx, appID)
	if err != nil || len(releases) == 0 {
		return
	}
	current := releases[0]
	steps := releaseDecisionSteps(current.ReleaseStatus, approved)
	for _, step := range steps {
		if _, err := s.releases.TransitionReleaseStatus(ctx, current.ID, step); err != nil {
			return
		}
	}
}

func releaseDecisionSteps(status distDomain.ReleaseStatus, approved bool) []distDomain.ReleaseStatus {
	if approved {
		switch status {
		case distDomain.ReleaseSubmittedForReview:
			return []distDomain.ReleaseStatus{distDomain.ReleaseInReview, distDomain.ReleaseApprovedPendingPub}
		case distDomain.ReleaseInReview:
			return []distDomain.ReleaseStatus{distDomain.ReleaseApprovedPendingPub}
		}
		return nil
	}
	switch status {
	case distDomain.ReleaseSubmittedForReview:
		return []distDomain.ReleaseStatus{distDomain.ReleaseInReview, distDomain.ReleaseChangesRequested}
	case distDomain.ReleaseInReview:
		return []distDomain.ReleaseStatus{distDomain.ReleaseChangesRequested}
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
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
