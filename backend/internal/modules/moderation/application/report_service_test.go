package application

import (
	"context"
	"strings"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/domain"
)

func TestReportFileClaimResolve(t *testing.T) {
	ctx := context.Background()
	svc := NewReportService()
	rep, err := svc.FileReport(ctx, FileReportInput{
		ReporterEmail: "user@example.com", TargetID: "com.evil.app",
		Category: "MALWARE_OR_PHISHING", Description: "This app steals credentials after asking for accessibility access",
	})
	if err != nil {
		t.Fatalf("file failed: %v", err)
	}
	if rep.ReportCategory != "MALWARE_SECURITY" || rep.Priority != "P0_CRITICAL" || rep.Status != domain.ReportOpen {
		t.Fatalf("unexpected report: %+v", rep)
	}
	claimed, err := svc.Claim(ctx, rep.ID, "mod-1")
	if err != nil || claimed.Status != domain.ReportInInvestigation {
		t.Fatalf("expected investigation, got %+v / %v", claimed, err)
	}
	resolved, err := svc.Resolve(ctx, rep.ID, "mod-1", true, "App suspended <script>x</script>")
	if err != nil || resolved.Status != domain.ReportActionTaken {
		t.Fatalf("expected action taken, got %+v / %v", resolved, err)
	}
	if strings.Contains(resolved.ResolutionNotes, "<script") {
		t.Fatalf("resolution notes not sanitized: %q", resolved.ResolutionNotes)
	}
	if _, err := svc.Resolve(ctx, rep.ID, "mod-1", false, "again"); err != domain.ErrReportAlreadyClosed {
		t.Fatalf("expected already-closed, got %v", err)
	}
	if open := svc.List(ctx, domain.ReportOpen); len(open) != 0 {
		t.Fatalf("expected 0 open, got %d", len(open))
	}
}

func TestReportValidation(t *testing.T) {
	ctx := context.Background()
	svc := NewReportService()
	base := FileReportInput{
		ReporterEmail: "user@example.com", TargetID: "com.x", Category: "SPAM_ABUSE",
		Description: "This description is long enough to pass validation rules",
	}
	if _, err := svc.FileReport(ctx, base); err != nil {
		t.Fatalf("valid report rejected: %v", err)
	}
	bad := base
	bad.Category = "NOT_A_CATEGORY"
	if _, err := svc.FileReport(ctx, bad); err == nil {
		t.Fatalf("expected category error")
	}
	bad = base
	bad.ReporterEmail = "not-an-email"
	if _, err := svc.FileReport(ctx, bad); err == nil {
		t.Fatalf("expected email error")
	}
	bad = base
	bad.Description = "too short"
	if _, err := svc.FileReport(ctx, bad); err == nil {
		t.Fatalf("expected description error")
	}
}
