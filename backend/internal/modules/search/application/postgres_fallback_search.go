package application

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/search/domain"
)

type FallbackSearchEngine interface {
	Search(ctx context.Context, q domain.SearchQuery) (*domain.SearchResult, error)
	Suggest(ctx context.Context, prefix string, locale string, limit int) (*domain.SuggestResult, error)
	IndexDocument(ctx context.Context, doc domain.SearchDoc) error
}

type InMemoryFallbackSearchEngine struct {
	mu   sync.RWMutex
	docs map[string]domain.SearchDoc
}

func NewInMemoryFallbackSearchEngine() *InMemoryFallbackSearchEngine {
	return &InMemoryFallbackSearchEngine{
		docs: make(map[string]domain.SearchDoc),
	}
}

func (e *InMemoryFallbackSearchEngine) IndexDocument(ctx context.Context, doc domain.SearchDoc) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.docs[doc.AppID] = doc
	return nil
}

func (e *InMemoryFallbackSearchEngine) Search(ctx context.Context, q domain.SearchQuery) (*domain.SearchResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	start := time.Now()
	normKw := strings.ToLower(strings.TrimSpace(q.Keyword))
	locale := q.Locale
	if locale == "" {
		locale = "en"
	}

	var matched []domain.SearchDoc
	builder := NewSearchQueryBuilder()

	for _, doc := range e.docs {
		if q.Category != "" && !strings.EqualFold(doc.Category, q.Category) {
			continue
		}
		if q.DirectDownloadOnly && !doc.DirectDownloadAvailable {
			continue
		}

		score := builder.ComputeDeterministicScore(doc, normKw, locale)
		if score > 0 {
			docCopy := doc
			docCopy.Score = score
			matched = append(matched, docCopy)
		}
	}

	// Sort by descending hybrid score
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Score > matched[j].Score
	})

	total := int64(len(matched))
	startIdx := q.Offset
	if startIdx > len(matched) {
		startIdx = len(matched)
	}
	endIdx := startIdx + q.Limit
	if q.Limit <= 0 {
		endIdx = startIdx + 20
	}
	if endIdx > len(matched) {
		endIdx = len(matched)
	}

	return &domain.SearchResult{
		Items:        matched[startIdx:endIdx],
		TotalCount:   total,
		TookMs:       time.Since(start).Milliseconds(),
		FallbackUsed: true, // Marked as fallback search
	}, nil
}

func (e *InMemoryFallbackSearchEngine) Suggest(ctx context.Context, prefix string, locale string, limit int) (*domain.SuggestResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	start := time.Now()
	normPrefix := strings.ToLower(strings.TrimSpace(prefix))
	if normPrefix == "" {
		return &domain.SuggestResult{Suggestions: []domain.SuggestionItem{}, TookMs: 0}, nil
	}

	if limit <= 0 {
		limit = 5
	}

	var suggestions []domain.SuggestionItem
	seen := make(map[string]bool)

	for _, doc := range e.docs {
		// 1. Check title
		title := ""
		if t, ok := doc.TitleI18n[locale]; ok {
			title = t
		} else if tEn, ok := doc.TitleI18n["en"]; ok {
			title = tEn
		}

		if strings.HasPrefix(strings.ToLower(title), normPrefix) && !seen[title] {
			seen[title] = true
			suggestions = append(suggestions, domain.SuggestionItem{
				Text:    title,
				Type:    "APP",
				Package: doc.PackageName,
			})
			if len(suggestions) >= limit {
				break
			}
		}

		// 2. Check keywords
		for _, kw := range doc.Keywords {
			if strings.HasPrefix(strings.ToLower(kw), normPrefix) && !seen[kw] {
				seen[kw] = true
				suggestions = append(suggestions, domain.SuggestionItem{
					Text: kw,
					Type: "KEYWORD",
				})
				if len(suggestions) >= limit {
					break
				}
			}
		}
		if len(suggestions) >= limit {
			break
		}
	}

	return &domain.SuggestResult{
		Suggestions: suggestions,
		TookMs:      time.Since(start).Milliseconds(),
	}, nil
}
