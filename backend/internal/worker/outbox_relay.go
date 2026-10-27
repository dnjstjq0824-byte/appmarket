package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type OutboxStatus string

const (
	OutboxPending   OutboxStatus = "PENDING"
	OutboxProcessed OutboxStatus = "PROCESSED"
	OutboxFailed    OutboxStatus = "FAILED"
)

type OutboxEvent struct {
	ID            string       `json:"id"`
	AggregateType string       `json:"aggregate_type"`
	AggregateID   string       `json:"aggregate_id"`
	EventType     string       `json:"event_type"`
	PayloadJSON   string       `json:"payload_json"`
	Status        OutboxStatus `json:"status"`
	RetryCount    int          `json:"retry_count"`
	CreatedAt     time.Time    `json:"created_at"`
	ProcessedAt   *time.Time   `json:"processed_at,omitempty"`
}

// 6 Core Redis Streams Queues defined in the master blueprint
const (
	QueueAPKAnalysis   = "queue:apk_analysis"
	QueueSearchIndexer = "queue:search_indexer"
	QueueAdBilling     = "queue:ad_billing"
	QueueNotifications = "queue:notifications"
	QueueAnalytics     = "queue:analytics"
	QueueDeadLetter    = "queue:dead_letter"
)

// MapEventToQueue returns the destination queue for a given event type.
func MapEventToQueue(eventType string) string {
	switch eventType {
	case "APK_UPLOADED", "RELEASE_SUBMITTED", "VIRUS_SCAN_REQUESTED":
		return QueueAPKAnalysis
	case "APP_CATALOG_UPDATED", "RELEASE_APPROVED", "EXTERNAL_APP_SYNCED":
		return QueueSearchIndexer
	case "ORDER_COMPLETED", "REFUND_PROCESSED", "AD_CLICK_RECORDED", "SETTLEMENT_CALCULATED":
		return QueueAdBilling
	case "NOTIFICATION_DISPATCH", "USER_REGISTERED", "EMAIL_VERIFY_REQUESTED":
		return QueueNotifications
	case "INSTALL_RECORDED", "SEARCH_PERFORMED", "IMPRESSION_RECORDED":
		return QueueAnalytics
	default:
		return QueueNotifications
	}
}

type StreamPublisher interface {
	PublishStream(ctx context.Context, queue string, event *OutboxEvent) error
}

type InMemoryStreamPublisher struct {
	mu       sync.RWMutex
	Messages map[string][]*OutboxEvent
}

func NewInMemoryStreamPublisher() *InMemoryStreamPublisher {
	return &InMemoryStreamPublisher{
		Messages: make(map[string][]*OutboxEvent),
	}
}

func (p *InMemoryStreamPublisher) PublishStream(ctx context.Context, queue string, event *OutboxEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Messages[queue] = append(p.Messages[queue], event)
	return nil
}

type OutboxRepository interface {
	GetPendingEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)
	MarkProcessed(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, errStr string) error
}

type InMemoryOutboxRepository struct {
	mu     sync.RWMutex
	events map[string]*OutboxEvent
}

func NewInMemoryOutboxRepository() *InMemoryOutboxRepository {
	return &InMemoryOutboxRepository{
		events: make(map[string]*OutboxEvent),
	}
}

func (r *InMemoryOutboxRepository) AddEvent(event *OutboxEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[event.ID] = event
}

func (r *InMemoryOutboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]*OutboxEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var pending []*OutboxEvent
	for _, e := range r.events {
		if e.Status == OutboxPending {
			pending = append(pending, e)
			if len(pending) >= limit {
				break
			}
		}
	}
	return pending, nil
}

func (r *InMemoryOutboxRepository) MarkProcessed(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.events[id]
	if !ok {
		return fmt.Errorf("outbox event not found")
	}
	now := time.Now().UTC()
	e.Status = OutboxProcessed
	e.ProcessedAt = &now
	return nil
}

func (r *InMemoryOutboxRepository) MarkFailed(ctx context.Context, id string, errStr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.events[id]
	if !ok {
		return fmt.Errorf("outbox event not found")
	}
	e.RetryCount++
	if e.RetryCount >= 5 {
		e.Status = OutboxFailed
	}
	return nil
}

type OutboxRelayWorker struct {
	repo      OutboxRepository
	publisher StreamPublisher
}

func NewOutboxRelayWorker(repo OutboxRepository, publisher StreamPublisher) *OutboxRelayWorker {
	return &OutboxRelayWorker{
		repo:      repo,
		publisher: publisher,
	}
}

// ProcessBatch polls pending outbox events and reliably publishes them to the corresponding Redis Stream queue.
func (w *OutboxRelayWorker) ProcessBatch(ctx context.Context, batchSize int) (int, error) {
	events, err := w.repo.GetPendingEvents(ctx, batchSize)
	if err != nil {
		return 0, err
	}

	processedCount := 0
	for _, ev := range events {
		targetQueue := MapEventToQueue(ev.EventType)
		if err := w.publisher.PublishStream(ctx, targetQueue, ev); err != nil {
			_ = w.repo.MarkFailed(ctx, ev.ID, err.Error())
			continue
		}
		if err := w.repo.MarkProcessed(ctx, ev.ID); err != nil {
			return processedCount, err
		}
		processedCount++
	}

	return processedCount, nil
}

// CreateOutboxEvent creates a new pending outbox event.
func CreateOutboxEvent(aggregateType, aggregateID, eventType, payloadJSON string) *OutboxEvent {
	return &OutboxEvent{
		ID:            uuid.New().String(),
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		PayloadJSON:   payloadJSON,
		Status:        OutboxPending,
		RetryCount:    0,
		CreatedAt:     time.Now().UTC(),
	}
}
