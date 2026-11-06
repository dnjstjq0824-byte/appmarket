package presentation

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/analytics"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
)

// AnalyticsHandler ingests client telemetry batches.
type AnalyticsHandler struct {
	ingestor *analytics.FunnelIngestor
}

func NewAnalyticsHandler(ingestor *analytics.FunnelIngestor) *AnalyticsHandler {
	return &AnalyticsHandler{ingestor: ingestor}
}

func (h *AnalyticsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/events", h.handleIngestEvents)
	return r
}

func (h *AnalyticsHandler) handleIngestEvents(w http.ResponseWriter, r *http.Request) {
	var batch []analytics.AnalyticsEventEnvelope
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	if len(batch) == 0 || len(batch) > 500 {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BATCH", "Batch must contain 1-500 events", r.URL.Path)
		return
	}
	accepted, err := h.ingestor.IngestBatch(r.Context(), batch)
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INGEST_REJECTED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"accepted": accepted})
}
