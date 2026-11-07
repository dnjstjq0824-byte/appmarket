package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	adsApp "github.com/bookingnetwork/appmarket/backend/internal/modules/ads/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/analytics"
	commerceApp "github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	commerceDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/settlement"
	commerceSet "github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/settlement"
	iamApp "github.com/bookingnetwork/appmarket/backend/internal/modules/iam/application"
	iamRbac "github.com/bookingnetwork/appmarket/backend/internal/modules/iam/rbac"
	moderationApp "github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/application"
	securityApp "github.com/bookingnetwork/appmarket/backend/internal/modules/security/application"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/audit"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/console"
)

func adminTestRouter(t *testing.T) (*AdminHandler, *settlement.MonthlyPayoutService, *moderationApp.ReportService, *moderationApp.AppealService, *adsApp.CampaignService, analytics.EventStore) {
	t.Helper()
	ledger := commerceApp.NewLedgerService()
	orders := commerceApp.NewOrderService(ledger, "test-secret")
	payouts := settlement.NewMonthlyPayoutService(ledger)
	auditRepo := audit.NewInMemoryAuditRepository()
	reports := moderationApp.NewReportService()
	appeals := moderationApp.NewAppealService()
	campaigns := adsApp.NewCampaignService(nil)
	store := analytics.NewMemoryEventStore()
	identity := func(*http.Request) (string, error) { return "admin-1", nil }
	handler := NewAdminHandler(
		console.NewDecisionService(audit.NewHashChainLogger(auditRepo)),
		securityApp.NewSecurityService(nil),
		console.NewConnectorService(),
		payouts, auditRepo, orders, campaigns, store, reports, appeals, identity,
	)
	return handler, payouts, reports, appeals, campaigns, store
}

func doAdmin(t *testing.T, router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, _ := json.Marshal(body)
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

func TestAdminAuthorizerEnforced(t *testing.T) {
	handler, _, _, _, _, _ := adminTestRouter(t)
	jwtMgr, err := iamApp.NewJWTManager(nil, "https://auth.appmarket.local")
	if err != nil {
		t.Fatalf("jwt manager: %v", err)
	}
	handler.AttachAuthorizer(iamRbac.NewMatrixEnforcer(jwtMgr))
	router := handler.Routes()

	superToken, _ := jwtMgr.GenerateAccessToken("super-1", "super@test.local", []string{"ADMIN_SUPER"}, nil, true)
	reviewerToken, _ := jwtMgr.GenerateAccessToken("rev-1", "rev@test.local", []string{"ADMIN_APP_REVIEWER"}, nil, true)
	userToken, _ := jwtMgr.GenerateAccessToken("user-1", "user@test.local", []string{"USER"}, nil, false)

	post := func(token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/reviews/app-9/decision", strings.NewReader(`{"decision":"APPROVED","notes":"ok"}`))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	if w := post(""); w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", w.Code)
	}
	if w := post(userToken); w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for USER role, got %d", w.Code)
	}
	if w := post(reviewerToken); w.Code != http.StatusOK {
		t.Fatalf("expected 200 for reviewer, got %d %s", w.Code, w.Body.String())
	}
	if w := post(superToken); w.Code != http.StatusOK {
		t.Fatalf("expected 200 for super admin, got %d %s", w.Code, w.Body.String())
	}
}

func TestAdminReviewDecision(t *testing.T) {
	handler, _, _, _, _, _ := adminTestRouter(t)
	router := handler.Routes()
	w := doAdmin(t, router, http.MethodPost, "/reviews/app-1/decision", map[string]any{
		"decision": "APPROVED", "notes": "all checks pass", "checklist": map[string]bool{"icons": true},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("decision: %d %s", w.Code, w.Body.String())
	}
	w = doAdmin(t, router, http.MethodPost, "/reviews/app-2/decision", map[string]any{"decision": "REJECTED", "notes": ""})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected notes requirement, got %d", w.Code)
	}
	w = doAdmin(t, router, http.MethodGet, "/reviews/decisions", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "app-1") {
		t.Fatalf("list decisions: %d %s", w.Code, w.Body.String())
	}
}

