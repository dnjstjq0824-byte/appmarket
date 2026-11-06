package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/search/domain"
)

func TestOpenSearchEngineSearchSuggestIndex(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"result":"created"}`))
			return
		}
		_, _ = w.Write([]byte(`{"took":3,"hits":{"total":{"value":1},"hits":[{"_score":9.5,"_source":{"app_id":"app-1","package_name":"com.demo.app","title_i18n":{"en-US":"Demo App","ko-KR":"데모 앱"},"description_i18n":{},"developer_name":"Demo Studio","category":"TOOLS","bayesian_rating":4.5,"rating_count":10,"download_count":100,"direct_download_available":true,"rights_tier":"T1","keywords":["demo"]}}]}}`))
	}))
	defer server.Close()

	engine := NewOpenSearchEngine(server.URL, "", "", "apps", server.Client())
	res, err := engine.Search(ctx, domain.SearchQuery{Keyword: "demo", Limit: 10})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if res.TotalCount != 1 || len(res.Items) != 1 || res.Items[0].AppID != "app-1" || res.Items[0].Score != 9.5 {
		t.Fatalf("unexpected result: %+v", res)
	}
	sug, err := engine.Suggest(ctx, "de", "ko-KR", 5)
	if err != nil {
		t.Fatalf("suggest failed: %v", err)
	}
	if len(sug.Suggestions) == 0 || sug.Suggestions[0].Text != "데모 앱" {
		t.Fatalf("unexpected suggestions: %+v", sug)
	}
	if err := engine.IndexDocument(ctx, domain.SearchDoc{AppID: "app-1"}); err != nil {
		t.Fatalf("index failed: %v", err)
	}
	if err := engine.IndexDocument(ctx, domain.SearchDoc{}); err == nil {
		t.Fatalf("expected app_id validation error")
	}
}

func TestSearchServiceFallsBackOnPrimaryFailure(t *testing.T) {
	ctx := context.Background()
	primary := NewOpenSearchEngine("https://127.0.0.1:1", "", "", "apps", nil)
	fallback := NewInMemoryFallbackSearchEngine()
	svc := NewSearchService(primary, fallback)
	res, err := svc.Search(ctx, domain.SearchQuery{Keyword: "anything"})
	if err != nil {
		t.Fatalf("expected fallback result, got %v", err)
	}
	if !res.FallbackUsed {
		t.Fatalf("expected FallbackUsed=true when primary is down")
	}
}
