package recommendations

import (
	"testing"
	"time"
)

// TC-FUNC-011: Top 4 Charts, Bayesian Weighted Rating & Similar Apps Recommendation
func TestTCFUNC011ChartsAndRecommendations(t *testing.T) {
	calc := NewChartCalculator()

	t.Run("Sub-test 1: Bayesian Rating prevents 1-review manipulation distortion", func(t *testing.T) {
		singleReviewRating := ComputeBayesianRating(5.0, 1)
		establishedRating := ComputeBayesianRating(4.8, 10000)

		// 1 review of 5.0 should be dragged towards prior mean (4.1) -> ~4.13
		if singleReviewRating > 4.25 {
			t.Fatalf("expected 1-review rating to be dampened towards 4.1, got %.2f", singleReviewRating)
		}

		// 10,000 reviews of 4.8 should stay ~4.80
		if establishedRating < 4.75 {
			t.Fatalf("expected 10,000 review rating to stay ~4.80, got %.2f", establishedRating)
		}

		if establishedRating <= singleReviewRating {
			t.Fatalf("established app should decisively outrank 1-review app: %.2f vs %.2f",
				establishedRating, singleReviewRating)
		}
	})

	t.Run("Sub-test 2: NEW_NOTEWORTHY Chart strictly filters out apps older than 30 days", func(t *testing.T) {
		apps := []CatalogAppMetric{
			{
				AppID:       "app-new",
				Title:       "Brand New Quality App",
				Category:    "FINANCE",
				IsFree:      true,
				RawRating:   4.8,
				ReviewCount: 150,
				PublishedAt: time.Now().Add(-5 * 24 * time.Hour), // 5 days old
			},
			{
				AppID:       "app-old",
				Title:       "Ancient Popular App",
				Category:    "FINANCE",
				IsFree:      true,
				RawRating:   4.9,
				ReviewCount: 50000,
				PublishedAt: time.Now().Add(-120 * 24 * time.Hour), // 120 days old
			},
		}

		entries := calc.CalculateChart(ChartNewNoteworthy, apps, "KR", 10)
		if len(entries) != 1 {
			t.Fatalf("expected only 1 app in NEW_NOTEWORTHY, got %d", len(entries))
		}
		if entries[0].AppID != "app-new" {
			t.Fatalf("expected app-new, got %s", entries[0].AppID)
		}
	})

	t.Run("Sub-test 3: TRENDING_RISING Chart ranks by 48-hour velocity", func(t *testing.T) {
		apps := []CatalogAppMetric{
			{
				AppID:           "app-steady",
				Title:           "Steady App",
				Category:        "GAMES",
				IsFree:          true,
				TotalDownloads:  1000000,
				RecentDownloads: 50, // Low recent acceleration
				RawRating:       4.5,
				ReviewCount:     200,
			},
			{
				AppID:           "app-viral",
				Title:           "Viral Breakout App",
				Category:        "GAMES",
				IsFree:          true,
				TotalDownloads:  5000,
				RecentDownloads: 4500, // Very high recent acceleration
				RawRating:       4.5,
				ReviewCount:     200,
			},
		}

		entries := calc.CalculateChart(ChartTrendingRising, apps, "KR", 10)
		if len(entries) < 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].AppID != "app-viral" {
			t.Fatalf("expected viral app to rank #1 in TRENDING_RISING, got %s", entries[0].AppID)
		}
	})

	t.Run("Sub-test 4: Similar Apps service ranks by category and tag overlap", func(t *testing.T) {
		simService := NewSimilarAppsService()

		target := CatalogAppMetric{
			AppID:       "target-wallet",
			PackageName: "com.target.wallet",
			Title:       "Target Crypto Wallet",
			Category:    "FINANCE",
			Tags:        []string{"crypto", "bitcoin", "wallet"},
			RawRating:   4.5,
			ReviewCount: 500,
		}

		catalog := []CatalogAppMetric{
			{
				AppID:       "similar-wallet",
				PackageName: "com.defi.wallet",
				Title:       "DeFi Web3 Wallet",
				Category:    "FINANCE", // Same category
				Tags:        []string{"crypto", "wallet", "ethereum"}, // 2 shared tags
				RawRating:   4.6,
				ReviewCount: 800,
			},
			{
				AppID:       "unrelated-racing",
				PackageName: "com.turbo.racing",
				Title:       "Turbo Racing 3D",
				Category:    "GAMES", // Different category
				Tags:        []string{"racing", "cars"},
				RawRating:   4.9,
				ReviewCount: 10000,
			},
		}

		recommendations := simService.FindSimilarApps(target, catalog, 5)
		if len(recommendations) != 1 {
			t.Fatalf("expected exactly 1 similar app, got %d", len(recommendations))
		}
		if recommendations[0].AppID != "similar-wallet" {
			t.Fatalf("expected similar-wallet, got %s", recommendations[0].AppID)
		}
	})
}
