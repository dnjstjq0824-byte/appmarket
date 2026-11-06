package presentation

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/settlement"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/sanitize"
)

type IdentityResolver func(*http.Request) (string, error)

type CommerceHandler struct {
	orderService      *application.OrderService
	refundService     *settlement.RefundService
	invoiceService    *settlement.InvoiceService
	subscriptionService *settlement.SubscriptionService
	promoService      *application.PromoService
	pricing           *application.PricingService
	identity          IdentityResolver
	pgClient          application.PGClient
	publicBaseURL     string
}

// WithPricing attaches the shared pricing service for public price reads and
// order amount validation.
func (h *CommerceHandler) WithPricing(pricing *application.PricingService) *CommerceHandler {
	h.pricing = pricing
	return h
}

func NewCommerceHandler(
	orderService *application.OrderService,
	refundService *settlement.RefundService,
	invoiceService *settlement.InvoiceService,
	subscriptionService *settlement.SubscriptionService,
	identity IdentityResolver,
	pgClient application.PGClient,
	publicBaseURL string,
	promoService *application.PromoService,
) *CommerceHandler {
	if invoiceService == nil {
		invoiceService = settlement.NewInvoiceService()
	}
	if subscriptionService == nil {
		subscriptionService = settlement.NewSubscriptionService()
	}
	if promoService == nil {
		promoService = application.NewPromoService()
	}
	return &CommerceHandler{
		orderService:      orderService,
		refundService:     refundService,
		invoiceService:    invoiceService,
		subscriptionService: subscriptionService,
		promoService:      promoService,
		identity:          identity,
		pgClient:          pgClient,
		publicBaseURL:     publicBaseURL,
	}
}

func (h *CommerceHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Public Webhook (verified via HMAC-SHA256 signature)
	r.Post("/webhooks/pg", h.handlePGWebhook)

	// Public storefront price list (same tiers developers configure).
	r.Get("/apps/{appId}/prices", h.handleListPublicPrices)

	// Protected Endpoints
	r.Group(func(protected chi.Router) {
		protected.Use(h.requireAuth)
		protected.Post("/orders", h.handleCreateOrder)
		protected.Post("/orders/{orderId}/checkout", h.handleCreateCheckoutSession)
		protected.Post("/orders/{orderId}/refund", h.handleRequestRefund)
		protected.Get("/entitlements", h.handleListEntitlements)
		protected.Get("/invoices", h.handleListInvoices)
		protected.Get("/invoices/{invoiceId}", h.handleGetInvoice)
		protected.Post("/subscriptions", h.handleCreateSubscription)
		protected.Get("/subscriptions", h.handleListSubscriptions)
		protected.Post("/subscriptions/{subscriptionId}/cancel", h.handleCancelSubscription)
		protected.Post("/promos/redeem", h.handleRedeemPromo)
		protected.Post("/promos/validate", h.handleValidatePromo)
		protected.Get("/orders", h.handleListOrders)
	})

	return r
}

