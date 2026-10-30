package analytics_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/analytics"
)

func TestFunnelAnalytics_TC_FUNC_014_EventIngestionAndRollup(t *testing.T) {
	ctx := context.Background()
	store := analytics.NewMemoryEventStore()
	metricsStore := analytics.NewMemoryMetricsStore()
	ingestor := analytics.NewFunnelIngestor(store)
	worker := analytics.NewRollupWorker(store, metricsStore)

	appID := "app-test-prod-100"
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)

	// Prepare 100 Search Impressions
	var batch []analytics.AnalyticsEventEnvelope
	for i := 0; i < 100; i++ {
		batch = append(batch, analytics.AnalyticsEventEnvelope{
			EventID:     fmt.Sprintf("evt-imp-%d", i),
			EventName:   analytics.EventSearchResultView,
			OccurredAt:  now,
			SessionHash: fmt.Sprintf("session-%d", i),
			Platform:    "ANDROID_CLIENT",
			CountryCode: "KR",
			Properties: analytics.EventProperties{
				AppID: appID,
			},
		})
	}

	// 35 Detail Views
	for i := 0; i < 35; i++ {
		batch = append(batch, analytics.AnalyticsEventEnvelope{
			EventID:     fmt.Sprintf("evt-det-%d", i),
			EventName:   analytics.EventDetailView,
			OccurredAt:  now,
			SessionHash: fmt.Sprintf("session-%d", i),
			Platform:    "ANDROID_CLIENT",
			CountryCode: "KR",
			Properties: analytics.EventProperties{
				AppID: appID,
			},
		})
	}

	// 10 Downloads Started
	for i := 0; i < 10; i++ {
		batch = append(batch, analytics.AnalyticsEventEnvelope{
			EventID:     fmt.Sprintf("evt-dl-start-%d", i),
			EventName:   analytics.EventDownloadStarted,
			OccurredAt:  now,
			SessionHash: fmt.Sprintf("session-%d", i),
			Platform:    "ANDROID_CLIENT",
			CountryCode: "KR",
			Properties: analytics.EventProperties{
				AppID: appID,
			},
		})
	}

	// 10 Downloads Completed
	for i := 0; i < 10; i++ {
		batch = append(batch, analytics.AnalyticsEventEnvelope{
			EventID:     fmt.Sprintf("evt-dl-end-%d", i),
			EventName:   analytics.EventDownloadCompleted,
			OccurredAt:  now,
			SessionHash: fmt.Sprintf("session-%d", i),
			Platform:    "ANDROID_CLIENT",
			CountryCode: "KR",
			Properties: analytics.EventProperties{
				AppID: appID,
			},
		})
	}

	// 9 Installs Succeeded (9 / 10 = 90% success rate)
	for i := 0; i < 9; i++ {
		batch = append(batch, analytics.AnalyticsEventEnvelope{
			EventID:     fmt.Sprintf("evt-inst-%d", i),
			EventName:   analytics.EventInstallSucceeded,
			OccurredAt:  now,
			SessionHash: fmt.Sprintf("session-%d", i),
			Platform:    "ANDROID_CLIENT",
			CountryCode: "KR",
			Properties: analytics.EventProperties{
				AppID: appID,
			},
		})
	}

	// 2 Purchases Completed (Total: 15,000 KRW/cents)
	batch = append(batch, analytics.AnalyticsEventEnvelope{
		EventID:     "evt-buy-1",
		EventName:   analytics.EventPurchaseCompleted,
		OccurredAt:  now,
		SessionHash: "session-1",
		Platform:    "ANDROID_CLIENT",
		CountryCode: "KR",
		Properties: analytics.EventProperties{
			AppID:      appID,
			PriceCents: 5000,
		},
	})
	batch = append(batch, analytics.AnalyticsEventEnvelope{
		EventID:     "evt-buy-2",
		EventName:   analytics.EventPurchaseCompleted,
		OccurredAt:  now,
		SessionHash: "session-2",
		Platform:    "ANDROID_CLIENT",
		CountryCode: "KR",
		Properties: analytics.EventProperties{
			AppID:      appID,
			PriceCents: 10000,
		},
	})

	// 1. Test Ingestion Batch
	accepted, err := ingestor.IngestBatch(ctx, batch)
	if err != nil {
		t.Fatalf("failed to ingest batch: %v", err)
	}
	expectedCount := 100 + 35 + 10 + 10 + 9 + 2
	if accepted != expectedCount {
		t.Fatalf("expected %d events accepted, got %d", expectedCount, accepted)
	}

	// 2. Test Ingest Single Install Succeeded via dedicated endpoint
	err = ingestor.IngestInstall(ctx, analytics.AnalyticsEventEnvelope{
		EventID:     "evt-inst-standalone",
		SessionHash: "session-standalone",
		Platform:    "ANDROID_CLIENT",
		CountryCode: "KR",
		Properties: analytics.EventProperties{
			AppID: appID,
		},
		OccurredAt: now,
	})
	if err != nil {
		t.Fatalf("failed to ingest standalone install: %v", err)
	}

	// 3. Run Daily Rollup Worker
	rolledUp, err := worker.RunRollup(ctx)
	if err != nil {
		t.Fatalf("rollup worker failed: %v", err)
	}
	if rolledUp != 1 {
		t.Fatalf("expected 1 daily metric bucket aggregated, got %d", rolledUp)
	}

	// 4. Query Funnel Summary and Verify Exact KPI Numbers
	dateStr := now.Format("2006-01-02")
	summary, err := worker.GetAppFunnelSummary(ctx, appID, dateStr, dateStr)
	if err != nil {
		t.Fatalf("failed to query funnel summary: %v", err)
	}

	if summary.TotalImpressions != 100 {
		t.Errorf("expected 100 impressions, got %d", summary.TotalImpressions)
	}
	if summary.TotalDetailViews != 35 {
		t.Errorf("expected 35 detail views, got %d", summary.TotalDetailViews)
	}
	if summary.TotalDownloadsStarted != 10 {
		t.Errorf("expected 10 downloads started, got %d", summary.TotalDownloadsStarted)
	}
	if summary.TotalDownloadsCompleted != 10 {
		t.Errorf("expected 10 downloads completed, got %d", summary.TotalDownloadsCompleted)
	}
	// 9 from batch + 1 standalone = 10 installs
	if summary.TotalInstallsSucceeded != 10 {
		t.Errorf("expected 10 installs succeeded, got %d", summary.TotalInstallsSucceeded)
	}
	if summary.TotalPurchasesCompleted != 2 {
		t.Errorf("expected 2 purchases completed, got %d", summary.TotalPurchasesCompleted)
	}
	if summary.TotalRevenueCents != 15000 {
		t.Errorf("expected 15000 cents total revenue, got %d", summary.TotalRevenueCents)
	}

	// CTR: 35 / 100 = 0.35 (35%)
	if math.Abs(summary.OverallSearchToDetailCTR-0.35) > 1e-4 {
		t.Errorf("expected CTR = 0.35, got %f", summary.OverallSearchToDetailCTR)
	}

	// CVR: 10 / 35 = ~0.285714 (28.57%)
	expectedCVR := 10.0 / 35.0
	if math.Abs(summary.OverallDetailToDownloadCVR-expectedCVR) > 1e-4 {
		t.Errorf("expected CVR = %f, got %f", expectedCVR, summary.OverallDetailToDownloadCVR)
	}

	// Install Success Rate: 10 / 10 = 1.0 (100%)
	if math.Abs(summary.OverallInstallSuccessRate-1.0) > 1e-4 {
		t.Errorf("expected Install Success Rate = 1.0, got %f", summary.OverallInstallSuccessRate)
	}
}
