package console

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/connector/fdroid"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/connector/partner"
)

var (
	ErrConnectorNotFound = errors.New("ERR_CONNECTOR_NOT_FOUND")
	ErrConnectorNotSyncable = errors.New("ERR_CONNECTOR_NOT_SYNCABLE")
	ErrConnectorSyncFailed = errors.New("ERR_CONNECTOR_SYNC_FAILED")
)

// Connector is a registered external source.
type Connector struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	FeedURL     string   `json:"feedUrl,omitempty"`
	Fingerprint string   `json:"fingerprint,omitempty"`
	Packages    []string `json:"packages,omitempty"`
	Syncable    bool     `json:"syncable"`
	LastSyncAt  *time.Time `json:"lastSyncAt,omitempty"`
	LastItemCount int     `json:"lastItemCount,omitempty"`
}

// SyncResult reports one manual sync run.
type SyncResult struct {
	ConnectorID string    `json:"connectorId"`
	SyncedAt    time.Time `json:"syncedAt"`
	ItemCount   int       `json:"itemCount"`
}

// ConnectorService keeps the external-source registry and runs manual syncs.
type ConnectorService struct {
	mu         sync.RWMutex
	connectors map[string]*Connector
	http       *http.Client
}

func NewConnectorService() *ConnectorService {
	s := &ConnectorService{
		connectors: make(map[string]*Connector),
		http:       &http.Client{Timeout: 30 * time.Second},
	}
	s.connectors["fdroid-org"] = &Connector{
		ID: "fdroid-org", Name: "F-Droid Official", Kind: "FDROID",
		FeedURL: "https://f-droid.org/repo/index-v2.json", Syncable: true,
	}
	s.connectors["galaxy-store"] = &Connector{
		ID: "galaxy-store", Name: "Galaxy Store", Kind: "OFFICIAL_LINK", Syncable: false,
	}
	return s
}

func (s *ConnectorService) ListConnectors(_ context.Context) []Connector {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Connector, 0, len(s.connectors))
	for _, c := range s.connectors {
		out = append(out, *c)
	}
	return out
}

func (s *ConnectorService) RegisterConnector(_ context.Context, c Connector) Connector {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := c
	s.connectors[c.ID] = &cp
	return cp
}

// TriggerSync fetches the feed and runs the rights-guarded ingest.
func (s *ConnectorService) TriggerSync(ctx context.Context, id string) (*SyncResult, error) {
	s.mu.RLock()
	c, ok := s.connectors[id]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrConnectorNotFound
	}
	if !c.Syncable {
		return nil, ErrConnectorNotSyncable
	}
	count := 0
	switch strings.ToUpper(c.Kind) {
	case "FDROID":
		body, sig, err := s.fetchFeed(ctx, c.FeedURL)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrConnectorSyncFailed, err)
		}
		syncer := fdroid.NewFDroidSyncer(c.Fingerprint)
		items, err := syncer.SyncFeed(body, sig)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrConnectorSyncFailed, err)
		}
		count = len(items)
	case "PARTNER":
		syncer := partner.NewPartnerSyncer()
		for _, pkg := range c.Packages {
			if strings.TrimSpace(pkg) == "" {
				continue
			}
			if _, err := syncer.IngestPartnerApp(c.Name, c.FeedURL, pkg); err == nil {
				count++
			}
		}
	default:
		return nil, fmt.Errorf("%w: kind %s", ErrConnectorNotSyncable, c.Kind)
	}
	now := time.Now().UTC()
	s.mu.Lock()
	c.LastSyncAt = &now
	c.LastItemCount = count
	s.mu.Unlock()
	return &SyncResult{ConnectorID: id, SyncedAt: now, ItemCount: count}, nil
}

func (s *ConnectorService) fetchFeed(ctx context.Context, feedURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("feed status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, "", err
	}
	return body, resp.Header.Get("X-Feed-Signature"), nil
}