func (h *CommerceHandler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.identity == nil {
			errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Authentication required", r.URL.Path)
			return
		}
		userID, err := h.identity(r)
		if err != nil || userID == "" {
			errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Authentication required", r.URL.Path)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *CommerceHandler) handleListPublicPrices(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	var tiers []domain.PriceTier
	if h.pricing != nil {
		tiers = h.pricing.ListTiers(r.Context(), appID)
	}
	if tiers == nil {
		tiers = []domain.PriceTier{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": tiers})
}

type createOrderReq struct {
	AppID          string          `json:"app_id"`
	OrgID          string          `json:"org_id"`
	GrossAmount    int64           `json:"gross_amount"`
	Currency       domain.Currency `json:"currency"`
	IdempotencyKey string          `json:"idempotency_key"`
	PromoCode      string          `json:"promo_code"`
}

func (h *CommerceHandler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)

	var req createOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}

	if req.Currency == "" {
		req.Currency = domain.CurrencyKRW
	}

	// When the developer configured list prices, the charged amount must match
	// the (promo-adjusted) list price exactly; otherwise clients could underpay.
	if h.pricing != nil {
		tiers := h.pricing.ListTiers(r.Context(), req.AppID)
		if len(tiers) > 0 {
			var promo *domain.Promotion
			if req.PromoCode != "" {
				if h.promoService == nil {
					errors.WriteProblem(w, http.StatusServiceUnavailable, "ERR_PROMO_UNAVAILABLE", "Promotion service is not configured", r.URL.Path)
					return
				}
				// Preview first: a mismatched amount must not burn the code.
				preview, err := h.promoService.ValidateWithKey(r.Context(), req.PromoCode, req.AppID, userID, req.IdempotencyKey)
				if err != nil {
					writePromoProblem(w, err, r.URL.Path)
					return
				}
				promo = preview
			}
			matched := false
			for _, t := range tiers {
				if t.Currency == req.Currency && application.ApplyDiscount(t.ListPriceMinor, promo) == req.GrossAmount {
					matched = true
					break
				}
			}
			if !matched {
				errors.WriteProblem(w, http.StatusUnprocessableEntity, "ERR_ORDER_AMOUNT_MISMATCH", "Gross amount does not match the app list price", r.URL.Path)
				return
			}
			if req.PromoCode != "" {
				if _, err := h.promoService.RedeemWithKey(r.Context(), req.PromoCode, req.AppID, userID, req.IdempotencyKey); err != nil {
					writePromoProblem(w, err, r.URL.Path)
					return
				}
			}
		} else if req.PromoCode != "" {
			errors.WriteProblem(w, http.StatusUnprocessableEntity, "ERR_ORDER_AMOUNT_MISMATCH", "Promo codes require a configured list price", r.URL.Path)
			return
		}
	}

	if req.GrossAmount < 0 || (req.GrossAmount == 0 && req.PromoCode == "") {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_AMOUNT", "Gross amount must be positive", r.URL.Path)
		return
	}

	order, err := h.orderService.CreateOrder(r.Context(), userID, req.AppID, req.OrgID, req.GrossAmount, req.Currency, req.IdempotencyKey)
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_ORDER_CREATION_FAILED", err.Error(), r.URL.Path)
		return
	}
	if order.GrossAmount == 0 {
		order, err = h.orderService.CompleteZeroValueOrder(order.ID)
		if err != nil {
			errors.WriteProblem(w, http.StatusBadRequest, "ERR_ORDER_CREATION_FAILED", err.Error(), r.URL.Path)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(order)
}

func (h *CommerceHandler) handleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")
	userID, _ := h.identity(r)

	order, err := h.orderService.GetOrder(orderID)
	if err != nil || order == nil {
		errors.WriteProblem(w, http.StatusNotFound, "ERR_ORDER_NOT_FOUND", "Order not found", r.URL.Path)
		return
	}
	if order.UserID != userID {
		errors.WriteProblem(w, http.StatusForbidden, "ERR_FORBIDDEN", "Forbidden order access", r.URL.Path)
		return
	}
	if order.Status != domain.OrderPending {
		errors.WriteProblem(w, http.StatusConflict, "ERR_ORDER_NOT_PAYABLE", "Only pending orders can start checkout", r.URL.Path)
		return
	}
	if h.pgClient == nil {
		errors.WriteProblem(w, http.StatusServiceUnavailable, "ERR_PG_UNAVAILABLE", "Payment gateway is not configured", r.URL.Path)
		return
	}

	base := strings.TrimRight(h.publicBaseURL, "/")
	session, err := h.pgClient.CreateCheckoutSession(r.Context(), application.PGSessionRequest{
		OrderID:     order.ID,
		AmountKRW:   order.GrossAmount,
		OrderName:   "App Market Digital License",
		CustomerKey: order.UserID,
		SuccessURL:  base + "/checkout/success",
		FailURL:     base + "/checkout/fail",
	})
	if err != nil {
		errors.WriteProblem(w, http.StatusBadGateway, "ERR_PG_SESSION_FAILED", err.Error(), r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(session)
}

type pgWebhookBody struct {
	OrderID     string `json:"order_id"`
	PGPaymentID string `json:"pg_payment_id"`
	Status      string `json:"status"`
}

func (h *CommerceHandler) handlePGWebhook(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get("X-PG-Signature")
	if signature == "" {
		signature = r.Header.Get("X-Signature")
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", "Failed to read webhook payload", r.URL.Path)
		return
	}

	var body pgWebhookBody
	if err := json.Unmarshal(payload, &body); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", "Invalid JSON payload", r.URL.Path)
		return
	}

	order, err := h.orderService.ProcessPaymentWebhook(r.Context(), body.OrderID, body.PGPaymentID, payload, signature)
	if err != nil {
		if err == domain.ErrInvalidWebhookSignature {
			errors.WriteProblem(w, http.StatusUnauthorized, "ERR_INVALID_WEBHOOK_SIGNATURE", err.Error(), r.URL.Path)
			return
		}
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_WEBHOOK_PROCESS_FAILED", err.Error(), r.URL.Path)
		return
	}

	invoice := h.invoiceService.IssueFromOrder(order)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":         "SUCCESS",
		"order_id":       order.ID,
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
	})
}

type refundReq struct {
	Reason string `json:"reason"`
}

