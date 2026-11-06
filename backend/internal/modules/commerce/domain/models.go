package domain

import (
	"errors"
	"time"
)

var (
	ErrLedgerUnbalanced               = errors.New("ERR_LEDGER_UNBALANCED_DEBIT_NOT_EQUAL_CREDIT")
	ErrDuplicateTransactionIdempotent = errors.New("ERR_DUPLICATE_TRANSACTION_IDEMPOTENT")
	ErrEntitlementRequired            = errors.New("ERR_ENTITLEMENT_REQUIRED")
	ErrOrderAlreadyPaid               = errors.New("ERR_ORDER_ALREADY_PAID")
	ErrOrderAlreadyRefunded           = errors.New("ERR_ORDER_ALREADY_REFUNDED")
	ErrInvalidWebhookSignature        = errors.New("ERR_INVALID_WEBHOOK_SIGNATURE")
	ErrSelfApprovalForbidden          = errors.New("ERR_SELF_APPROVAL_FORBIDDEN_DUAL_CONTROL")
	ErrEscrowNotMatured               = errors.New("ERR_ESCROW_NOT_MATURED_14_DAYS")
)

type Currency string

const (
	CurrencyKRW Currency = "KRW"
	CurrencyUSD Currency = "USD"
)

type OrderStatus string

const (
	OrderPending   OrderStatus = "PENDING"
	OrderPaid      OrderStatus = "PAID"
	OrderRefunded  OrderStatus = "REFUNDED"
	OrderCancelled OrderStatus = "CANCELLED"
)

type EntitlementStatus string

const (
	EntitlementActive   EntitlementStatus = "ACTIVE"
	EntitlementRevoked  EntitlementStatus = "REVOKED_REFUNDED"
	EntitlementExpired  EntitlementStatus = "EXPIRED"
)

