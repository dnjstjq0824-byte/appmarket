package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	catalogDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
	catalogInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/infrastructure"
	commerceApp "github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	developerApp "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/application"
	developerInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/infrastructure"
	developerPres "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/presentation"
	distApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	adsApp "github.com/bookingnetwork/appmarket/backend/internal/modules/ads/application"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/console"
)

func consoleTestRouter(t *testing.T) (*developerPres.DeveloperHandler, string) {
	t.Helper()
	ctx := context.Background()
	devRepo := developerInfra.NewInMemoryDeveloperRepository()
	orgs := developerApp.NewOrgService(devRepo)
	members := developerApp.NewMemberService(devRepo)
	keys := developerApp.NewAPIKeyService(devRepo)
	catalogRepo := catalogInfra.NewInMemoryCatalogRepository()
	catalogSvc := application.NewCatalogService(catalogRepo, catalogInfra.NewInMemoryCache())
	artifacts := distApp.NewArtifactService(distApp.NewInMemoryDistributionRepo(), nil)
	ledger := commerceApp.NewLedgerService()
	orders := commerceApp.NewOrderService(ledger, "test-secret")
	promos := commerceApp.NewPromoService()
	pricing := commerceApp.NewPricingService()
	campaigns := adsApp.NewCampaignService(nil)

	resolver := console.NewOrgAppResolver(orgs, members, catalogSvc)
	handler := developerPres.NewDeveloperHandler(orgs, members, keys, func(*http.Request) (string, error) {
		return "user-1", nil
	})
	handler.AttachConsole(developerPres.ConsoleDeps{
		Resolver: resolver,
		Uploads:  console.NewUploadService(artifacts),
		Rollouts: console.NewRolloutService(artifacts),
		Catalog:  console.NewCatalogViewService(resolver, catalogSvc),
		Metadata: console.NewMetadataService(resolver, catalogSvc),
		Ads:      console.NewAdsViewService(resolver, campaigns),
		Commerce: console.NewCommerceViewService(resolver, promos, pricing, orders),
		Lineage:  console.NewLineageService(resolver, artifacts),
	})

	org, err := orgs.CreateOrg(ctx, "acme", "Acme")
	if err != nil {
		t.Fatalf("org create failed: %v", err)
	}
	if _, err := members.AddMember(ctx, org.ID, "user-1", "DEV_OWNER"); err != nil {
		t.Fatalf("member add failed: %v", err)
	}
	orgID := org.ID
	catalogRepo.AddApp(
		&catalogDomain.CatalogApp{ID: "app-acme-01", CanonicalSlug: "acme-app", PrimaryPackageName: "com.acme.app", DeveloperOrgID: &orgID, CatalogStatus: catalogDomain.CatalogStatusPublished},
		[]catalogDomain.Localization{{ID: "loc-1", AppID: "app-acme-01", Locale: "en-US", Title: "Acme"}},
		nil,
		[]catalogDomain.Distribution{{ID: "d1", AppID: "app-acme-01"}},
		nil, nil,
	)
	return handler, org.ID
}

