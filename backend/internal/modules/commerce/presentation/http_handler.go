package presentation

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/settlement"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
)

type IdentityResolver func(*http.Request) (string, error)

type CommerceHandler struct {
	orderService  *application.OrderService
	refundService *settlement.RefundService
	identity      IdentityResolver
}

func NewCommerceHandler(
	orderService *application.OrderService,
	refundService *settlement.RefundService,
	identity IdentityResolver,
) *CommerceHandler {
	return &CommerceHandler{
		orderService:  orderService,
		refundService: refundService,
		identity:      identity,
	}
}

func (h *CommerceHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Public Webhook (verified via HMAC-SHA256 signature)
	r.Post("/webhooks/pg", h.handlePGWebhook)

	// Protected Endpoints
	r.Group(func(protected chi.Router) {
		protected.Use(h.requireAuth)
		protected.Post("/orders", h.handleCreateOrder)
		protected.Post("/orders/{orderId}/checkout", h.handleCreateCheckoutSession)
		protected.Post("/orders/{orderId}/refund", h.handleRequestRefund)
		protected.Get("/entitlements", h.handleListEntitlements)
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

type createOrderReq struct {
	AppID          string          `json:"app_id"`
	OrgID          string          `json:"org_id"`
	GrossAmount    int64           `json:"gross_amount"`
	Currency       domain.Currency `json:"currency"`
	IdempotencyKey string          `json:"idempotency_key"`
}

func (h *CommerceHandler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)

	var req createOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}

	if req.GrossAmount <= 0 {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_AMOUNT", "Gross amount must be positive", r.URL.Path)
		return
	}
	if req.Currency == "" {
		req.Currency = domain.CurrencyKRW
	}

	order, err := h.orderService.CreateOrder(r.Context(), userID, req.AppID, req.OrgID, req.GrossAmount, req.Currency, req.IdempotencyKey)
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_ORDER_CREATION_FAILED", err.Error(), r.URL.Path)
		return
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

	// Toss Payments / PortOne client checkout session config
	checkoutSession := map[string]any{
		"order_id":      order.ID,
		"order_name":    "App Market Digital License",
		"amount":        order.GrossAmount,
		"currency":      order.Currency,
		"customer_key":  order.UserID,
		"status":        order.Status,
		"success_url":   "https://appmarket.bookingnetwork.internal/checkout/success",
		"fail_url":      "https://appmarket.bookingnetwork.internal/checkout/fail",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(checkoutSession)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "SUCCESS",
		"order_id": order.ID,
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
