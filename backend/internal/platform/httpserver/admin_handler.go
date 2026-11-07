package httpserver

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/ads/application"
	adsDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/ads/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/analytics"
	catalogApp "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	catalogDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
	commerceApp "github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	commerceDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/settlement"
	developerApp "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/application"
	developerDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
	iamApp "github.com/bookingnetwork/appmarket/backend/internal/modules/iam/application"
	iamRbac "github.com/bookingnetwork/appmarket/backend/internal/modules/iam/rbac"
	moderationApp "github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/application"
	moderationDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/domain"
	securityApp "github.com/bookingnetwork/appmarket/backend/internal/modules/security/application"
	securityDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/security/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/audit"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/console"
	problem "github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
)

// AdminHandler aggregates back-office operations. It lives in httpserver
// because it composes services from many modules, which modules themselves
// must never import (see architecture_test.go).
type AdminHandler struct {
	decisions  *console.DecisionService
	security   *securityApp.SecurityService
	connectors *console.ConnectorService
	payouts    *settlement.MonthlyPayoutService
	auditRepo  audit.AuditRepository
	orders     *commerceApp.OrderService
	refunds    *settlement.RefundService
	campaigns  *application.CampaignService
	eventStore analytics.EventStore
	reports    *moderationApp.ReportService
	appeals    *moderationApp.AppealService
	identity   func(*http.Request) (string, error)
	authorizer *iamRbac.MatrixEnforcer
	verifier   *developerApp.OrgService
	catalog    *catalogApp.CatalogService
	jwt        *iamApp.JWTManager
}

// AttachAuthorizer enables RBAC permission enforcement on admin routes.
// The composition root must always attach it; a nil authorizer leaves the
// legacy identity-only gate (used by unit tests without JWT).
func (h *AdminHandler) AttachAuthorizer(authorizer *iamRbac.MatrixEnforcer) {
	h.authorizer = authorizer
}

// AttachVerifications enables the KYB review queue backed by the org service.
func (h *AdminHandler) AttachVerifications(verifier *developerApp.OrgService) {
	h.verifier = verifier
}

// AttachCatalog enables the app review queue backed by the catalog service.
func (h *AdminHandler) AttachCatalog(catalog *catalogApp.CatalogService) {
	h.catalog = catalog
}

// AttachRefunds enables forced refunds backed by the commerce refund pipeline.
func (h *AdminHandler) AttachRefunds(orders *commerceApp.OrderService, refunds *settlement.RefundService) {
	h.orders = orders
	h.refunds = refunds
}

// AttachJWTManager enables signing-key rotation. Without it the rotate
// endpoint reports unavailable instead of failing on a nil manager.
func (h *AdminHandler) AttachJWTManager(jwt *iamApp.JWTManager) {
	h.jwt = jwt
}

func NewAdminHandler(
	decisions *console.DecisionService,
	security *securityApp.SecurityService,
	connectors *console.ConnectorService,
	payouts *settlement.MonthlyPayoutService,
	auditRepo audit.AuditRepository,
	orders *commerceApp.OrderService,
	campaigns *application.CampaignService,
	eventStore analytics.EventStore,
	reports *moderationApp.ReportService,
	appeals *moderationApp.AppealService,
	identity func(*http.Request) (string, error),
) *AdminHandler {
	return &AdminHandler{
		decisions: decisions, security: security, connectors: connectors,
		payouts: payouts, auditRepo: auditRepo, orders: orders,
		campaigns: campaigns, eventStore: eventStore, reports: reports,
		appeals: appeals, identity: identity,
	}
}

