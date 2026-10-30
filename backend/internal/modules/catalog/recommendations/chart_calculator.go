package recommendations

import (
	"math"
	"sort"
	"strings"
	"time"
)

type ChartType string

const (
	ChartTopFree       ChartType = "TOP_FREE"
	ChartTopPaid       ChartType = "TOP_PAID"
	ChartTrendingRising ChartType = "TRENDING_RISING"
	ChartNewNoteworthy  ChartType = "NEW_NOTEWORTHY"
)

type CatalogAppMetric struct {
	AppID           string
	Slug            string
	PackageName     string
	Title           string
	Category        string
	IsFree          bool
	PriceCents      int64
	RawRating       float64
	ReviewCount     int
	TotalDownloads  int64
	RecentDownloads int64 // Last 48 hours
	PublishedAt     time.Time
	CountryBlock    []string
	RightsLevel     string
	Tags            []string
}

type ChartEntry struct {
	Rank           int     `json:"rank"`
	AppID          string  `json:"app_id"`
	Slug           string  `json:"slug"`
	PackageName    string  `json:"package_name"`
	Title          string  `json:"title"`
	Category       string  `json:"category"`
	BayesianRating float64 `json:"bayesian_rating"`
	Score          float64 `json:"score"`
}

// ComputeBayesianRating computes the Bayesian weighted rating:
// WeightedRating = (v / (v + m)) * R + (m / (v + m)) * C
// m = 25 (minimum threshold), C = 4.1 (catalog prior mean)
func ComputeBayesianRating(rawRating float64, reviewCount int) float64 {
	const m = 25.0
	const C = 4.1

	v := float64(reviewCount)
	if v <= 0 {
		return C
	}

	weighted := (v / (v + m)) * rawRating + (m / (v + m)) * C
	return math.Round(weighted*100) / 100
}

type ChartCalculator struct {
	catalogMeanRating float64
	minReviews        float64
}

func NewChartCalculator() *ChartCalculator {
	return &ChartCalculator{
		catalogMeanRating: 4.1,
		minReviews:        25.0,
	}
}

// CalculateChart produces the ranked top N chart entries for a given ChartType
func (c *ChartCalculator) CalculateChart(chartType ChartType, apps []CatalogAppMetric, countryCode string, limit int) []ChartEntry {
	if limit <= 0 {
		limit = 50
	}

	country := strings.ToUpper(strings.TrimSpace(countryCode))
	var candidates []CatalogAppMetric

	// 1. Filter out country-blocked or prohibited apps
	for _, app := range apps {
		if app.RightsLevel == "LEVEL_5_RESTRICTED_PROHIBITED" {
			continue
		}
		isBlocked := false
		for _, b := range app.CountryBlock {
			if strings.EqualFold(b, country) {
				isBlocked = true
				break
			}
		}
		if isBlocked {
			continue
		}
		candidates = append(candidates, app)
	}

	type ScoredApp struct {
		App            CatalogAppMetric
		BayesianRating float64
		Score          float64
	}

	var scored []ScoredApp

	now := time.Now().UTC()

	for _, app := range candidates {
		bayesian := ComputeBayesianRating(app.RawRating, app.ReviewCount)
		score := 0.0

		switch chartType {
		case ChartTopFree:
			if !app.IsFree {
				continue
			}
			// Downloads weighted by rating
			score = float64(app.TotalDownloads) * (bayesian / 5.0)

		case ChartTopPaid:
			if app.IsFree {
				continue
			}
			score = float64(app.TotalDownloads) * (bayesian / 5.0)

		case ChartTrendingRising:
			// Velocity: recent 48h downloads acceleration
			score = float64(app.RecentDownloads) * 2.0 * (bayesian / 5.0)

		case ChartNewNoteworthy:
			// Only apps published within 30 days
			daysSince := now.Sub(app.PublishedAt).Hours() / 24.0
			if daysSince > 30.0 {
				continue
			}
			// Cold-start freshness boost + Bayesian rating
			freshnessBoost := math.Max(0.1, 1.0-(daysSince/30.0)*0.5)
			score = bayesian * freshnessBoost * math.Log10(float64(app.TotalDownloads+10))
		}

		scored = append(scored, ScoredApp{
			App:            app,
			BayesianRating: bayesian,
			Score:          score,
		})
	}

	// Sort descending by score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	results := make([]ChartEntry, 0, len(scored))
	for i, s := range scored {
		if i >= limit {
			break
		}
		results = append(results, ChartEntry{
			Rank:           i + 1,
			AppID:          s.App.AppID,
			Slug:           s.App.Slug,
			PackageName:    s.App.PackageName,
			Title:          s.App.Title,
			Category:       s.App.Category,
			BayesianRating: s.BayesianRating,
			Score:          s.Score,
		})
	}

	return results
}