func doConsole(t *testing.T, router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestConsoleOrgsMembersKeysVerification(t *testing.T) {
	handler, orgID := consoleTestRouter(t)
	router := handler.Routes()

	w := doConsole(t, router, http.MethodGet, "/orgs", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list orgs: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodGet, "/orgs/"+orgID+"/members", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "user-1") {
		t.Fatalf("list members: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodPost, "/orgs/"+orgID+"/verification", map[string]string{"doc_type": "BRN", "doc_url": "https://x.test/doc", "domain": "acme.test"})
	if w.Code != http.StatusCreated {
		t.Fatalf("submit verification: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodGet, "/orgs/"+orgID+"/verification", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get verification: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodPost, "/orgs/"+orgID+"/api-keys", map[string]any{"name": "ci", "scopes": []string{"release:upload"}})
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("issue key: %d %s", w.Code, w.Body.String())
	}
	var issued map[string]any
	if err := json.NewDecoder(w.Body).Decode(&issued); err != nil {
		t.Fatalf("decode key: %v", err)
	}
	keyID, _ := issued["id"].(string)
	if keyID == "" {
		if k, ok := issued["key"].(map[string]any); ok {
			keyID, _ = k["id"].(string)
		}
	}
	w = doConsole(t, router, http.MethodGet, "/orgs/"+orgID+"/api-keys", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "ci") {
		t.Fatalf("list keys: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "key_secret") || strings.Contains(w.Body.String(), "KeyHash") {
		t.Fatalf("key secret leaked in list: %s", w.Body.String())
	}
	if keyID != "" {
		w = doConsole(t, router, http.MethodDelete, "/orgs/"+orgID+"/api-keys/"+keyID, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("revoke key: %d %s", w.Code, w.Body.String())
		}
	}
}

func TestConsoleUploadsAndRollout(t *testing.T) {
	handler, _ := consoleTestRouter(t)
	router := handler.Routes()
	sha := strings.Repeat("b", 64)

	w := doConsole(t, router, http.MethodPost, "/apps/current/uploads", map[string]any{"fileName": "app.apk", "fileSize": 2048, "expectedSha256": sha})
	if w.Code != http.StatusCreated {
		t.Fatalf("init upload: %d %s", w.Code, w.Body.String())
	}
	var session map[string]any
	if err := json.NewDecoder(w.Body).Decode(&session); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	uploadID, _ := session["uploadId"].(string)
	if uploadID == "" {
		t.Fatalf("missing uploadId: %v", session)
	}
	w = doConsole(t, router, http.MethodPost, "/apps/current/uploads/"+uploadID+"/complete", map[string]string{"sha256": sha})
	if w.Code != http.StatusOK {
		t.Fatalf("complete upload: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodGet, "/apps/current/releases", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list releases: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodPatch, "/releases/current/rollout", map[string]any{"rolloutPercentage": 25})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 no-release, got %d %s", w.Code, w.Body.String())
	}
}

func TestConsoleMetadataAdsPromosPricing(t *testing.T) {
	handler, _ := consoleTestRouter(t)
	router := handler.Routes()

	w := doConsole(t, router, http.MethodGet, "/apps/current/metadata", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Acme") {
		t.Fatalf("list metadata: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodPut, "/apps/current/metadata", map[string]string{
		"locale": "ko-KR", "title": "에크미", "shortDescription": "짧은", "fullDescription": "전체 설명", "whatsNew": "신규",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("save metadata: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodPost, "/apps/current/ads/campaigns", map[string]any{
		"appName": "Acme", "targetKeywords": []string{"tools"}, "maxBidCpc": 200, "dailyBudget": 10000,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create campaign: %d %s", w.Code, w.Body.String())
	}
	var camp map[string]any
	_ = json.NewDecoder(w.Body).Decode(&camp)
	w = doConsole(t, router, http.MethodGet, "/apps/current/ads/campaigns", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list campaigns: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodPost, "/apps/current/promotions", map[string]any{
		"campaignName": "Launch", "discountType": "PERCENTAGE", "discountValue": 15,
		"startsAt": "2025-01-01T00:00:00Z", "endsAt": "2027-01-01T00:00:00Z",
		"maxUses": 50, "codeCount": 2, "codePrefix": "ACME",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create promotion: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodGet, "/apps/current/promotions", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Launch") {
		t.Fatalf("list promotions: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodPost, "/apps/current/pricing", map[string]any{
		"countryCode": "KR", "currency": "KRW", "listPriceMinor": 15000, "taxInclusive": true,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("set pricing: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodGet, "/apps/current/pricing", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "15000") {
		t.Fatalf("list pricing: %d %s", w.Code, w.Body.String())
	}
	w = doConsole(t, router, http.MethodGet, "/apps/current/lineage", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("lineage: %d %s", w.Code, w.Body.String())
	}
}

func TestConsoleOrgSettlement(t *testing.T) {
	handler, orgID := consoleTestRouter(t)
	router := handler.Routes()
	w := doConsole(t, router, http.MethodGet, "/orgs/"+orgID+"/settlements", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "payoutTotal") {
		t.Fatalf("settlement: %d %s", w.Code, w.Body.String())
	}
}
