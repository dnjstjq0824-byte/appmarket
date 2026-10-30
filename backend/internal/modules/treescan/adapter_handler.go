package treescan

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrFeatureDisabled       = errors.New("ERR_TREESCAN_INTEGRATION_DISABLED")
	ErrUnauthorizedService   = errors.New("ERR_UNAUTHORIZED_TREESCAN_SERVICE_TOKEN")
	ErrInvalidCursor         = errors.New("ERR_INVALID_CATALOG_CURSOR")
)

type FederatedItem struct {
	AppID              string   `json:"app_id"`
	CanonicalSlug      string   `json:"canonical_slug"`
	PrimaryPackageName string   `json:"primary_package_name"`
	Title              string   `json:"title"`
	DeveloperName      string   `json:"developer_name"`
	IconURL            string   `json:"icon_url"`
	BayesianRating     float64  `json:"bayesian_rating"`
	RatingCount        int      `json:"rating_count"`
	MonetizationType   string   `json:"monetization_type"`
	CanDirectInstall   bool     `json:"can_direct_install"`
	AvailableSources   []string `json:"available_sources"`
	ProvenanceSummary  string   `json:"provenance_summary"`
	CanonicalMarketURL string   `json:"canonical_market_url"`
}

type FederatedSearchResponse struct {
	Query   string          `json:"query"`
	TookMs  int64           `json:"took_ms"`
	Items   []FederatedItem `json:"items"`
	Total   int             `json:"total"`
}

type FeedAction string

const (
	ActionIndexUpsert     FeedAction = "INDEX_UPSERT"
	ActionRemoveFromIndex FeedAction = "REMOVE_FROM_INDEX"
)

type CatalogFeedItem struct {
	AppID          string     `json:"app_id"`
	PackageName    string     `json:"package_name"`
	Title          string     `json:"title"`
	VersionCode    int64      `json:"version_code"`
	CatalogStatus  string     `json:"catalog_status"`
	Action         FeedAction `json:"action"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Cursor         string     `json:"cursor"`
}

type CatalogFeedResponse struct {
	Items      []CatalogFeedItem `json:"items"`
	NextCursor string            `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}

type AdapterHandler struct {
	mu           sync.RWMutex
	enabled      bool
	serviceToken string
	catalogItems []CatalogFeedItem
	searchIndex  []FederatedItem
}

func NewAdapterHandler(serviceToken string) *AdapterHandler {
	// Standalone default: disabled unless environment variable is explicitly set
	enabled := strings.EqualFold(os.Getenv("TREESCAN_INTEGRATION_ENABLED"), "true")
	if serviceToken == "" {
		serviceToken = "ts_sec_prod_treescan_federated_2026"
	}

	h := &AdapterHandler{
		enabled:      enabled,
		serviceToken: serviceToken,
		catalogItems: make([]CatalogFeedItem, 0),
		searchIndex:  make([]FederatedItem, 0),
	}

	// Initialize standard catalog items
	h.seedInitialData()
	return h
}

func (h *AdapterHandler) SetEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = enabled
}

func (h *AdapterHandler) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.enabled
}

