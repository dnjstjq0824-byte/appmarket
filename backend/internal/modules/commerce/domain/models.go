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
