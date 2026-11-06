package presentation_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/analytics"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/analytics/presentation"
)

func TestIngestEvents(t *testing.T) {
	store := analytics.NewMemoryEventStore()
	handler := presentation.NewAnalyticsHandler(analytics.NewFunnelIngestor(store))
	router := handler.Routes()

	batch := []analytics.AnalyticsEventEnvelope{{
		EventID: "evt-1", EventName: "APP_VIEW", OccurredAt: time.Now().UTC(),
		SessionHash: "sess-1", Platform: "ANDROID", CountryCode: "KR",
		Properties: analytics.EventProperties{AppID: "app-1"},
	}}
	raw, _ := json.Marshal(batch)
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("ingest: %d %s", w.Code, w.Body.String())
	}
	req = httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte(`[]`)))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected empty batch rejection, got %d", w.Code)
	}
}
