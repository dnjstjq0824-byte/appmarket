package analytics

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInvalidEventName = errors.New("invalid or empty event_name")
	ErrMissingAppID     = errors.New("missing app_id in event properties")
	ErrEmptyBatch       = errors.New("empty event batch")
)

const (
	EventSearchResultView  = "app.search.result_view"
	EventDetailView        = "app.detail.view"
	EventDownloadStarted   = "app.download.started"
	EventDownloadCompleted = "app.download.completed"
	EventInstallSucceeded  = "app.install.succeeded"
	EventPurchaseCompleted = "app.purchase.completed"
)

type DeviceInfo struct {
	SDKInt     int    `json:"sdk_int"`
	ABI        string `json:"abi"`
	FormFactor string `json:"form_factor"`
}

type EventProperties struct {
	AppID          string `json:"app_id"`
	DistributionID string `json:"distribution_id,omitempty"`
	ReleaseID      string `json:"release_id,omitempty"`
	SourceType     string `json:"source_type,omitempty"`
	DurationMs     int64  `json:"duration_ms,omitempty"`
	PriceCents     int64  `json:"price_cents,omitempty"`
}

type AnalyticsEventEnvelope struct {
	EventID     string          `json:"event_id"`
	EventName   string          `json:"event_name"`
	OccurredAt  time.Time       `json:"occurred_at"`
	SessionHash string          `json:"session_hash"`
	Platform    string          `json:"platform"`
	CountryCode string          `json:"country_code"`
	Locale      string          `json:"locale"`
	Device      DeviceInfo      `json:"device"`
	Properties  EventProperties `json:"properties"`
}

type EventStore interface {
	SaveBatch(ctx context.Context, events []AnalyticsEventEnvelope) error
	GetEventsByApp(ctx context.Context, appID string, from, to time.Time) ([]AnalyticsEventEnvelope, error)
	GetAllEvents(ctx context.Context) ([]AnalyticsEventEnvelope, error)
}

type MemoryEventStore struct {
	mu     sync.RWMutex
	events []AnalyticsEventEnvelope
}

func NewMemoryEventStore() *MemoryEventStore {
	return &MemoryEventStore{
		events: make([]AnalyticsEventEnvelope, 0),
	}
}

func (s *MemoryEventStore) SaveBatch(ctx context.Context, events []AnalyticsEventEnvelope) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, events...)
	return nil
}

func (s *MemoryEventStore) GetEventsByApp(ctx context.Context, appID string, from, to time.Time) ([]AnalyticsEventEnvelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var res []AnalyticsEventEnvelope
	for _, e := range s.events {
		if e.Properties.AppID == appID {
			if !from.IsZero() && e.OccurredAt.Before(from) {
				continue
			}
			if !to.IsZero() && e.OccurredAt.After(to) {
				continue
			}
			res = append(res, e)
		}
	}
	return res, nil
}

func (s *MemoryEventStore) GetAllEvents(ctx context.Context) ([]AnalyticsEventEnvelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]AnalyticsEventEnvelope, len(s.events))
	copy(res, s.events)
	return res, nil
}

type FunnelIngestor struct {
	store EventStore
}

func NewFunnelIngestor(store EventStore) *FunnelIngestor {
	return &FunnelIngestor{store: store}
}

func (fi *FunnelIngestor) IngestBatch(ctx context.Context, batch []AnalyticsEventEnvelope) (int, error) {
	if len(batch) == 0 {
		return 0, ErrEmptyBatch
	}

	validEvents := make([]AnalyticsEventEnvelope, 0, len(batch))
	now := time.Now().UTC()

	for i, e := range batch {
		if e.EventName == "" {
			return 0, fmt.Errorf("batch[%d]: %w", i, ErrInvalidEventName)
		}
		if e.Properties.AppID == "" {
			return 0, fmt.Errorf("batch[%d]: %w", i, ErrMissingAppID)
		}
		if e.OccurredAt.IsZero() {
			e.OccurredAt = now
		}
		if e.CountryCode == "" {
			e.CountryCode = "UNKNOWN"
		}
		if e.Platform == "" {
			e.Platform = "ANDROID_CLIENT"
		}
		validEvents = append(validEvents, e)
	}

	if err := fi.store.SaveBatch(ctx, validEvents); err != nil {
		return 0, err
	}

	return len(validEvents), nil
}

func (fi *FunnelIngestor) IngestInstall(ctx context.Context, event AnalyticsEventEnvelope) error {
	event.EventName = EventInstallSucceeded
	if event.Properties.AppID == "" {
		return ErrMissingAppID
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}
	return fi.store.SaveBatch(ctx, []AnalyticsEventEnvelope{event})
}

