package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/search/domain"
)

// OpenSearchEngine queries an OpenSearch 2.x cluster over plain HTTPS.
// Any transport or query failure returns an error so SearchService falls back
// to the in-memory engine; this type never fabricates results.
type OpenSearchEngine struct {
	baseURL string
	user    string
	pass    string
	index   string
	http    *http.Client
}

func NewOpenSearchEngine(baseURL, user, pass, index string, httpClient *http.Client) *OpenSearchEngine {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	if index == "" {
		index = "apps"
	}
	return &OpenSearchEngine{
		baseURL: strings.TrimRight(baseURL, "/"),
		user:    user,
		pass:    pass,
		index:   index,
		http:    httpClient,
	}
}

type openSearchHits struct {
	Took int64 `json:"took"`
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Items []struct {
			Score  float64         `json:"_score"`
			Source domain.SearchDoc `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (e *OpenSearchEngine) do(ctx context.Context, method, path string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, e.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if e.user != "" {
		req.SetBasicAuth(e.user, e.pass)
	}
	resp, err := e.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("opensearch request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("opensearch status %d", resp.StatusCode)
	}
	return raw, nil
}

// Search runs a BM25 multi-match across title, description, package, and
// developer fields with optional category and direct-download filters.
func (e *OpenSearchEngine) Search(ctx context.Context, q domain.SearchQuery) (*domain.SearchResult, error) {
	must := []any{
		map[string]any{"multi_match": map[string]any{
			"query":  q.Keyword,
			"fields": []string{"title_i18n.*^3", "package_name^2", "developer_name^2", "description_i18n.*", "keywords"},
		}},
	}
	var filters []any
	if q.Category != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"category": q.Category}})
	}
	if q.DirectDownloadOnly {
		filters = append(filters, map[string]any{"term": map[string]any{"direct_download_available": true}})
	}
	limit := q.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	query := map[string]any{
		"size": limit,
		"from": max(q.Offset, 0),
		"query": map[string]any{"bool": map[string]any{
			"must":   must,
			"filter": filters,
		}},
	}
	raw, err := e.do(ctx, http.MethodPost, "/"+e.index+"/_search", query)
	if err != nil {
		return nil, err
	}
	var parsed openSearchHits
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("opensearch decode failed: %w", err)
	}
	res := &domain.SearchResult{TotalCount: parsed.Hits.Total.Value, TookMs: parsed.Took}
	for _, hit := range parsed.Hits.Items {
		doc := hit.Source
		doc.Score = hit.Score
		res.Items = append(res.Items, doc)
	}
	return res, nil
}

// Suggest returns prefix completions over titles and developer names.
func (e *OpenSearchEngine) Suggest(ctx context.Context, prefix, locale string, limit int) (*domain.SuggestResult, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	query := map[string]any{
		"size": limit,
		"_source": []string{"package_name", "developer_name", "title_i18n"},
		"query": map[string]any{"bool": map[string]any{"should": []any{
			map[string]any{"match_phrase_prefix": map[string]any{"title_i18n." + suggestLocale(locale): prefix}},
			map[string]any{"match_phrase_prefix": map[string]any{"developer_name": prefix}},
			map[string]any{"match_phrase_prefix": map[string]any{"package_name": prefix}},
		}}},
	}
	raw, err := e.do(ctx, http.MethodPost, "/"+e.index+"/_search", query)
	if err != nil {
		return nil, err
	}
	var parsed openSearchHits
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("opensearch decode failed: %w", err)
	}
	res := &domain.SuggestResult{TookMs: parsed.Took}
	seen := make(map[string]bool)
	for _, hit := range parsed.Hits.Items {
		title := hit.Source.TitleI18n[locale]
		if title == "" {
			title = hit.Source.TitleI18n["en-US"]
		}
		if title != "" && !seen["a:"+title] {
			seen["a:"+title] = true
			res.Suggestions = append(res.Suggestions, domain.SuggestionItem{Text: title, Type: "APP", Package: hit.Source.PackageName})
		}
		if hit.Source.DeveloperName != "" && !seen["d:"+hit.Source.DeveloperName] {
			seen["d:"+hit.Source.DeveloperName] = true
			res.Suggestions = append(res.Suggestions, domain.SuggestionItem{Text: hit.Source.DeveloperName, Type: "DEVELOPER"})
		}
		if len(res.Suggestions) >= limit {
			break
		}
	}
	return res, nil
}

func suggestLocale(locale string) string {
	if locale == "" {
		return "en-US"
	}
	return locale
}

// IndexDocument upserts one app document by app id.
func (e *OpenSearchEngine) IndexDocument(ctx context.Context, doc domain.SearchDoc) error {
	if doc.AppID == "" {
		return fmt.Errorf("app_id is required for indexing")
	}
	_, err := e.do(ctx, http.MethodPut, "/"+e.index+"/_doc/"+doc.AppID, doc)
	return err
}