func (h *AdminHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireAuth)
	if h.authorizer != nil {
		r.Use(h.authorizer.AuthenticateMiddleware)
	}
	guard := func(permission string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			if h.authorizer == nil {
				return next
			}
			return h.authorizer.RequirePermissionMiddleware(permission)(next)
		}
	}
	mfa := func(next http.Handler) http.Handler {
		if h.authorizer == nil {
			return next
		}
		return h.authorizer.RequireMFAMiddleware(next)
	}
	r.With(guard("admin:review:decide")).Post("/reviews/{appId}/decision", h.handleReviewDecision)
	r.With(guard("admin:review:read")).Get("/reviews/decisions", h.handleListDecisions)
	r.With(guard("admin:killswitch:execute"), mfa).Post("/security/kill-switch/{packageName}", h.handleKillSwitchActivate)
	r.With(guard("admin:security:read")).Get("/security/kill-switch/{packageName}", h.handleKillSwitchStatus)
	r.With(guard("admin:security:read")).Get("/security/kill-switches", h.handleListKillSwitches)
	r.With(guard("admin:killswitch:execute"), mfa).Delete("/security/kill-switch/{packageName}", h.handleKillSwitchLift)
	r.With(guard("admin:security:decide"), mfa).Post("/security/jwt/rotate", h.handleJWTRotate)
	r.With(guard("admin:security:read")).Get("/security/scans", h.handleListScans)
	r.With(guard("admin:search:manage")).Get("/connectors", h.handleListConnectors)
	r.With(guard("admin:connector:manage")).Post("/connectors", h.handleRegisterConnector)
	r.With(guard("admin:connector:manage")).Post("/connectors/{id}/trigger-sync", h.handleTriggerSync)
	r.With(guard("admin:finance:read")).Get("/settlements", h.handleListSettlements)
	r.With(guard("admin:finance:read")).Post("/settlements", h.handleCreateSettlement)
	r.With(guard("admin:settlement:approve"), mfa).Post("/settlements/{id}/approve", h.handleApproveSettlement)
	r.With(guard("admin:audit:read")).Get("/audit-logs", h.handleAuditLogs)
	r.With(guard("admin:finance:read")).Get("/orders", h.handleListOrders)
	if h.refunds != nil {
		r.With(guard("admin:refund:decide")).Post("/orders/{id}/refund", h.handleForcedRefund)
	}
	r.With(guard("admin:review:read")).Get("/ads/campaigns", h.handleListAdCampaigns)
	r.With(guard("admin:review:decide")).Post("/ads/campaigns/{id}/review", h.handleReviewAdCampaign)
	r.With(guard("admin:audit:read")).Get("/analytics/overview", h.handleAnalyticsOverview)
	r.With(guard("admin:moderation:read")).Get("/reports", h.handleListReports)
	r.With(guard("admin:moderation:decide")).Post("/reports/{id}/claim", h.handleClaimReport)
	r.With(guard("admin:moderation:decide")).Post("/reports/{id}/resolve", h.handleResolveReport)
	r.With(guard("admin:moderation:read")).Get("/appeals", h.handleListAppeals)
	r.With(guard("admin:moderation:decide")).Post("/appeals/{id}/decide", h.handleDecideAppeal)
	if h.verifier != nil {
		r.With(guard("admin:kyb:read")).Get("/verifications", h.handleListVerifications)
		r.With(guard("admin:kyb:decide")).Post("/verifications/{orgId}/decision", h.handleDecideVerification)
	}
	if h.catalog != nil {
		r.With(guard("admin:review:read")).Get("/apps/review-queue", h.handleReviewQueue)
		r.With(guard("admin:review:read")).Get("/apps/{appId}", h.handleAdminGetApp)
	}
	return r
}

