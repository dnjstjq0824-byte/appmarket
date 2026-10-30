package analytics

import (
	"context"
	"fmt"
	"sync"
)

type AppMetricsDaily struct {
	AppID                       string  `json:"app_id"`
	Date                        string  `json:"date"` // YYYY-MM-DD
	Platform                    string  `json:"platform"`
	CountryCode                 string  `json:"country_code"`
	Impressions                 int64   `json:"impressions"`
	DetailViews                 int64   `json:"detail_views"`
	DownloadsStarted            int64   `json:"downloads_started"`
	DownloadsCompleted          int64   `json:"downloads_completed"`
	InstallsSucceeded           int64   `json:"installs_succeeded"`
	PurchasesCompleted          int64   `json:"purchases_completed"`
	TotalRevenueCents           int64   `json:"total_revenue_cents"`
	SearchToDetailCTR           float64 `json:"search_to_detail_ctr"`
	DetailToDownloadCVR         float64 `json:"detail_to_download_cvr"`
	DownloadToInstallSuccessRate float64 `json:"download_to_install_success_rate"`
	IsCompatibilityAlert        bool    `json:"is_compatibility_alert"`
}

type MetricsStore interface {
	UpsertMetrics(ctx context.Context, metrics []*AppMetricsDaily) error
	QueryMetrics(ctx context.Context, appID string, fromDate, toDate string) ([]*AppMetricsDaily, error)
}

type MemoryMetricsStore struct {
	mu      sync.RWMutex
	metrics map[string]*AppMetricsDaily
}

func NewMemoryMetricsStore() *MemoryMetricsStore {
	return &MemoryMetricsStore{
		metrics: make(map[string]*AppMetricsDaily),
	}
}

func (s *MemoryMetricsStore) UpsertMetrics(ctx context.Context, metrics []*AppMetricsDaily) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range metrics {
		key := fmt.Sprintf("%s:%s:%s:%s", m.AppID, m.Date, m.Platform, m.CountryCode)
		s.metrics[key] = m
	}
	return nil
}

func (s *MemoryMetricsStore) QueryMetrics(ctx context.Context, appID string, fromDate, toDate string) ([]*AppMetricsDaily, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var res []*AppMetricsDaily
	for _, m := range s.metrics {
		if m.AppID == appID {
			if fromDate != "" && m.Date < fromDate {
				continue
			}
			if toDate != "" && m.Date > toDate {
				continue
			}
			res = append(res, m)
		}
	}
	return res, nil
}

type RollupWorker struct {
	eventStore   EventStore
	metricsStore MetricsStore
}

func NewRollupWorker(eventStore EventStore, metricsStore MetricsStore) *RollupWorker {
	return &RollupWorker{
		eventStore:   eventStore,
		metricsStore: metricsStore,
	}
}

func (rw *RollupWorker) AggregateEvents(events []AnalyticsEventEnvelope) []*AppMetricsDaily {
	grouped := make(map[string]*AppMetricsDaily)

	for _, e := range events {
		dateStr := e.OccurredAt.UTC().Format("2006-01-02")
		key := fmt.Sprintf("%s:%s:%s:%s", e.Properties.AppID, dateStr, e.Platform, e.CountryCode)

		m, exists := grouped[key]
		if !exists {
			m = &AppMetricsDaily{
				AppID:       e.Properties.AppID,
				Date:        dateStr,
				Platform:    e.Platform,
				CountryCode: e.CountryCode,
			}
			grouped[key] = m
		}

		switch e.EventName {
		case EventSearchResultView:
			m.Impressions++
		case EventDetailView:
			m.DetailViews++
		case EventDownloadStarted:
			m.DownloadsStarted++
		case EventDownloadCompleted:
			m.DownloadsCompleted++
		case EventInstallSucceeded:
			m.InstallsSucceeded++
		case EventPurchaseCompleted:
			m.PurchasesCompleted++
			m.TotalRevenueCents += e.Properties.PriceCents
		}
	}

	result := make([]*AppMetricsDaily, 0, len(grouped))
	for _, m := range grouped {
		if m.Impressions > 0 {
			m.SearchToDetailCTR = float64(m.DetailViews) / float64(m.Impressions)
		}
		if m.DetailViews > 0 {
			m.DetailToDownloadCVR = float64(m.DownloadsStarted) / float64(m.DetailViews)
		}
		if m.DownloadsCompleted > 0 {
			m.DownloadToInstallSuccessRate = float64(m.InstallsSucceeded) / float64(m.DownloadsCompleted)
			if m.DownloadsCompleted >= 5 && m.DownloadToInstallSuccessRate < 0.90 {
				m.IsCompatibilityAlert = true
			}
		}
		result = append(result, m)
	}

	return result
}

func (rw *RollupWorker) RunRollup(ctx context.Context) (int, error) {
	events, err := rw.eventStore.GetAllEvents(ctx)
	if err != nil {
		return 0, err
	}

	aggregated := rw.AggregateEvents(events)
	if len(aggregated) > 0 {
		if err := rw.metricsStore.UpsertMetrics(ctx, aggregated); err != nil {
			return 0, err
		}
	}
	return len(aggregated), nil
}

type FunnelSummary struct {
	TotalImpressions           int64   `json:"total_impressions"`
	TotalDetailViews           int64   `json:"total_detail_views"`
	TotalDownloadsStarted      int64   `json:"total_downloads_started"`
	TotalDownloadsCompleted    int64   `json:"total_downloads_completed"`
	TotalInstallsSucceeded     int64   `json:"total_installs_succeeded"`
	TotalPurchasesCompleted    int64   `json:"total_purchases_completed"`
	TotalRevenueCents          int64   `json:"total_revenue_cents"`
	OverallSearchToDetailCTR   float64 `json:"overall_search_to_detail_ctr"`
	OverallDetailToDownloadCVR float64 `json:"overall_detail_to_download_cvr"`
	OverallInstallSuccessRate  float64 `json:"overall_install_success_rate"`
	DailyBreakdown             []*AppMetricsDaily `json:"daily_breakdown"`
}

func (rw *RollupWorker) GetAppFunnelSummary(ctx context.Context, appID string, fromDate, toDate string) (*FunnelSummary, error) {
	daily, err := rw.metricsStore.QueryMetrics(ctx, appID, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	summary := &FunnelSummary{
		DailyBreakdown: daily,
	}

	for _, d := range daily {
		summary.TotalImpressions += d.Impressions
		summary.TotalDetailViews += d.DetailViews
		summary.TotalDownloadsStarted += d.DownloadsStarted
		summary.TotalDownloadsCompleted += d.DownloadsCompleted
		summary.TotalInstallsSucceeded += d.InstallsSucceeded
		summary.TotalPurchasesCompleted += d.PurchasesCompleted
		summary.TotalRevenueCents += d.TotalRevenueCents
	}

	if summary.TotalImpressions > 0 {
		summary.OverallSearchToDetailCTR = float64(summary.TotalDetailViews) / float64(summary.TotalImpressions)
	}
	if summary.TotalDetailViews > 0 {
		summary.OverallDetailToDownloadCVR = float64(summary.TotalDownloadsStarted) / float64(summary.TotalDetailViews)
	}
	if summary.TotalDownloadsCompleted > 0 {
		summary.OverallInstallSuccessRate = float64(summary.TotalInstallsSucceeded) / float64(summary.TotalDownloadsCompleted)
	}

	return summary, nil
}
