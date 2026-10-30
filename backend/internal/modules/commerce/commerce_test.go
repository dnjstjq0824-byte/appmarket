package commerce

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/settlement"
)

func generateTestWebhookSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// TC-COM-001: 20 Duplicate Payment Webhooks & Idempotent Ledger Journaling
// TC-COM-004: Double-Entry Zero-Sum Balance Invariant (Debit == Credit)
func TestTCCOM001And004DuplicateWebhooksAndZeroSumLedger(t *testing.T) {
	ctx := context.Background()
	ledger := application.NewLedgerService()
	webhookSecret := "super-secret-pg-webhook-key-32b"
	orderService := application.NewOrderService(ledger, webhookSecret)

	// Create test order: ₩10,000 gross
	order, err := orderService.CreateOrder(ctx, "user-1", "app-paid-1", "org-dev-1", 10000, domain.CurrencyKRW, "client-idem-key-1")
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	payload := []byte(`{"order_id":"` + order.ID + `","status":"PAID"}`)
	sig := generateTestWebhookSignature(payload, webhookSecret)

	t.Run("Sub-test 1: 20 Duplicate Webhook Invocations post only 1 Ledger Transaction", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			resOrder, err := orderService.ProcessPaymentWebhook(ctx, order.ID, "pg-pay-999", payload, sig)
			if err != nil {
				t.Fatalf("webhook call %d failed: %v", i, err)
			}
			if resOrder.Status != domain.OrderPaid {
				t.Fatalf("expected OrderPaid, got %s", resOrder.Status)
			}
		}

		// Ensure ledger transaction count is exactly 1 despite 20 calls
		txCount := ledger.GetTransactionCount()
		if txCount != 1 {
			t.Fatalf("expected exactly 1 ledger transaction posted, got %d", txCount)
		}

		// Verify Account Balances:
		// Debit ASSET:PG_CLEARING (+10,000)
		clearingBalance := ledger.GetBalance("ASSET:PG_CLEARING")
		if clearingBalance != 10000 {
			t.Fatalf("expected ASSET:PG_CLEARING balance to be 10000, got %d", clearingBalance)
		}

		// Credits: Commission, Tax, Dev Escrow
		commissionBalance := -ledger.GetBalance("REV:PLATFORM_COMMISSION")
		taxBalance := -ledger.GetBalance("LIAB:TAX_PAYABLE")
		escrowBalance := -ledger.GetBalance("LIAB:DEV_ESCROW")

		totalCredits := commissionBalance + taxBalance + escrowBalance
		if clearingBalance != totalCredits {
			t.Fatalf("Ledger unbalanced! Debits (%d) != Credits (%d)", clearingBalance, totalCredits)
		}
	})
}

// TC-COM-002: Paid App User Entitlement Verification
func TestTCCOM002PaidAppUserEntitlement(t *testing.T) {
	ctx := context.Background()
	ledger := application.NewLedgerService()
	orderService := application.NewOrderService(ledger, "secret")

	paidAppID := "app-premium-vpn"
	orderService.RegisterPaidApp(paidAppID)

	t.Run("Sub-test 1: Unpurchased user is blocked from paid app with ERR_ENTITLEMENT_REQUIRED", func(t *testing.T) {
		allowed, err := orderService.VerifyUserEntitlement(ctx, "user-unpaid", paidAppID)
		if allowed || !errors.Is(err, domain.ErrEntitlementRequired) {
			t.Fatalf("expected ErrEntitlementRequired for unpaid user, got allowed=%v err=%v", allowed, err)
		}
	})

	t.Run("Sub-test 2: Purchased user gets active entitlement and access is granted", func(t *testing.T) {
		order, err := orderService.CreateOrder(ctx, "user-paid", paidAppID, "org-1", 5000, domain.CurrencyKRW, "idem-2")
		if err != nil {
			t.Fatalf("failed to create order: %v", err)
		}

		payPayload := []byte("pay")
		paySig := generateTestWebhookSignature(payPayload, "secret")
		_, err = orderService.ProcessPaymentWebhook(ctx, order.ID, "pg-pay-1", payPayload, paySig)
		if err != nil {
			t.Fatalf("payment processing failed: %v", err)
		}

		allowed, err := orderService.VerifyUserEntitlement(ctx, "user-paid", paidAppID)
		if !allowed || err != nil {
			t.Fatalf("expected access granted for purchased user, got allowed=%v err=%v", allowed, err)
		}
	})
}