func (h *AdapterHandler) seedInitialData() {
	now := time.Now().UTC()

	h.searchIndex = []FederatedItem{
		{
			AppID:              "0192ce10-4b00-7000-8000-000000000301",
			CanonicalSlug:      "osmand-maps-navigation",
			PrimaryPackageName: "net.osmand.plus",
			Title:              "OsmAnd — 오프라인 지도 및 내비게이션",
			DeveloperName:      "OsmAnd B.V.",
			IconURL:            "https://cdn.appmarket.bookingnetwork.internal/icons/osmand.webp",
			BayesianRating:     4.72,
			RatingCount:        1840,
			MonetizationType:   "FREE",
			CanDirectInstall:   true,
			AvailableSources:   []string{"BN_HOSTED", "FDROID_REPO", "GOOGLE_PLAY_LINK"},
			ProvenanceSummary:  "BN Verified Direct APK & F-Droid Official",
			CanonicalMarketURL: "https://appmarket.bookingnetwork.internal/apps/osmand-maps-navigation?utm_source=treescan",
		},
		{
			AppID:              "0192ce10-4b00-7000-8000-000000000302",
			CanonicalSlug:      "vector-pro-studio",
			PrimaryPackageName: "com.creativedesign.vectorpro",
			Title:              "Vector Pro Studio",
			DeveloperName:      "Creative Design Labs",
			IconURL:            "https://cdn.appmarket.bookingnetwork.internal/icons/vectorpro.webp",
			BayesianRating:     4.85,
			RatingCount:        520,
			MonetizationType:   "PAID",
			CanDirectInstall:   true,
			AvailableSources:   []string{"BN_HOSTED"},
			ProvenanceSummary:  "BN Verified Direct APK",
			CanonicalMarketURL: "https://appmarket.bookingnetwork.internal/apps/vector-pro-studio?utm_source=treescan",
		},
	}

	h.catalogItems = []CatalogFeedItem{
		{
			AppID:         "0192ce10-4b00-7000-8000-000000000301",
			PackageName:   "net.osmand.plus",
			Title:         "OsmAnd — 오프라인 지도 및 내비게이션",
			VersionCode:   450,
			CatalogStatus: "PUBLISHED",
			Action:        ActionIndexUpsert,
			UpdatedAt:     now.Add(-2 * time.Hour),
			Cursor:        fmt.Sprintf("%d_1", now.Add(-2*time.Hour).Unix()),
		},
		{
			AppID:         "0192ce10-4b00-7000-8000-000000000302",
			PackageName:   "com.creativedesign.vectorpro",
			Title:         "Vector Pro Studio",
			VersionCode:   240,
			CatalogStatus: "PUBLISHED",
			Action:        ActionIndexUpsert,
			UpdatedAt:     now.Add(-1 * time.Hour),
			Cursor:        fmt.Sprintf("%d_2", now.Add(-1*time.Hour).Unix()),
		},
		{
			AppID:         "0192ce10-4b00-7000-8000-000000000303",
			PackageName:   "com.malicious.spyware",
			Title:         "Fake Malware Flashlight",
			VersionCode:   12,
			CatalogStatus: "SUSPENDED_BY_ADMIN",
			Action:        ActionRemoveFromIndex,
			UpdatedAt:     now,
			Cursor:        fmt.Sprintf("%d_3", now.Unix()),
		},
	}
}

func (h *AdapterHandler) FederatedSearch(ctx context.Context, query string, limit int, token string) (*FederatedSearchResponse, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 1. Feature Flag Protection (Phase 10 independence)
	if !h.enabled {
		return nil, ErrFeatureDisabled
	}

	// 2. Service Token Auth Header Check
	if token != h.serviceToken {
		return nil, ErrUnauthorizedService
	}

	start := time.Now()

	// Lightweight in-memory projection filter
	var matched []FederatedItem
	qLower := strings.ToLower(query)

	for _, item := range h.searchIndex {
		if strings.Contains(strings.ToLower(item.Title), qLower) ||
			strings.Contains(strings.ToLower(item.PrimaryPackageName), qLower) ||
			query == "" {
			matched = append(matched, item)
			if limit > 0 && len(matched) >= limit {
				break
			}
		}
	}

	tookMs := time.Since(start).Milliseconds()

	return &FederatedSearchResponse{
		Query:  query,
		TookMs: tookMs,
		Items:  matched,
		Total:  len(matched),
	}, nil
}

func (h *AdapterHandler) CatalogFeed(ctx context.Context, cursor string, limit int, token string) (*CatalogFeedResponse, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.enabled {
		return nil, ErrFeatureDisabled
	}

	if token != h.serviceToken {
		return nil, ErrUnauthorizedService
	}

	if limit <= 0 {
		limit = 10
	}

	var results []CatalogFeedItem
	include := cursor == ""

	for _, item := range h.catalogItems {
		if !include {
			if item.Cursor == cursor {
				include = true // Start including after cursor
			}
			continue
		}

		results = append(results, item)
		if len(results) >= limit {
			break
		}
	}

	var nextCursor string
	hasMore := false
	if len(results) > 0 {
		nextCursor = results[len(results)-1].Cursor
		if len(results) == limit {
			hasMore = true
		}
	}

	return &CatalogFeedResponse{
		Items:      results,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