func (h *CommerceHandler) handleRequestRefund(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")
	userID, _ := h.identity(r)

	var req refundReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Customer requested cancellation within 48h statutory window"
	}
	req.Reason = sanitize.Strict(req.Reason)

	order, err := h.orderService.GetOrder(orderID)
	if err != nil || order == nil {
		errors.WriteProblem(w, http.StatusNotFound, "ERR_ORDER_NOT_FOUND", "Order not found", r.URL.Path)
		return
	}
	if order.UserID != userID {
		errors.WriteProblem(w, http.StatusForbidden, "ERR_FORBIDDEN", "Forbidden order access", r.URL.Path)
		return
	}

	if err := h.refundService.ProcessRefund(r.Context(), order, req.Reason); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_REFUND_FAILED", err.Error(), r.URL.Path)
		return
	}
	h.invoiceService.MarkRefunded(order.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(order)
}

func (h *CommerceHandler) handleListEntitlements(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)
	entitlements := h.orderService.ListUserEntitlements(userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(entitlements)
}

func (h *CommerceHandler) handleListInvoices(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)
	invoices := h.invoiceService.ListByUser(userID)
	if invoices == nil {
		invoices = []domain.Invoice{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(invoices)
}

func (h *CommerceHandler) handleGetInvoice(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)
	invoice := h.invoiceService.Get(chi.URLParam(r, "invoiceId"))
	if invoice == nil {
		errors.WriteProblem(w, http.StatusNotFound, "ERR_INVOICE_NOT_FOUND", "Invoice not found", r.URL.Path)
		return
	}
	if invoice.UserID != userID {
		errors.WriteProblem(w, http.StatusForbidden, "ERR_FORBIDDEN", "Forbidden invoice access", r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(invoice)
}

type createSubscriptionReq struct {
	AppID           string                 `json:"app_id"`
	OrgID           string                 `json:"org_id"`
	BillingInterval domain.BillingInterval `json:"billing_interval"`
	AmountMinor     int64                  `json:"amount_minor"`
	Currency        domain.Currency        `json:"currency"`
}

func (h *CommerceHandler) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)

	var req createSubscriptionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	sub, err := h.subscriptionService.CreateSubscription(userID, req.AppID, req.OrgID, req.BillingInterval, req.AmountMinor, req.Currency)
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_SUBSCRIPTION_CREATE_FAILED", err.Error(), r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sub)
}

func (h *CommerceHandler) handleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)
	subs := h.subscriptionService.ListByUser(userID)
	if subs == nil {
		subs = []domain.Subscription{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(subs)
}

type redeemPromoReq struct {
	Code  string `json:"code"`
	AppID string `json:"appId"`
}

func (h *CommerceHandler) handleRedeemPromo(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)
	var req redeemPromoReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	redemption, err := h.promoService.Redeem(r.Context(), req.Code, req.AppID, userID)
	if err != nil {
		status := http.StatusBadRequest
		switch err {
		case domain.ErrPromoNotFound:
			status = http.StatusNotFound
		case domain.ErrPromoExpired:
			status = http.StatusGone
		case domain.ErrPromoExhausted, domain.ErrPromoAlreadyUsed:
			status = http.StatusConflict
		}
		errors.WriteProblem(w, status, "ERR_PROMO_REDEEM_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(redemption)
}

func (h *CommerceHandler) handleValidatePromo(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)
	var req redeemPromoReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	promo, err := h.promoService.Validate(r.Context(), req.Code, req.AppID, userID)
	if err != nil {
		status := http.StatusBadRequest
		switch err {
		case domain.ErrPromoNotFound:
			status = http.StatusNotFound
		case domain.ErrPromoExpired:
			status = http.StatusGone
		case domain.ErrPromoExhausted, domain.ErrPromoAlreadyUsed:
			status = http.StatusConflict
		}
		errors.WriteProblem(w, status, "ERR_PROMO_VALIDATE_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"promotion": promo, "code": req.Code})
}

func writePromoProblem(w http.ResponseWriter, err error, path string) {
	status := http.StatusBadRequest
	switch err {
	case domain.ErrPromoNotFound:
		status = http.StatusNotFound
	case domain.ErrPromoExpired:
		status = http.StatusGone
	case domain.ErrPromoExhausted, domain.ErrPromoAlreadyUsed:
		status = http.StatusConflict
	}
	errors.WriteProblem(w, status, "ERR_PROMO_REDEEM_FAILED", err.Error(), path)
}

func (h *CommerceHandler) handleListOrders(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)
	orders := h.orderService.ListOrdersByUser(userID)
	if orders == nil {
		orders = []domain.Order{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(orders)
}

func (h *CommerceHandler) handleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)
	sub, err := h.subscriptionService.CancelSubscription(chi.URLParam(r, "subscriptionId"), userID)
	if err != nil {
		status := http.StatusBadRequest
		if err == domain.ErrSubscriptionNotFound {
			status = http.StatusNotFound
		} else if err == domain.ErrSubscriptionAccessDenied {
			status = http.StatusForbidden
		} else if err == domain.ErrSubscriptionAlreadyEnded {
			status = http.StatusConflict
		}
		errors.WriteProblem(w, status, "ERR_SUBSCRIPTION_CANCEL_FAILED", err.Error(), r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(sub)
}