func TestAdminKillSwitchLifecycle(t *testing.T) {
	handler, _, _, _, _, _ := adminTestRouter(t)
	router := handler.Routes()
	w := doAdmin(t, router, http.MethodPost, "/security/kill-switch/com.evil.app", map[string]string{
		"scope": "GLOBAL", "reasonCode": "MALWARE_CONFIRMED", "evidence": "av verdict",
	})
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "true") {
		t.Fatalf("activate: %d %s", w.Code, w.Body.String())
	}
	w = doAdmin(t, router, http.MethodGet, "/security/kill-switch/com.evil.app", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d %s", w.Code, w.Body.String())
	}
	w = doAdmin(t, router, http.MethodDelete, "/security/kill-switch/com.evil.app", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "false") {
		t.Fatalf("lift: %d %s", w.Code, w.Body.String())
	}
	w = doAdmin(t, router, http.MethodGet, "/security/scans", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list scans: %d %s", w.Code, w.Body.String())
	}
}

func TestAdminSettlementDualControl(t *testing.T) {
	handler, payouts, _, _, _, _ := adminTestRouter(t)
	router := handler.Routes()
	w := doAdmin(t, router, http.MethodPost, "/settlements", map[string]any{"orgId": "org-1", "amount": 8500, "currency": "KRW"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.NewDecoder(w.Body).Decode(&created)
	payoutID, _ := created["id"].(string)

	w = doAdmin(t, router, http.MethodPost, "/settlements/"+payoutID+"/approve", map[string]string{"approver_id": "admin-1", "initiator_id": "admin-1"})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 self-approval, got %d %s", w.Code, w.Body.String())
	}
	w = doAdmin(t, router, http.MethodPost, "/settlements/"+payoutID+"/approve", map[string]string{"approver_id": "admin-2", "initiator_id": "someone-else"})
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 initiator mismatch, got %d %s", w.Code, w.Body.String())
	}
	_ = payouts
	w = doAdmin(t, router, http.MethodPost, "/settlements/"+payoutID+"/approve", map[string]string{"approver_id": "admin-2", "initiator_id": "admin-1"})
	if w.Code != http.StatusOK {
		t.Fatalf("approve: %d %s", w.Code, w.Body.String())
	}
	w = doAdmin(t, router, http.MethodGet, "/settlements", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), payoutID) {
		t.Fatalf("list: %d %s", w.Code, w.Body.String())
	}
}

func TestAdminAuditOrdersAdsAnalytics(t *testing.T) {
	handler, _, _, _, campaigns, store := adminTestRouter(t)
	router := handler.Routes()
	ctx := context.Background()

	w := doAdmin(t, router, http.MethodGet, "/audit-logs", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "chainValid") {
		t.Fatalf("audit: %d %s", w.Code, w.Body.String())
	}
	w = doAdmin(t, router, http.MethodGet, "/orders", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("orders: %d %s", w.Code, w.Body.String())
	}
	c, err := campaigns.CreateCampaign(ctx, adsApp.CreateCampaignInput{OrgID: "org-1", AppID: "app-1", TargetKeywords: []string{"k"}, MaxBidCPC: 100, DailyBudget: 1000})
	if err != nil {
		t.Fatalf("campaign seed failed: %v", err)
	}
	w = doAdmin(t, router, http.MethodGet, "/ads/campaigns?status=PENDING_REVIEW", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), c.ID) {
		t.Fatalf("ads queue: %d %s", w.Code, w.Body.String())
	}
	w = doAdmin(t, router, http.MethodPost, "/ads/campaigns/"+c.ID+"/review", map[string]bool{"approve": true})
	if w.Code != http.StatusOK {
		t.Fatalf("ads review: %d %s", w.Code, w.Body.String())
	}
	if err := store.SaveBatch(ctx, []analytics.AnalyticsEventEnvelope{{
		EventID: "e1", EventName: "APP_VIEW", OccurredAt: time.Now().UTC(),
		SessionHash: "s", Platform: "ANDROID", CountryCode: "KR",
		Properties: analytics.EventProperties{AppID: "app-1"},
	}}); err != nil {
		t.Fatalf("seed events failed: %v", err)
	}
	w = doAdmin(t, router, http.MethodGet, "/analytics/overview", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "APP_VIEW") {
		t.Fatalf("overview: %d %s", w.Code, w.Body.String())
	}
}

