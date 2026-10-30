package treescan_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/treescan"
)

func TestTreeScan_TC_API_003_FederatedSearchAndFeedContract(t *testing.T) {
	ctx := context.Background()
	serviceToken := "ts_sec_prod_treescan_federated_2026"
	adapter := treescan.NewAdapterHandler(serviceToken)

	// Step 1: When disabled (default independence status)
	res, err := adapter.FederatedSearch(ctx, "지도", 10, serviceToken)
	if err != treescan.ErrFeatureDisabled || res != nil {
		t.Fatalf("expected ErrFeatureDisabled when flag is off, got: %v", err)
	}

	// Step 2: Enable integration adapter (Phase 10)
	adapter.SetEnabled(true)

	// Step 3: Test Unauthorized Service Token
	_, err = adapter.FederatedSearch(ctx, "지도", 10, "wrong-token")
	if err != treescan.ErrUnauthorizedService {
		t.Fatalf("expected ErrUnauthorizedService, got: %v", err)
	}

	// Step 4: Test Federated Search Performance & Schema
	searchRes, err := adapter.FederatedSearch(ctx, "오프라인 지도", 10, serviceToken)
	if err != nil {
		t.Fatalf("federated search failed: %v", err)
	}

	// Latency requirement: took_ms must be < 80ms
	if searchRes.TookMs >= 80 {
		t.Errorf("expected took_ms < 80ms, got: %d ms", searchRes.TookMs)
	}

	if len(searchRes.Items) == 0 {
		t.Fatalf("expected search results, got 0 items")
	}

	item := searchRes.Items[0]
	if item.PrimaryPackageName != "net.osmand.plus" {
		t.Errorf("expected primary_package_name net.osmand.plus, got: %s", item.PrimaryPackageName)
	}
	if item.CanonicalSlug != "osmand-maps-navigation" {
		t.Errorf("expected canonical slug osmand-maps-navigation, got: %s", item.CanonicalSlug)
	}
	if !item.CanDirectInstall {
		t.Errorf("expected can_direct_install == true")
	}

	// Step 5: Test Incremental Catalog Feed
	feedRes, err := adapter.CatalogFeed(ctx, "", 2, serviceToken)
	if err != nil {
		t.Fatalf("catalog feed failed: %v", err)
	}
	if len(feedRes.Items) != 2 {
		t.Fatalf("expected 2 items with limit=2, got: %d", len(feedRes.Items))
	}
	if !feedRes.HasMore || feedRes.NextCursor == "" {
		t.Errorf("expected has_more == true and valid next_cursor")
	}

	// Verify action types (UPSERT vs REMOVE)
	feedPage2, err := adapter.CatalogFeed(ctx, feedRes.NextCursor, 5, serviceToken)
	if err != nil {
		t.Fatalf("catalog feed page 2 failed: %v", err)
	}
	hasTakedown := false
	for _, it := range feedPage2.Items {
		if it.Action == treescan.ActionRemoveFromIndex {
			hasTakedown = true
			if it.CatalogStatus != "SUSPENDED_BY_ADMIN" {
				t.Errorf("expected SUSPENDED_BY_ADMIN status on takedown item")
			}
		}
	}
	if !hasTakedown {
		t.Errorf("expected takedown item in feed page 2")
	}
}

func TestTreeScan_TC_OPS_008_RFC8693SSOTokenExchange(t *testing.T) {
	ctx := context.Background()
	serviceToken := "ts_sec_prod_treescan_federated_2026"
	adapter := treescan.NewAdapterHandler(serviceToken)
	ssoService := treescan.NewSSOTokenExchangeService(adapter, "am_secret_test")

	validExp := time.Now().Add(10 * time.Minute).Unix()
	expiredExp := time.Now().Add(-10 * time.Minute).Unix()

	// 1. When Feature Flag is OFF -> Fails cleanly with ErrFeatureDisabled
	req := treescan.TokenExchangeRequest{
		GrantType:        treescan.RFC8693GrantType,
		SubjectTokenType: treescan.RFC8693IdTokenType,
		SubjectToken:     fmt.Sprintf("https://auth.treescan.internal|ts_sub_101|user@treescan.io|appmarket-api|%d", validExp),
	}
	_, err := ssoService.ExchangeToken(ctx, req)
	if err != treescan.ErrFeatureDisabled {
		t.Fatalf("expected ErrFeatureDisabled when disabled, got: %v", err)
	}

	// 2. Enable Feature Flag
	adapter.SetEnabled(true)

	// 3. Test Invalid Grant Type
	badGrantReq := req
	badGrantReq.GrantType = "authorization_code"
	_, err = ssoService.ExchangeToken(ctx, badGrantReq)
	if err != treescan.ErrInvalidGrantType {
		t.Fatalf("expected ErrInvalidGrantType, got: %v", err)
	}

	// 4. Test Expired Subject Token
	expiredReq := req
	expiredReq.SubjectToken = fmt.Sprintf("https://auth.treescan.internal|ts_sub_101|user@treescan.io|appmarket-api|%d", expiredExp)
	_, err = ssoService.ExchangeToken(ctx, expiredReq)
	if err != treescan.ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got: %v", err)
	}

	// 5. Test Successful Token Exchange (JIT User Provisioning)
	exchangeRes, err := ssoService.ExchangeToken(ctx, req)
	if err != nil {
		t.Fatalf("token exchange failed: %v", err)
	}

	if exchangeRes.TokenType != "Bearer" {
		t.Errorf("expected Bearer token type, got: %s", exchangeRes.TokenType)
	}
	if exchangeRes.UserID != "usr-jit-ts-ts_sub_101" {
		t.Errorf("expected JIT user ID usr-jit-ts-ts_sub_101, got: %s", exchangeRes.UserID)
	}
	if exchangeRes.Email != "user@treescan.io" {
		t.Errorf("expected user@treescan.io, got: %s", exchangeRes.Email)
	}
	if exchangeRes.AccessToken == "" || exchangeRes.RefreshToken == "" {
		t.Errorf("expected non-empty access and refresh tokens")
	}

	// 6. Test Idempotent JIT mapping for same user
	exchangeRes2, err := ssoService.ExchangeToken(ctx, req)
	if err != nil {
		t.Fatalf("subsequent token exchange failed: %v", err)
	}
	if exchangeRes2.UserID != exchangeRes.UserID {
		t.Errorf("subsequent token exchange must map to identical local user ID")
	}
}

