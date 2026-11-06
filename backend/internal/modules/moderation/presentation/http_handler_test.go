package presentation_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/presentation"
)

func moderationRouter() *presentation.ModerationHandler {
	verifier := application.NewInMemoryInstallVerifier()
	verifier.RecordInstallation("user-1", "app-1")
	reviews := application.NewReviewService(verifier, application.NewAntiFraudDetector())
	return presentation.NewModerationHandler(reviews, application.NewReportService(), application.NewAppealService(), func(*http.Request) (string, error) {
		return "user-1", nil
	})
}

func doMod(t *testing.T, router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestFileReportValidation(t *testing.T) {
	router := moderationRouter().Routes()
	w := doMod(t, router, http.MethodPost, "/reports", map[string]any{
		"appIdentifier": "com.evil.app", "category": "MALWARE_OR_PHISHING",
		"reporterEmail": "user@example.com", "description": "Steals credentials after accessibility permission is granted", "confirmAccuracy": true,
	})
	if w.Code != http.StatusCreated || !strings.Contains(w.Body.String(), "ticketId") {
		t.Fatalf("file report: %d %s", w.Code, w.Body.String())
	}
	w = doMod(t, router, http.MethodPost, "/reports", map[string]any{
		"appIdentifier": "com.evil.app", "category": "MALWARE_OR_PHISHING",
		"reporterEmail": "user@example.com", "description": "Steals credentials after accessibility permission is granted", "confirmAccuracy": false,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected accuracy requirement, got %d", w.Code)
	}
	w = doMod(t, router, http.MethodPost, "/reports", map[string]any{
		"appIdentifier": "com.evil.app", "category": "MALWARE_OR_PHISHING",
		"reporterEmail": "user@example.com", "description": "short", "confirmAccuracy": true,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected description validation, got %d", w.Code)
	}
}

func TestPostAndListReviews(t *testing.T) {
	router := moderationRouter().Routes()
	w := doMod(t, router, http.MethodPost, "/reviews", map[string]any{
		"appId": "app-1", "rating": 5, "title": "Great", "comment": "Works perfectly on my device", "versionCode": 10,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("post review: %d %s", w.Code, w.Body.String())
	}
	// Zero account age trips the anti-fraud detector, so the review is held
	// out of the public list. Both halves are asserted.
	if !strings.Contains(w.Body.String(), "FLAGGED_SUSPICIOUS") {
		t.Fatalf("expected suspicious flag, got %s", w.Body.String())
	}
	w = doMod(t, router, http.MethodGet, "/apps/app-1/reviews", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list reviews: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "Great") {
		t.Fatalf("flagged review leaked into public list: %s", w.Body.String())
	}
}

func TestSubmitAppeal(t *testing.T) {
	router := moderationRouter().Routes()
	w := doMod(t, router, http.MethodPost, "/appeals", map[string]any{
		"orgId": "org-1", "releaseId": "rel-1", "reason": "Permission is justified for NFC wallet use", "originalReviewerId": "rev-1",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("submit appeal: %d %s", w.Code, w.Body.String())
	}
}
