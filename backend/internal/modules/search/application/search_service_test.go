package application

import (
	"context"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/search/domain"
)

// TC-FUNC-010: BM25 + Quality Hybrid Ranking & Autocomplete
func TestTCFUNC010SearchRankingAndAutocomplete(t *testing.T) {
	ctx := context.Background()
	engine := NewInMemoryFallbackSearchEngine()

	// Seed 3 test apps
	// App 1: Exact title match, High rating (4.8), Direct download available
	app1 := domain.SearchDoc{
		AppID:       "app-1",
		PackageName: "com.secure.authenticator",
		TitleI18n: map[string]string{
			"en": "Secure Authenticator 2FA",
			"ko": "시큐어 OTP 인증기",
		},
		DescriptionI18n: map[string]string{
			"en": "Two-factor authentication code generator.",
		},
		DeveloperName:           "Security Labs",
		Category:                "TOOLS",
		BayesianRating:          4.8,
		RatingCount:             12000,
		DownloadCount:           500000,
		DirectDownloadAvailable: true,
		RightsTier:              "TIER_1_DIRECT_MIRROR",
		PublishedAt:             time.Now().Add(-10 * 24 * time.Hour),
		Keywords:                []string{"authenticator", "otp", "2fa", "security"},
	}

	// App 2: Keyword match only, Low rating (3.0), Direct download available
	app2 := domain.SearchDoc{
		AppID:       "app-2",
		PackageName: "com.basic.tools",
		TitleI18n: map[string]string{
			"en": "Basic Utility Toolkit",
			"ko": "기본 유틸 툴킷",
		},
		DescriptionI18n: map[string]string{
			"en": "Contains simple 2fa generator and flashlight.",
		},
		DeveloperName:           "SimpleDev",
		Category:                "TOOLS",
		BayesianRating:          3.0,
		RatingCount:             50,
		DownloadCount:           1000,
		DirectDownloadAvailable: true,
		RightsTier:              "TIER_1_DIRECT_MIRROR",
		PublishedAt:             time.Now().Add(-100 * 24 * time.Hour),
		Keywords:                []string{"tools", "utility", "2fa"},
	}

	// App 3: Title match, High rating (4.9), External link only (Tier 5)
	app3 := domain.SearchDoc{
		AppID:       "app-3",
		PackageName: "com.external.authenticator",
		TitleI18n: map[string]string{
			"en": "Cloud Authenticator Pro",
			"ko": "클라우드 OTP 프로",
		},
		DescriptionI18n: map[string]string{
			"en": "Authenticator stored in cloud.",
		},
		DeveloperName:           "CloudCorp",
		Category:                "TOOLS",
		BayesianRating:          4.9,
		RatingCount:             20000,
		DownloadCount:           1000000,
		DirectDownloadAvailable: false, // External only
		RightsTier:              "TIER_5_EXTERNAL_LINK_ONLY",
		PublishedAt:             time.Now().Add(-5 * 24 * time.Hour),
		Keywords:                []string{"authenticator", "otp"},
	}

	_ = engine.IndexDocument(ctx, app1)
	_ = engine.IndexDocument(ctx, app2)
	_ = engine.IndexDocument(ctx, app3)

	t.Run("Sub-test 1: Hybrid Ranking prioritization (Title match + Direct download boost)", func(t *testing.T) {
		res, err := engine.Search(ctx, domain.SearchQuery{
			Keyword: "Authenticator",
			Locale:  "en",
			Limit:   10,
		})
		if err != nil {
			t.Fatalf("expected search success, got %v", err)
		}

		if len(res.Items) < 2 {
			t.Fatalf("expected at least 2 results, got %d", len(res.Items))
		}

		// App 1 has direct download boost + title match -> should rank #1
		if res.Items[0].AppID != "app-1" {
			t.Fatalf("expected app-1 to be ranked #1 due to direct download and quality boost, got %s (score=%.2f vs %.2f)",
				res.Items[0].AppID, res.Items[0].Score, res.Items[1].Score)
		}
	})

	t.Run("Sub-test 2: Fast Autocomplete Prefix Matching (<50ms)", func(t *testing.T) {
		start := time.Now()
		sug, err := engine.Suggest(ctx, "Sec", "en", 5)
		latency := time.Since(start)

		if err != nil {
			t.Fatalf("expected suggest success, got %v", err)
		}
		if latency > 50*time.Millisecond {
			t.Fatalf("autocomplete exceeded 50ms latency: took %v", latency)
		}

		if len(sug.Suggestions) == 0 {
			t.Fatalf("expected autocomplete suggestions for 'Sec'")
		}
		if sug.Suggestions[0].Text != "Secure Authenticator 2FA" {
			t.Fatalf("expected top suggestion 'Secure Authenticator 2FA', got %s", sug.Suggestions[0].Text)
		}
	})

	t.Run("Sub-test 3: DSL Query Builder generation", func(t *testing.T) {
		builder := NewSearchQueryBuilder()
		dsl, err := builder.BuildOpenSearchDSL(domain.SearchQuery{
			Keyword: "Authenticator",
			Locale:  "en",
			Limit:   20,
		})
		if err != nil {
			t.Fatalf("expected DSL build success, got %v", err)
		}
		if len(dsl) == 0 {
			t.Fatalf("expected non-empty OpenSearch DSL JSON")
		}
	})
}

// TC-OPS-005: OpenSearch Circuit Breaker & Automatic PostgreSQL Fallback
func TestTCOPS005SearchCircuitBreakerAndFallback(t *testing.T) {
	ctx := context.Background()

	primaryMock := NewInMemoryFallbackSearchEngine()
	fallbackMock := NewInMemoryFallbackSearchEngine()

	testDoc := domain.SearchDoc{
		AppID:       "app-fallback-1",
		PackageName: "com.fallback.test",
		TitleI18n: map[string]string{
			"en": "Resilience Search App",
		},
		DeveloperName:           "Resilience Team",
		Category:                "TOOLS",
		BayesianRating:          4.5,
		DirectDownloadAvailable: true,
		PublishedAt:             time.Now(),
	}

	_ = fallbackMock.IndexDocument(ctx, testDoc)

	service := NewSearchService(primaryMock, fallbackMock)

	t.Run("Sub-test 1: OpenSearch Failure triggers Fallback without HTTP error", func(t *testing.T) {
		// Simulate OpenSearch outage
		service.SetPrimarySimulatedDown(true)

		res, err := service.Search(ctx, domain.SearchQuery{
			Keyword: "Resilience",
			Locale:  "en",
			Limit:   10,
		})

		if err != nil {
			t.Fatalf("expected transparent fallback search without error, got %v", err)
		}

		if !res.FallbackUsed {
			t.Fatalf("expected FallbackUsed=true when primary is down")
		}

		if len(res.Items) != 1 || res.Items[0].AppID != "app-fallback-1" {
			t.Fatalf("expected fallback search to return app-fallback-1, got %d items", len(res.Items))
		}
	})

	t.Run("Sub-test 2: Autocomplete also falls back smoothly during outage", func(t *testing.T) {
		service.SetPrimarySimulatedDown(true)

		sug, err := service.Suggest(ctx, "Resil", "en", 5)
		if err != nil {
			t.Fatalf("expected suggest fallback success, got %v", err)
		}
		if len(sug.Suggestions) == 0 || sug.Suggestions[0].Text != "Resilience Search App" {
			t.Fatalf("expected fallback suggestion 'Resilience Search App', got %v", sug)
		}
	})
}
