package domain

import (
	"errors"
	"time"
)

var (
	ErrSearchEngineUnavailable = errors.New("ERR_SEARCH_ENGINE_UNAVAILABLE")
	ErrInvalidSearchQuery      = errors.New("ERR_INVALID_SEARCH_QUERY")
)

type SearchDoc struct {
	AppID                   string            `json:"app_id"`
	PackageName             string            `json:"package_name"`
	TitleI18n               map[string]string `json:"title_i18n"`
	DescriptionI18n         map[string]string `json:"description_i18n"`
	DeveloperName           string            `json:"developer_name"`
	Category                string            `json:"category"`
	BayesianRating          float64           `json:"bayesian_rating"`
	RatingCount             int               `json:"rating_count"`
	DownloadCount           int64             `json:"download_count"`
	DirectDownloadAvailable bool              `json:"direct_download_available"`
	RightsTier              string            `json:"rights_tier"`
	PublishedAt             time.Time         `json:"published_at"`
	Keywords                []string          `json:"keywords"`
	Score                   float64           `json:"score,omitempty"`
}

type SearchQuery struct {
	Keyword                 string `json:"keyword"`
	Locale                  string `json:"locale"`
	Category                string `json:"category"`
	DirectDownloadOnly      bool   `json:"direct_download_only"`
	Limit                   int    `json:"limit"`
	Offset                  int    `json:"offset"`
}

type SearchResult struct {
	Items        []SearchDoc `json:"items"`
	TotalCount   int64       `json:"total_count"`
	TookMs       int64       `json:"took_ms"`
	FallbackUsed bool        `json:"fallback_used"`
}

type SuggestionItem struct {
	Text      string `json:"text"`
	Type      string `json:"type"` // "APP" | "DEVELOPER" | "KEYWORD"
	Package   string `json:"package,omitempty"`
}

type SuggestResult struct {
	Suggestions []SuggestionItem `json:"suggestions"`
	TookMs      int64            `json:"took_ms"`
}