type Order struct {
	ID             string      `json:"id"`
	UserID         string      `json:"user_id"`
	AppID          string      `json:"app_id"`
	OrgID          string      `json:"org_id"`
	Currency       Currency    `json:"currency"`
	GrossAmount    int64       `json:"gross_amount"`     // In minor units (KRW Won, USD Cents)
	TaxAmount      int64       `json:"tax_amount"`       // VAT (10%)
	CommissionAmount int64     `json:"commission_amount"`// 15% Platform fee
	DevPayoutAmount int64      `json:"dev_payout_amount"`// 85% - Tax
	Status         OrderStatus `json:"status"`
	IdempotencyKey string      `json:"idempotency_key"`
	PGPaymentID    string      `json:"pg_payment_id,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	PaidAt         *time.Time  `json:"paid_at,omitempty"`
	RefundedAt     *time.Time  `json:"refunded_at,omitempty"`
}

type UserEntitlement struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	AppID     string            `json:"app_id"`
	OrderID   string            `json:"order_id"`
	Status    EntitlementStatus `json:"status"`
	GrantedAt time.Time         `json:"granted_at"`
	RevokedAt *time.Time        `json:"revoked_at,omitempty"`
}

type LedgerEntryType string

const (
	EntryDebit  LedgerEntryType = "DEBIT"
	EntryCredit LedgerEntryType = "CREDIT"
)

type LedgerEntry struct {
	Account   string          `json:"account"` // ASSET:PG_CLEARING, REV:COMMISSION, LIAB:TAX, LIAB:DEV_ESCROW
	Type      LedgerEntryType `json:"type"`
	Amount    int64           `json:"amount"` // Positive minor units
}

type LedgerTransaction struct {
	ID             string        `json:"id"`
	ReferenceID    string        `json:"reference_id"` // OrderID or RefundID
	Description    string        `json:"description"`
	Entries        []LedgerEntry `json:"entries"`
	CreatedAt      time.Time     `json:"created_at"`
}

type InvoiceStatus string

const (
	InvoiceIssued   InvoiceStatus = "ISSUED"
	InvoiceRefunded InvoiceStatus = "REFUNDED"
	InvoiceVoid     InvoiceStatus = "VOID"
)

// Invoice is the buyer-facing receipt issued exactly once per paid order.
type Invoice struct {
	ID               string        `json:"id"`
	InvoiceNumber    string        `json:"invoice_number"`
	OrderID          string        `json:"order_id"`
	UserID           string        `json:"user_id"`
	AppID            string        `json:"app_id"`
	OrgID            string        `json:"org_id"`
	Currency         Currency      `json:"currency"`
	GrossAmount      int64         `json:"gross_amount"`
	TaxAmount        int64         `json:"tax_amount"`
	CommissionAmount int64         `json:"commission_amount"`
	DevPayoutAmount  int64         `json:"dev_payout_amount"`
	PGPaymentID      string        `json:"pg_payment_id,omitempty"`
	Status           InvoiceStatus `json:"status"`
	IssuedAt         time.Time     `json:"issued_at"`
}

type SubscriptionStatus string

const (
	SubscriptionActive    SubscriptionStatus = "ACTIVE"
	SubscriptionCancelled SubscriptionStatus = "CANCELLED"
	SubscriptionExpired   SubscriptionStatus = "EXPIRED"
)

type BillingInterval string

const (
	BillingMonthly BillingInterval = "MONTHLY"
	BillingYearly  BillingInterval = "YEARLY"
)

var (
	ErrSubscriptionNotFound       = errors.New("ERR_SUBSCRIPTION_NOT_FOUND")
	ErrSubscriptionAlreadyEnded   = errors.New("ERR_SUBSCRIPTION_ALREADY_ENDED")
	ErrInvalidBillingInterval     = errors.New("ERR_INVALID_BILLING_INTERVAL")
	ErrInvalidSubscriptionAmount  = errors.New("ERR_INVALID_SUBSCRIPTION_AMOUNT")
	ErrSubscriptionAccessDenied   = errors.New("ERR_SUBSCRIPTION_ACCESS_DENIED")
)

type DiscountType string

const (
	DiscountPercentage DiscountType = "PERCENTAGE"
	DiscountFixedMinor DiscountType = "FIXED_MINOR"
	DiscountFreeRedeem DiscountType = "FREE_REDEEM"
)

var (
	ErrPromoNotFound      = errors.New("ERR_PROMO_NOT_FOUND")
	ErrPromoExpired       = errors.New("ERR_PROMO_EXPIRED")
	ErrPromoExhausted     = errors.New("ERR_PROMO_EXHAUSTED")
	ErrPromoAlreadyUsed   = errors.New("ERR_PROMO_ALREADY_USED")
	ErrPromoInvalid       = errors.New("ERR_PROMO_INVALID")
	ErrPriceTierInvalid   = errors.New("ERR_PRICE_TIER_INVALID")
)

// Promotion is a discount campaign; PromoCode holds its redeemable codes.
type Promotion struct {
	ID            string       `json:"id"`
	AppID         string       `json:"app_id"`
	CampaignName  string       `json:"campaign_name"`
	DiscountType  DiscountType `json:"discount_type"`
	DiscountValue int          `json:"discount_value"`
	StartsAt      time.Time    `json:"starts_at"`
	EndsAt        time.Time    `json:"ends_at"`
	MaxUses       int          `json:"max_uses"`
	UsedCount     int          `json:"used_count"`
	CreatedAt     time.Time    `json:"created_at"`
}

type PromoCode struct {
	ID          string `json:"id"`
	PromotionID string `json:"promotion_id"`
	Code        string `json:"code"`
	MaxUses     int    `json:"max_uses"`
	UsedCount   int    `json:"used_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// PriceTier is a per-country list price for an app.
type PriceTier struct {
	AppID        string   `json:"app_id"`
	CountryCode  string   `json:"country_code"`
	Currency     Currency `json:"currency"`
	ListPriceMinor int64  `json:"list_price_minor"`
	TaxInclusive bool     `json:"tax_inclusive"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Subscription is a recurring billing relationship for SUBSCRIPTION_READY apps.
type Subscription struct {
	ID                 string             `json:"id"`
	UserID             string             `json:"user_id"`
	AppID              string             `json:"app_id"`
	OrgID              string             `json:"org_id"`
	BillingInterval    BillingInterval    `json:"billing_interval"`
	AmountMinor        int64              `json:"amount_minor"`
	Currency           Currency           `json:"currency"`
	Status             SubscriptionStatus `json:"status"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	CancelledAt        *time.Time         `json:"cancelled_at,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
}