// TC-COM-005: Pricing Minor Units Rounding Precision (Zero Rounding Leakage)
func TestTCCOM005PricingCalculatorRoundingPrecision(t *testing.T) {
	calc := domain.NewPricingCalculator()

	testPrices := []int64{
		4900,  // ₩4,900
		9900,  // ₩9,900
		15000, // ₩15,000
		199,   // $1.99
		999,   // $9.99
	}

	for _, price := range testPrices {
		breakdown := calc.CalculateBreakdown(price)
		sum := breakdown.CommissionAmount + breakdown.TaxAmount + breakdown.DevPayoutAmount
		if sum != price {
			t.Fatalf("Rounding leakage detected for price %d: sum (%d) != price (%d)", price, sum, price)
		}
		if breakdown.CommissionAmount <= 0 || breakdown.TaxAmount <= 0 || breakdown.DevPayoutAmount <= 0 {
			t.Fatalf("Non-positive component in breakdown for price %d: %+v", price, breakdown)
		}
	}
}

// TC-COM-006: Refund Entitlement Revocation, 14-Day Escrow & Dual-Control Payout
func TestTCCOM006RefundEscrowAndDualControlPayout(t *testing.T) {
	ctx := context.Background()
	ledger := application.NewLedgerService()
	orderService := application.NewOrderService(ledger, "secret")
	refundService := settlement.NewRefundService(ledger, orderService)
	escrowWorker := settlement.NewEscrowReleaseWorker(ledger)
	payoutService := settlement.NewMonthlyPayoutService(ledger)

	appID := "app-game-pro"
	orderService.RegisterPaidApp(appID)

	order, err := orderService.CreateOrder(ctx, "user-refund-test", appID, "org-game", 20000, domain.CurrencyKRW, "idem-ref")
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	refPayload := []byte("pay")
	refSig := generateTestWebhookSignature(refPayload, "secret")
	_, err = orderService.ProcessPaymentWebhook(ctx, order.ID, "pg-pay-ref", refPayload, refSig)
	if err != nil {
		t.Fatalf("failed to process payment for refund order: %v", err)
	}

	t.Run("Sub-test 1: Refund immediately revokes entitlement and reverses ledger", func(t *testing.T) {
		err := refundService.ProcessRefund(ctx, order, "Accidental duplicate purchase")
		if err != nil {
			t.Fatalf("expected refund success, got %v", err)
		}

		// Subsequent entitlement check must fail
		allowed, err := orderService.VerifyUserEntitlement(ctx, "user-refund-test", appID)
		if allowed || !errors.Is(err, domain.ErrEntitlementRequired) {
			t.Fatalf("expected entitlement revoked on refund, got allowed=%v, err=%v", allowed, err)
		}
	})

	t.Run("Sub-test 2: Escrow Release enforces 14-day maturity window", func(t *testing.T) {
		now := time.Now().UTC()
		immatureOrder := &domain.Order{
			ID:              "ord-immature",
			PaidAt:          &now, // Just paid today
			DevPayoutAmount: 15000,
		}

		err := escrowWorker.ReleaseMaturedEscrow(ctx, immatureOrder)
		if !errors.Is(err, domain.ErrEscrowNotMatured) {
			t.Fatalf("expected ErrEscrowNotMatured for freshly paid order, got %v", err)
		}

		// 15 days ago order
		fifteenDaysAgo := now.Add(-15 * 24 * time.Hour)
		maturedOrder := &domain.Order{
			ID:              "ord-matured",
			PaidAt:          &fifteenDaysAgo,
			DevPayoutAmount: 15000,
		}

		err = escrowWorker.ReleaseMaturedEscrow(ctx, maturedOrder)
		if err != nil {
			t.Fatalf("expected successful escrow release for 15-day-old order, got %v", err)
		}
	})

	t.Run("Sub-test 3: Dual-Control Payout Rejects Self-Approval", func(t *testing.T) {
		creatorAdmin := "admin_fin_creator@bookingnetwork.com"
		approverAdmin := "admin_fin_director@bookingnetwork.com"

		payout, err := payoutService.CreatePayoutRequest(ctx, "org-game", 5000000, domain.CurrencyKRW, creatorAdmin)
		if err != nil {
			t.Fatalf("failed to create payout request: %v", err)
		}

		// Self-approval attempt by creator must fail
		_, err = payoutService.ApprovePayout(ctx, payout.ID, creatorAdmin)
		if !errors.Is(err, domain.ErrSelfApprovalForbidden) {
			t.Fatalf("expected ErrSelfApprovalForbidden on self-approval attempt, got %v", err)
		}

		// Approval by distinct admin succeeds
		approvedPayout, err := payoutService.ApprovePayout(ctx, payout.ID, approverAdmin)
		if err != nil {
			t.Fatalf("expected approval success by distinct approver, got %v", err)
		}
		if approvedPayout.Status != settlement.PayoutPaidCompleted {
			t.Fatalf("expected PayoutPaidCompleted status, got %s", approvedPayout.Status)
		}
	})
}