func (h *AdminHandler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.identity == nil {
			problem.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Authentication required", r.URL.Path)
			return
		}
		if userID, err := h.identity(r); err != nil || userID == "" {
			problem.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Authentication required", r.URL.Path)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *AdminHandler) callerID(r *http.Request) string {
	userID, _ := h.identity(r)
	return userID
}

func writeAdminJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type reviewDecisionReq struct {
	Decision  string          `json:"decision"`
	Notes     string          `json:"notes"`
	Checklist map[string]bool `json:"checklist"`
}

func (h *AdminHandler) handleReviewDecision(w http.ResponseWriter, r *http.Request) {
	var req reviewDecisionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	decision, err := h.decisions.Decide(r.Context(), chi.URLParam(r, "appId"), req.Decision, req.Notes, req.Checklist, h.callerID(r), r.RemoteAddr, r.UserAgent())
	if err != nil {
		if err == catalogDomain.ErrAppNotFound {
			problem.WriteProblem(w, http.StatusNotFound, "ERR_APP_NOT_FOUND", err.Error(), r.URL.Path)
			return
		}
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_DECISION_INVALID", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, decision)
}

func (h *AdminHandler) handleListDecisions(w http.ResponseWriter, r *http.Request) {
	writeAdminJSON(w, http.StatusOK, h.decisions.ListDecisions(r.Context()))
}

func (h *AdminHandler) handleListVerifications(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = string(developerDomain.OrgPendingReview)
	}
	list, err := h.verifier.ListVerificationsByStatus(r.Context(), status)
	if err != nil {
		problem.WriteProblem(w, http.StatusInternalServerError, "ERR_VERIFICATION_LIST_FAILED", err.Error(), r.URL.Path)
		return
	}
	if list == nil {
		list = []developerDomain.DeveloperVerification{}
	}
	writeAdminJSON(w, http.StatusOK, list)
}

type verificationDecisionReq struct {
	Approved bool   `json:"approved"`
	Notes    string `json:"notes"`
}

func (h *AdminHandler) handleDecideVerification(w http.ResponseWriter, r *http.Request) {
	var req verificationDecisionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	org, err := h.verifier.DecideVerification(r.Context(), chi.URLParam(r, "orgId"), req.Approved, req.Notes, h.callerID(r))
	if err != nil {
		if err == developerDomain.ErrOrgNotFound {
			problem.WriteProblem(w, http.StatusNotFound, "ERR_ORG_NOT_FOUND", err.Error(), r.URL.Path)
			return
		}
		if err == developerDomain.ErrVerificationNotFound {
			problem.WriteProblem(w, http.StatusNotFound, "ERR_VERIFICATION_NOT_FOUND", err.Error(), r.URL.Path)
			return
		}
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_VERIFICATION_DECISION_FAILED", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, org)
}

type killSwitchReq struct {
	Scope      string `json:"scope"`
	ReasonCode string `json:"reasonCode"`
	Evidence   string `json:"evidence"`
}

func (h *AdminHandler) handleReviewQueue(w http.ResponseWriter, r *http.Request) {
	queue, err := h.catalog.ListReviewQueue(r.Context())
	if err != nil {
		problem.WriteProblem(w, http.StatusInternalServerError, "ERR_REVIEW_QUEUE_FAILED", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, queue)
}

func (h *AdminHandler) handleAdminGetApp(w http.ResponseWriter, r *http.Request) {
	app, err := h.catalog.GetAppByID(r.Context(), chi.URLParam(r, "appId"))
	if err != nil {
		if err == catalogDomain.ErrAppNotFound {
			problem.WriteProblem(w, http.StatusNotFound, "ERR_APP_NOT_FOUND", err.Error(), r.URL.Path)
			return
		}
		problem.WriteProblem(w, http.StatusInternalServerError, "ERR_APP_DETAIL_FAILED", err.Error(), r.URL.Path)
		return
	}
	detail, err := h.catalog.GetAppDetail(r.Context(), app.CanonicalSlug, "")
	if err != nil {
		problem.WriteProblem(w, http.StatusInternalServerError, "ERR_APP_DETAIL_FAILED", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, detail)
}

func (h *AdminHandler) handleKillSwitchActivate(w http.ResponseWriter, r *http.Request) {
	var req killSwitchReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	if strings.TrimSpace(req.ReasonCode) == "" {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_KILLSWITCH_INVALID", "reasonCode is required", r.URL.Path)
		return
	}
	pkg := chi.URLParam(r, "packageName")
	reason := strings.TrimSpace(req.Scope + " " + req.ReasonCode + " " + req.Evidence)
	if err := h.security.TriggerKillSwitch(r.Context(), pkg, reason, h.callerID(r)); err != nil {
		problem.WriteProblem(w, http.StatusInternalServerError, "ERR_KILLSWITCH_FAILED", err.Error(), r.URL.Path)
		return
	}
	state, _ := h.security.KillSwitchState(r.Context(), pkg)
	writeAdminJSON(w, http.StatusOK, state)
}

func (h *AdminHandler) handleKillSwitchStatus(w http.ResponseWriter, r *http.Request) {
	state, err := h.security.KillSwitchState(r.Context(), chi.URLParam(r, "packageName"))
	if err != nil || state == nil {
		problem.WriteProblem(w, http.StatusNotFound, "ERR_KILLSWITCH_NOT_FOUND", "No kill-switch record", r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, state)
}

func (h *AdminHandler) handleKillSwitchLift(w http.ResponseWriter, r *http.Request) {
	pkg := chi.URLParam(r, "packageName")
	if err := h.security.RevokeKillSwitch(r.Context(), pkg); err != nil {
		problem.WriteProblem(w, http.StatusInternalServerError, "ERR_KILLSWITCH_FAILED", err.Error(), r.URL.Path)
		return
	}
	state, _ := h.security.KillSwitchState(r.Context(), pkg)
	writeAdminJSON(w, http.StatusOK, state)
}

// handleJWTRotate activates a fresh signing key. Tokens minted before the
// rotation keep validating through the one-generation grace window, and JWKS
// publishes both keys until the next rotation.
func (h *AdminHandler) handleJWTRotate(w http.ResponseWriter, r *http.Request) {
	if h.jwt == nil {
		problem.WriteProblem(w, http.StatusServiceUnavailable, "ERR_JWT_ROTATION_UNAVAILABLE", "key rotation is not wired", r.URL.Path)
		return
	}
	previousKID := h.jwt.KeyID()
	if err := h.jwt.RotateKeys(nil); err != nil {
		problem.WriteProblem(w, http.StatusInternalServerError, "ERR_JWT_ROTATION_FAILED", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, map[string]any{
		"kid":          h.jwt.KeyID(),
		"previous_kid": previousKID,
		"rotated_by":   h.callerID(r),
		"rotated_at":   time.Now().UTC(),
		"grace":        "tokens signed by the previous key remain valid until the next rotation",
	})
}

func (h *AdminHandler) handleListKillSwitches(w http.ResponseWriter, r *http.Request) {
	list, err := h.security.ListKillSwitches(r.Context())
	if err != nil {
		problem.WriteProblem(w, http.StatusInternalServerError, "ERR_KILLSWITCH_LIST_FAILED", err.Error(), r.URL.Path)
		return
	}
	if list == nil {
		list = []securityDomain.KillSwitchState{}
	}
	writeAdminJSON(w, http.StatusOK, list)
}

func (h *AdminHandler) handleListScans(w http.ResponseWriter, r *http.Request) {
	if artifactID := r.URL.Query().Get("artifactId"); artifactID != "" {
		rep, err := h.security.ScanReport(r.Context(), artifactID)
		if err != nil || rep == nil {
			problem.WriteProblem(w, http.StatusNotFound, "ERR_SCAN_NOT_FOUND", "No scan report", r.URL.Path)
			return
		}
		writeAdminJSON(w, http.StatusOK, rep)
		return
	}
	list, err := h.security.ListScanReports(r.Context())
	if err != nil {
		problem.WriteProblem(w, http.StatusInternalServerError, "ERR_SCAN_LIST_FAILED", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, list)
}

func (h *AdminHandler) handleListConnectors(w http.ResponseWriter, r *http.Request) {
	writeAdminJSON(w, http.StatusOK, h.connectors.ListConnectors(r.Context()))
}

func (h *AdminHandler) handleRegisterConnector(w http.ResponseWriter, r *http.Request) {
	var c console.Connector
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	if strings.TrimSpace(c.ID) == "" {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_CONNECTOR_INVALID", "Connector id is required", r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusCreated, h.connectors.RegisterConnector(r.Context(), c))
}

func (h *AdminHandler) handleTriggerSync(w http.ResponseWriter, r *http.Request) {
	res, err := h.connectors.TriggerSync(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		status := http.StatusBadGateway
		code := "ERR_CONNECTOR_SYNC_FAILED"
		if err == console.ErrConnectorNotFound {
			status, code = http.StatusNotFound, "ERR_CONNECTOR_NOT_FOUND"
		} else if err == console.ErrConnectorNotSyncable {
			status, code = http.StatusConflict, "ERR_CONNECTOR_NOT_SYNCABLE"
		}
		problem.WriteProblem(w, status, code, err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, res)
}

func (h *AdminHandler) handleListSettlements(w http.ResponseWriter, r *http.Request) {
	list := h.payouts.ListPayouts()
	if list == nil {
		list = []settlement.SettlementPayout{}
	}
	writeAdminJSON(w, http.StatusOK, list)
}

type createSettlementReq struct {
	OrgID    string `json:"orgId"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func (h *AdminHandler) handleCreateSettlement(w http.ResponseWriter, r *http.Request) {
	var req createSettlementReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	currency := commerceDomain.Currency(req.Currency)
	if currency == "" {
		currency = commerceDomain.CurrencyKRW
	}
	payout, err := h.payouts.CreatePayoutRequest(r.Context(), req.OrgID, req.Amount, currency, h.callerID(r))
	if err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_SETTLEMENT_CREATE_FAILED", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusCreated, payout)
}

type approveSettlementReq struct {
	ApproverID  string `json:"approver_id"`
	InitiatorID string `json:"initiator_id"`
}

func (h *AdminHandler) handleApproveSettlement(w http.ResponseWriter, r *http.Request) {
	var req approveSettlementReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	payoutID := chi.URLParam(r, "id")
	existing := h.payouts.GetPayout(payoutID)
	if existing == nil {
		problem.WriteProblem(w, http.StatusNotFound, "ERR_SETTLEMENT_NOT_FOUND", "Payout not found", r.URL.Path)
		return
	}
	if req.InitiatorID != "" && req.InitiatorID != existing.CreatedBy {
		problem.WriteProblem(w, http.StatusConflict, "ERR_SETTLEMENT_INITIATOR_MISMATCH", "Initiator does not match payout creator", r.URL.Path)
		return
	}
	payout, err := h.payouts.ApprovePayout(r.Context(), payoutID, req.ApproverID)
	if err != nil {
		status := http.StatusBadRequest
		code := "ERR_SETTLEMENT_APPROVE_FAILED"
		if err == commerceDomain.ErrSelfApprovalForbidden {
			status, code = http.StatusForbidden, "ERR_DUAL_CONTROL_SELF_APPROVAL_FORBIDDEN"
		}
		problem.WriteProblem(w, status, code, err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, payout)
}

func (h *AdminHandler) handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	records, err := h.auditRepo.GetAllRecords()
	if err != nil {
		problem.WriteProblem(w, http.StatusInternalServerError, "ERR_AUDIT_LIST_FAILED", err.Error(), r.URL.Path)
		return
	}
	valid, failedAt, _ := audit.VerifyAuditChain(records)
	if len(records) > limit {
		records = records[len(records)-limit:]
	}
	writeAdminJSON(w, http.StatusOK, map[string]any{
		"chainValid": valid, "chainFailedAt": failedAt, "records": records,
	})
}

func (h *AdminHandler) handleListOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var orders []commerceDomain.Order
	switch {
	case q.Get("user") != "":
		orders = h.orders.ListOrdersByUser(q.Get("user"))
	case q.Get("org") != "":
		orders = h.orders.ListOrdersByOrg(q.Get("org"))
	default:
		orders = h.orders.ListAllOrders()
	}
	if orders == nil {
		orders = []commerceDomain.Order{}
	}
	writeAdminJSON(w, http.StatusOK, orders)
}

func (h *AdminHandler) handleForcedRefund(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "Admin forced refund"
	}
	order, err := h.orders.GetOrder(chi.URLParam(r, "id"))
	if err != nil || order == nil {
		problem.WriteProblem(w, http.StatusNotFound, "ERR_ORDER_NOT_FOUND", "Order not found", r.URL.Path)
		return
	}
	if err := h.refunds.ProcessRefund(r.Context(), order, reason); err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_REFUND_FAILED", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, order)
}

func (h *AdminHandler) handleListAdCampaigns(w http.ResponseWriter, r *http.Request) {
	list := h.campaigns.ListCampaignsByStatus(r.Context(), adsDomain.CampaignStatus(r.URL.Query().Get("status")))
	if list == nil {
		list = []adsDomain.AdCampaign{}
	}
	writeAdminJSON(w, http.StatusOK, list)
}

func (h *AdminHandler) handleReviewAdCampaign(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Approve bool `json:"approve"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	c, err := h.campaigns.ReviewCampaign(r.Context(), chi.URLParam(r, "id"), req.Approve)
	if err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_AD_REVIEW_FAILED", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, c)
}

func (h *AdminHandler) handleAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	events, err := h.eventStore.GetAllEvents(r.Context())
	if err != nil {
		problem.WriteProblem(w, http.StatusInternalServerError, "ERR_ANALYTICS_FAILED", err.Error(), r.URL.Path)
		return
	}
	byName := map[string]int{}
	byPlatform := map[string]int{}
	byCountry := map[string]int{}
	byApp := map[string]int{}
	revenueByCountry := map[string]int64{}
	purchaseCountByCountry := map[string]int{}
	var totalRevenueCents int64
	last24h := 0
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, e := range events {
		byName[e.EventName]++
		byPlatform[e.Platform]++
		byCountry[e.CountryCode]++
		if e.Properties.AppID != "" {
			byApp[e.Properties.AppID]++
		}
		if e.EventName == analytics.EventPurchaseCompleted {
			totalRevenueCents += e.Properties.PriceCents
			revenueByCountry[e.CountryCode] += e.Properties.PriceCents
			purchaseCountByCountry[e.CountryCode]++
		}
		if e.OccurredAt.After(cutoff) {
			last24h++
		}
	}
	writeAdminJSON(w, http.StatusOK, map[string]any{
		"totalEvents":       len(events),
		"last24h":           last24h,
		"totalRevenueCents": totalRevenueCents,
		"byEvent":           topEntries(byName, 20),
		"byPlatform":        topEntries(byPlatform, 10),
		"byCountry":         topEntries(byCountry, 10),
		"byApp":             topEntries(byApp, 10),
		"revenueByCountry":  topRevenueEntries(revenueByCountry, purchaseCountByCountry, 10),
	})
}

func topRevenueEntries(revenue map[string]int64, counts map[string]int, n int) []map[string]any {
	type kv struct {
		key     string
		revenue int64
		count   int
	}
	pairs := make([]kv, 0, len(revenue))
	for k, v := range revenue {
		pairs = append(pairs, kv{k, v, counts[k]})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].revenue > pairs[j].revenue })
	if len(pairs) > n {
		pairs = pairs[:n]
	}
	out := make([]map[string]any, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, map[string]any{"key": p.key, "count": p.count, "revenueCents": p.revenue})
	}
	return out
}

func topEntries(counts map[string]int, n int) []map[string]any {
	type kv struct {
		key   string
		count int
	}
	pairs := make([]kv, 0, len(counts))
	for k, v := range counts {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].count > pairs[j].count })
	if len(pairs) > n {
		pairs = pairs[:n]
	}
	out := make([]map[string]any, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, map[string]any{"key": p.key, "count": p.count})
	}
	return out
}

func (h *AdminHandler) handleListReports(w http.ResponseWriter, r *http.Request) {
	list := h.reports.List(r.Context(), moderationDomain.ReportStatus(r.URL.Query().Get("status")))
	if list == nil {
		list = []moderationDomain.ModerationReport{}
	}
	writeAdminJSON(w, http.StatusOK, list)
}

func (h *AdminHandler) handleClaimReport(w http.ResponseWriter, r *http.Request) {
	rep, err := h.reports.Claim(r.Context(), chi.URLParam(r, "id"), h.callerID(r))
	if err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_REPORT_CLAIM_FAILED", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, rep)
}

func (h *AdminHandler) handleResolveReport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ActionTaken bool   `json:"action_taken"`
		Notes       string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	rep, err := h.reports.Resolve(r.Context(), chi.URLParam(r, "id"), h.callerID(r), req.ActionTaken, req.Notes)
	if err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_REPORT_RESOLVE_FAILED", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, rep)
}

func (h *AdminHandler) handleListAppeals(w http.ResponseWriter, r *http.Request) {
	writeAdminJSON(w, http.StatusOK, h.appeals.ListAppeals())
}

func (h *AdminHandler) handleDecideAppeal(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Decision string `json:"decision"`
		Notes    string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	decision := moderationDomain.AppealStatus(req.Decision)
	if decision != moderationDomain.AppealApproved && decision != moderationDomain.AppealRejected {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_APPEAL_INVALID", "decision must be APPROVED or REJECTED", r.URL.Path)
		return
	}
	appeal, err := h.appeals.DecideAppeal(r.Context(), chi.URLParam(r, "id"), h.callerID(r), decision, req.Notes)
	if err != nil {
		problem.WriteProblem(w, http.StatusBadRequest, "ERR_APPEAL_DECIDE_FAILED", err.Error(), r.URL.Path)
		return
	}
	writeAdminJSON(w, http.StatusOK, appeal)
}
