package application

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/sanitize"
)

// frontendCategory maps storefront report reasons to schema categories.
var frontendCategory = map[string]string{
	"MALWARE_OR_PHISHING":      "MALWARE_SECURITY",
	"DMCA_COPYRIGHT":           "COPYRIGHT_DMCA",
	"TRADEMARK_IMPERSONATION":  "IMPERSONATION",
	"PRIVACY_VIOLATION":        "PRIVACY_VIOLATION",
	"FRAUDULENT_BILLING":       "BILLING_FRAUD",
	"HARMFUL_CONTENT":          "ILLEGAL_CONTENT",
	"MALWARE_SECURITY":         "MALWARE_SECURITY",
	"IMPERSONATION":            "IMPERSONATION",
	"COPYRIGHT_DMCA":           "COPYRIGHT_DMCA",
	"ILLEGAL_CONTENT":          "ILLEGAL_CONTENT",
	"BILLING_FRAUD":            "BILLING_FRAUD",
	"SPAM_ABUSE":               "SPAM_ABUSE",
}

func priorityFor(category string) string {
	switch category {
	case "MALWARE_SECURITY":
		return "P0_CRITICAL"
	case "ILLEGAL_CONTENT", "BILLING_FRAUD", "COPYRIGHT_DMCA":
		return "P1_HIGH"
	default:
		return "P2_NORMAL"
	}
}

type ReportService struct {
	mu      sync.RWMutex
	reports map[string]*domain.ModerationReport
}

func NewReportService() *ReportService {
	return &ReportService{reports: make(map[string]*domain.ModerationReport)}
}

type FileReportInput struct {
	ReporterUserID string
	ReporterEmail  string
	TargetType     string
	TargetID       string
	Category       string
	Description    string
}

// FileReport validates and stores an abuse report, returning the ticket.
func (s *ReportService) FileReport(_ context.Context, in FileReportInput) (*domain.ModerationReport, error) {
	category, ok := frontendCategory[strings.ToUpper(strings.TrimSpace(in.Category))]
	if !ok {
		return nil, fmt.Errorf("%w: unknown category %q", domain.ErrReportInvalid, in.Category)
	}
	email := strings.ToLower(strings.TrimSpace(in.ReporterEmail))
	if email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("%w: valid reporter email required", domain.ErrReportInvalid)
	}
	if strings.TrimSpace(in.TargetID) == "" {
		return nil, fmt.Errorf("%w: report target required", domain.ErrReportInvalid)
	}
	if len(strings.TrimSpace(in.Description)) < 20 {
		return nil, fmt.Errorf("%w: description must be at least 20 characters", domain.ErrReportInvalid)
	}
	targetType := strings.ToUpper(strings.TrimSpace(in.TargetType))
	switch targetType {
	case "APP", "REVIEW", "DEVELOPER_ORG", "":
		if targetType == "" {
			targetType = "APP"
		}
	default:
		return nil, fmt.Errorf("%w: unknown target type %q", domain.ErrReportInvalid, in.TargetType)
	}
	now := time.Now().UTC()
	rep := &domain.ModerationReport{
		ID:             fmt.Sprintf("rep-%d", now.UnixNano()),
		ReporterUserID: in.ReporterUserID,
		ReporterEmail:  email,
		TargetType:     targetType,
		TargetID:       sanitize.Strict(in.TargetID),
		ReportCategory: category,
		Description:    sanitize.Strict(in.Description),
		Priority:       priorityFor(category),
		Status:         domain.ReportOpen,
		CreatedAt:      now,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reports[rep.ID] = rep
	cp := *rep
	return &cp, nil
}

// Claim assigns a moderator and moves OPEN reports into investigation.
func (s *ReportService) Claim(_ context.Context, reportID, moderatorID string) (*domain.ModerationReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rep, ok := s.reports[reportID]
	if !ok {
		return nil, domain.ErrReportNotFound
	}
	if rep.Status == domain.ReportActionTaken || rep.Status == domain.ReportDismissed {
		return nil, domain.ErrReportAlreadyClosed
	}
	rep.AssignedModeratorID = moderatorID
	if rep.Status == domain.ReportOpen {
		rep.Status = domain.ReportInInvestigation
	}
	cp := *rep
	return &cp, nil
}

// Resolve closes a report with an action or dismissal.
func (s *ReportService) Resolve(_ context.Context, reportID, moderatorID string, actionTaken bool, notes string) (*domain.ModerationReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rep, ok := s.reports[reportID]
	if !ok {
		return nil, domain.ErrReportNotFound
	}
	if rep.Status == domain.ReportActionTaken || rep.Status == domain.ReportDismissed {
		return nil, domain.ErrReportAlreadyClosed
	}
	now := time.Now().UTC()
	rep.AssignedModeratorID = moderatorID
	rep.ResolutionNotes = sanitize.Strict(notes)
	rep.ResolvedAt = &now
	if actionTaken {
		rep.Status = domain.ReportActionTaken
	} else {
		rep.Status = domain.ReportDismissed
	}
	cp := *rep
	return &cp, nil
}

func (s *ReportService) List(_ context.Context, status domain.ReportStatus) []domain.ModerationReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.ModerationReport
	for _, rep := range s.reports {
		if status == "" || rep.Status == status {
			out = append(out, *rep)
		}
	}
	return out
}
