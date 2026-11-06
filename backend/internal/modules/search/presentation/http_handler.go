package presentation

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/search/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/search/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
)

// SearchHandler serves federated search and suggestion endpoints.
type SearchHandler struct {
	svc *application.SearchService
}

func NewSearchHandler(svc *application.SearchService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

func (h *SearchHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.handleSearch)
	r.Get("/suggest", h.handleSuggest)
	return r
}

func (h *SearchHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	offset, _ := strconv.Atoi(q.Get("offset"))
	if offset < 0 {
		offset = 0
	}
	res, err := h.svc.Search(r.Context(), domain.SearchQuery{
		Keyword:            q.Get("q"),
		Locale:             q.Get("locale"),
		Category:           q.Get("category"),
		DirectDownloadOnly: q.Get("direct_only") == "true",
		Limit:              limit,
		Offset:             offset,
	})
	if err != nil {
		errors.WriteProblem(w, http.StatusBadGateway, "ERR_SEARCH_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *SearchHandler) handleSuggest(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	res, err := h.svc.Suggest(r.Context(), q.Get("q"), q.Get("locale"), limit)
	if err != nil {
		errors.WriteProblem(w, http.StatusBadGateway, "ERR_SUGGEST_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}
