package application

import (
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/search/domain"
)

type SearchQueryBuilder struct{}

func NewSearchQueryBuilder() *SearchQueryBuilder {
	return &SearchQueryBuilder{}
}

// BuildOpenSearchDSL generates the production OpenSearch 2.x JSON DSL with multi_match and function_score
func (b *SearchQueryBuilder) BuildOpenSearchDSL(q domain.SearchQuery) (string, error) {
	locale := q.Locale
	if locale == "" {
		locale = "en"
	}

	titleField := "title_i18n." + locale
	descField := "description_i18n." + locale

	dslMap := map[string]interface{}{
		"from": q.Offset,
		"size": q.Limit,
		"query": map[string]interface{}{
			"function_score": map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": []interface{}{
							map[string]interface{}{
								"multi_match": map[string]interface{}{
									"query":  q.Keyword,
									"fields": []string{titleField + "^3.0", "keywords^2.0", "developer_name^1.5", descField + "^1.0"},
									"type":   "best_fields",
								},
							},
						},
					},
				},
				"functions": []interface{}{
					// Bayesian rating factor
					map[string]interface{}{
						"field_value_factor": map[string]interface{}{
							"field":    "bayesian_rating",
							"factor":   0.2,
							"modifier": "log1p",
							"missing":  2.5,
						},
						"weight": 1.5,
					},
					// Direct download availability boost (honest store priority)
					map[string]interface{}{
						"filter": map[string]interface{}{
							"term": map[string]interface{}{"direct_download_available": true},
						},
						"weight": 1.25,
					},
				},
				"score_mode": "multiply",
				"boost_mode": "multiply",
			},
		},
	}

	bytes, err := json.Marshal(dslMap)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ComputeDeterministicScore calculates the BM25 + Quality hybrid score for in-memory ranking & tests
func (b *SearchQueryBuilder) ComputeDeterministicScore(doc domain.SearchDoc, keyword, locale string) float64 {
	if keyword == "" {
		return doc.BayesianRating
	}

	normKw := strings.ToLower(strings.TrimSpace(keyword))
	var matchScore float64 = 0.0

	// Check Title (boost 3.0)
	title := ""
	if t, exists := doc.TitleI18n[locale]; exists {
		title = t
	} else if tEn, exists := doc.TitleI18n["en"]; exists {
		title = tEn
	}
	if strings.Contains(strings.ToLower(title), normKw) {
		matchScore += 3.0
	}

	// Check Keywords (boost 2.0)
	for _, kw := range doc.Keywords {
		if strings.Contains(strings.ToLower(kw), normKw) {
			matchScore += 2.0
			break
		}
	}

	// Check Developer Name (boost 1.5)
	if strings.Contains(strings.ToLower(doc.DeveloperName), normKw) {
		matchScore += 1.5
	}

	// Check Description (boost 1.0)
	desc := ""
	if d, exists := doc.DescriptionI18n[locale]; exists {
		desc = d
	}
	if strings.Contains(strings.ToLower(desc), normKw) {
		matchScore += 1.0
	}

	if matchScore == 0.0 {
		return 0.0
	}

	// Quality Multipliers:
	// 1. Bayesian Rating multiplier (1.0 to 1.5)
	ratingMultiplier := 1.0 + (doc.BayesianRating / 5.0) * 0.5

	// 2. Direct download priority (1.25x if directly hosted on CDN vs external outlink)
	directMultiplier := 1.0
	if doc.DirectDownloadAvailable {
		directMultiplier = 1.25
	}

	// 3. Freshness decay (within 180 days)
	daysSincePublished := time.Since(doc.PublishedAt).Hours() / 24.0
	freshnessMultiplier := math.Max(0.8, 1.0-(daysSincePublished/365.0)*0.2)

	return matchScore * ratingMultiplier * directMultiplier * freshnessMultiplier
}

