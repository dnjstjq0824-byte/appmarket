package console

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConnectorSyncWithMockFeed(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"repo":{"fingerprint":"AA:BB","timestamp":0,"url":"https://feed.test/repo"},"packages":{"org.example.one":{"packageName":"org.example.one"},"org.example.two":{"packageName":"org.example.two"}}}`))
	}))
	defer server.Close()

	svc := NewConnectorService()
	svc.RegisterConnector(ctx, Connector{ID: "mock-fdroid", Name: "Mock", Kind: "FDROID", FeedURL: server.URL, Fingerprint: "AA:BB", Syncable: true})
	res, err := svc.TriggerSync(ctx, "mock-fdroid")
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if res.ItemCount != 2 {
		t.Fatalf("expected 2 items, got %d", res.ItemCount)
	}
	svc.RegisterConnector(ctx, Connector{ID: "mock-partner", Name: "Partner", Kind: "PARTNER", FeedURL: "https://apps.samsung.com/app", Packages: []string{"com.a", "com.b", ""}, Syncable: true})
	pres, err := svc.TriggerSync(ctx, "mock-partner")
	if err != nil {
		t.Fatalf("partner sync failed: %v", err)
	}
	if pres.ItemCount != 2 {
		t.Fatalf("expected 2 partner items, got %d", pres.ItemCount)
	}
	if _, err := svc.TriggerSync(ctx, "galaxy-store"); err != ErrConnectorNotSyncable {
		t.Fatalf("expected not-syncable, got %v", err)
	}
	if _, err := svc.TriggerSync(ctx, "missing"); err != ErrConnectorNotFound {
		t.Fatalf("expected not-found, got %v", err)
	}
	if list := svc.ListConnectors(ctx); len(list) != 4 {
		t.Fatalf("expected 4 connectors, got %d", len(list))
	}
}