func TestAdminAnalyticsOverviewRevenue(t *testing.T) {
	handler, _, _, _, _, store := adminTestRouter(t)
	router := handler.Routes()
	ctx := context.Background()

	seed := []analytics.AnalyticsEventEnvelope{
		{EventID: "p1", EventName: analytics.EventPurchaseCompleted, OccurredAt: time.Now().UTC(), SessionHash: "s1", Platform: "ANDROID_CLIENT", CountryCode: "KR", Properties: analytics.EventProperties{AppID: "app-1", PriceCents: 15000}},
		{EventID: "p2", EventName: analytics.EventPurchaseCompleted, OccurredAt: time.Now().UTC(), SessionHash: "s2", Platform: "WEB_USER_DESKTOP", CountryCode: "US", Properties: analytics.EventProperties{AppID: "app-1", PriceCents: 5000}},
		{EventID: "v1", EventName: analytics.EventDetailView, OccurredAt: time.Now().UTC(), SessionHash: "s1", Platform: "ANDROID_CLIENT", CountryCode: "KR", Properties: analytics.EventProperties{AppID: "app-1"}},
	}
	if err := store.SaveBatch(ctx, seed); err != nil {
		t.Fatalf("seed events failed: %v", err)
	}
	w := doAdmin(t, router, http.MethodGet, "/analytics/overview", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("overview: %d %s", w.Code, w.Body.String())
	}
	var body struct {
		TotalEvents       int   `json:"totalEvents"`
		TotalRevenueCents int64 `json:"totalRevenueCents"`
		RevenueByCountry  []struct {
			Key     string `json:"key"`
			Count   int    `json:"count"`
			Revenue int64  `json:"revenueCents"`
		} `json:"revenueByCountry"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if body.TotalEvents != 3 {
		t.Fatalf("expected 3 events, got %d", body.TotalEvents)
	}
	if body.TotalRevenueCents != 20000 {
		t.Fatalf("expected 20000 revenue cents, got %d", body.TotalRevenueCents)
	}
	if len(body.RevenueByCountry) != 2 || body.RevenueByCountry[0].Key != "KR" || body.RevenueByCountry[0].Revenue != 15000 {
		t.Fatalf("unexpected revenueByCountry: %+v", body.RevenueByCountry)
	}
}

func TestAdminForcedRefund(t *testing.T) {
	handler, _, _, _, _, _ := adminTestRouter(t)
	ledger := commerceApp.NewLedgerService()
	orderSvc := commerceApp.NewOrderService(ledger, "test-secret")
	refundSvc := commerceSet.NewRefundService(ledger, orderSvc)
	handler.AttachRefunds(orderSvc, refundSvc)
	router := handler.Routes()

	order, err := orderSvc.CreateOrder(context.Background(), "user-9", "app-9", "org-9", 11000, "KRW", "idem-9")
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	w := doAdmin(t, router, http.MethodPost, "/orders/"+order.ID+"/refund", map[string]any{"reason": "MALWARE_SECURITY_TAKEDOWN"})
	if w.Code != http.StatusOK {
		t.Fatalf("refund: %d %s", w.Code, w.Body.String())
	}
	got, _ := orderSvc.GetOrder(order.ID)
	if got.Status != commerceDomain.OrderRefunded {
		t.Fatalf("expected REFUNDED, got %s", got.Status)
	}
	w2 := doAdmin(t, router, http.MethodPost, "/orders/ghost/refund", map[string]any{"reason": "x"})
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for ghost order, got %d", w2.Code)
	}
}

func TestAdminReportsAndAppeals(t *testing.T) {
	handler, _, reports, appeals, _, _ := adminTestRouter(t)
	router := handler.Routes()
	ctx := context.Background()
	rep, err := reports.FileReport(ctx, moderationApp.FileReportInput{
		ReporterEmail: "u@x.test", TargetID: "com.x", Category: "SPAM_ABUSE",
		Description: "Repeated spam reviews posted by a bot network farm",
	})
	if err != nil {
		t.Fatalf("seed report failed: %v", err)
	}
	w := doAdmin(t, router, http.MethodGet, "/reports?status=OPEN", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), rep.ID) {
		t.Fatalf("reports queue: %d %s", w.Code, w.Body.String())
	}
	w = doAdmin(t, router, http.MethodPost, "/reports/"+rep.ID+"/claim", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("claim: %d %s", w.Code, w.Body.String())
	}
	w = doAdmin(t, router, http.MethodPost, "/reports/"+rep.ID+"/resolve", map[string]any{"action_taken": false, "notes": "not a violation"})
	if w.Code != http.StatusOK {
		t.Fatalf("resolve: %d %s", w.Code, w.Body.String())
	}
	appeal, err := appeals.SubmitAppeal(ctx, "org-1", "rel-1", "Permission is justified for NFC wallet", "rev-1")
	if err != nil {
		t.Fatalf("seed appeal failed: %v", err)
	}
	w = doAdmin(t, router, http.MethodGet, "/appeals", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), appeal.ID) {
		t.Fatalf("appeals: %d %s", w.Code, w.Body.String())
	}
	w = doAdmin(t, router, http.MethodPost, "/appeals/"+appeal.ID+"/decide", map[string]string{"decision": "APPROVED", "notes": "justified"})
	if w.Code != http.StatusOK {
		t.Fatalf("decide: %d %s", w.Code, w.Body.String())
	}
}

func TestAdminJWTRotationEndpoint(t *testing.T) {
	handler, _, _, _, _, _ := adminTestRouter(t)
	jwtMgr, err := iamApp.NewJWTManager(nil, "https://auth.appmarket.local")
	if err != nil {
		t.Fatalf("jwt manager: %v", err)
	}
	handler.AttachAuthorizer(iamRbac.NewMatrixEnforcer(jwtMgr))
	handler.AttachJWTManager(jwtMgr)
	router := handler.Routes()

	superToken, _ := jwtMgr.GenerateAccessToken("super-1", "super@test.local", []string{"ADMIN_SUPER"}, nil, true)
	firstKID := jwtMgr.KeyID()
	doRotate := func(token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/security/jwt/rotate", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	w := doRotate(superToken)
	if w.Code != http.StatusOK {
		t.Fatalf("rotate: %d %s", w.Code, w.Body.String())
	}
	var receipt map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &receipt); err != nil {
		t.Fatalf("receipt decode: %v", err)
	}
	if receipt["kid"] == firstKID || receipt["previous_kid"] != firstKID {
		t.Fatalf("unexpected rotation receipt: %v", receipt)
	}
	if len(jwtMgr.GetJWKS()["keys"].([]map[string]any)) != 2 {
		t.Fatalf("expected dual-key JWKS after rotation")
	}

	// Pre-rotation token keeps passing the full middleware chain (grace).
	req := httptest.NewRequest(http.MethodGet, "/security/kill-switches", nil)
	req.Header.Set("Authorization", "Bearer "+superToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("grace token rejected: %d %s", w.Code, w.Body.String())
	}

	// Permission without MFA must not rotate.
	secToken, _ := jwtMgr.GenerateAccessToken("sec-1", "sec@test.local", []string{"ADMIN_SECURITY"}, nil, false)
	if w := doRotate(secToken); w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without MFA, got %d", w.Code)
	}
	// Ordinary users must not rotate.
	userToken, _ := jwtMgr.GenerateAccessToken("user-1", "user@test.local", []string{"USER"}, nil, true)
	if w := doRotate(userToken); w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for USER, got %d", w.Code)
	}
}

func TestAdminJWTRotationUnavailableWithoutManager(t *testing.T) {
	handler, _, _, _, _, _ := adminTestRouter(t)
	jwtMgr, err := iamApp.NewJWTManager(nil, "https://auth.appmarket.local")
	if err != nil {
		t.Fatalf("jwt manager: %v", err)
	}
	handler.AttachAuthorizer(iamRbac.NewMatrixEnforcer(jwtMgr))
	router := handler.Routes()

	superToken, _ := jwtMgr.GenerateAccessToken("super-1", "super@test.local", []string{"ADMIN_SUPER"}, nil, true)
	req := httptest.NewRequest(http.MethodPost, "/security/jwt/rotate", nil)
	req.Header.Set("Authorization", "Bearer "+superToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 without attached manager, got %d", w.Code)
	}
}
