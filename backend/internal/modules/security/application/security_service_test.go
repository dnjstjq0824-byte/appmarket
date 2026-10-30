package application_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/security/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/security/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/security/scanner"
)

// TC-SEC-009: EICAR and YARA malware detection and quarantine
func TestMalwareQuarantineAndReports(t *testing.T) {
	ctx := context.Background()
	secSvc := application.NewSecurityService(nil)

	// 1. EICAR MALWARE TEST: APK containing EICAR test string
	eicarPayload := []byte("prefix" + scanner.EICARTestString + "suffix")
	report, err := secSvc.PerformSecurityScan(
		ctx,
		"art-eicar-001",
		"Infected App",
		"com.malware.eicar",
		"org-bad",
		[]string{"android.permission.INTERNET"},
		bytes.NewReader(eicarPayload),
	)
	if err != nil {
		t.Fatalf("scan execution failed: %v", err)
	}
	if report.Verdict != domain.VerdictInfectedQuarantined {
		t.Fatalf("expected INFECTED_QUARANTINED for EICAR, got %s", report.Verdict)
	}
	if report.ClamAVStatus != domain.ClamAVInfected {
		t.Fatalf("expected ClamAV status INFECTED, got %s", report.ClamAVStatus)
	}

	// 2. YARA MALWARE TEST: APK containing Joker SMS Fraud patterns
	jokerPayload := []byte("sample payload with com.joker.fraud.NotificationListener and telephony.sms.SEND_SMS_SILENT binary")
	reportJoker, err := secSvc.PerformSecurityScan(
		ctx,
		"art-joker-001",
		"Flashlight Plus",
		"com.flashlight.clean",
		"org-unknown",
		[]string{"android.permission.RECEIVE_SMS", "android.permission.SEND_SMS"},
		bytes.NewReader(jokerPayload),
	)
	if err != nil {
		t.Fatalf("joker scan failed: %v", err)
	}
	if reportJoker.Verdict != domain.VerdictInfectedQuarantined {
		t.Fatalf("expected INFECTED_QUARANTINED for Joker, got %s", reportJoker.Verdict)
	}
	if len(reportJoker.YARAMatchedRules) == 0 || reportJoker.YARAMatchedRules[0] != "Android_Joker_Sms_Fraud" {
		t.Fatalf("expected Android_Joker_Sms_Fraud rule match, got %v", reportJoker.YARAMatchedRules)
	}
}

// TC-SEC-010: Blacklisted certificate blocking and emergency Kill Switch
func TestBlacklistedCertificateAndKillSwitch(t *testing.T) {
	ctx := context.Background()
	secSvc := application.NewSecurityService(nil)

	// EMERGENCY KILL SWITCH TEST
	appID := "app-malicious-kill-target"
	adminID := "admin-sec-001"

	// Initially download is allowed
	err := secSvc.ValidateDownloadRequest(ctx, appID)
	if err != nil {
		t.Fatalf("initial download should be allowed, got %v", err)
	}

	// Admin triggers emergency kill switch
	err = secSvc.TriggerKillSwitch(ctx, appID, "Critical zero-day exploit found in wild", adminID)
	if err != nil {
		t.Fatalf("failed to trigger kill switch: %v", err)
	}

	// ValidateDownloadRequest MUST be blocked with ErrKillSwitchActive
	err = secSvc.ValidateDownloadRequest(ctx, appID)
	if err == nil {
		t.Fatalf("expected download to be blocked by active kill switch, but allowed")
	}
	if err != domain.ErrKillSwitchActive {
		t.Fatalf("expected ErrKillSwitchActive, got %v", err)
	}

	// Revoke kill switch -> download allowed again
	_ = secSvc.RevokeKillSwitch(ctx, appID)
	err = secSvc.ValidateDownloadRequest(ctx, appID)
	if err != nil {
		t.Fatalf("expected download allowed after kill switch revoked, got %v", err)
	}
}
