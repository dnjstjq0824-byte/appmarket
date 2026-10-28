package presentation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
)

// CatalogHandler serves catalog endpoints.
type CatalogHandler struct {
	svc *application.CatalogService
}

// NewCatalogHandler constructs a CatalogHandler.
func NewCatalogHandler(svc *application.CatalogService) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

// Routes returns the chi router for catalog endpoints.
func (h *CatalogHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/apps", h.handleListApps)
	r.Get("/apps/{slug}", h.handleGetApp)
	r.Get("/categories", h.handleGetCategories)
	r.Get("/charts/{chartType}", h.handleGetChart)
	return r
}

func (h *CatalogHandler) handleListApps(w http.ResponseWriter, r *http.Request) {
	filter := domain.AppListFilter{
		CategoryID: r.URL.Query().Get("category"),
		TagID:      r.URL.Query().Get("tag"),
		Cursor:     r.URL.Query().Get("cursor"),
	}
	if r.URL.Query().Get("direct_only") == "true" {
		filter.DirectOnly = true
	}
	if r.URL.Query().Get("free_only") == "true" {
		filter.FreeOnly = true
	}
	if v := r.URL.Query().Get("min_rating"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MinRating = f
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Limit = n
		}
	}

	summaries, cursor, err := h.svc.ListApps(r.Context(), filter)
	if err != nil {
		errors.WriteProblem(w, http.StatusInternalServerError, "ERR_CATALOG_LIST_FAILED", err.Error(), r.URL.Path)
		return
	}
	resp := map[string]any{
		"items":  summaries,
		"cursor": cursor,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *CatalogHandler) handleGetApp(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	locale := r.Header.Get("Accept-Language")
	if locale == "" {
		locale = r.URL.Query().Get("locale")
	}
	if locale == "" {
		locale = "en-US"
	}
	if idx := strings.Index(locale, ","); idx >= 0 {
		locale = locale[:idx]
	}
	locale = strings.TrimSpace(locale)

	detail, err := h.svc.GetAppDetail(r.Context(), slug, locale)
	if err != nil {
		if err == domain.ErrAppNotFound {
			errors.WriteProblem(w, http.StatusNotFound, "ERR_APP_NOT_FOUND", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusInternalServerError, "ERR_CATALOG_DETAIL_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}

func (h *CatalogHandler) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.GetCategories(r.Context())
	if err != nil {
		errors.WriteProblem(w, http.StatusInternalServerError, "ERR_CATEGORIES_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": cats})
}

func (h *CatalogHandler) handleGetChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"chart_type": chi.URLParam(r, "chartType"), "items": []any{}})
}
