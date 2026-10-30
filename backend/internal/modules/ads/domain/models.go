package domain

import (
	"errors"
	"time"
)

var (
	ErrAdCampaignNotFound       = errors.New("ERR_AD_CAMPAIGN_NOT_FOUND")
	ErrClickFraudDuplicate      = errors.New("ERR_CLICK_FRAUD_DUPLICATE_DEDUPLICATED")
	ErrClickFraudFastBot        = errors.New("ERR_CLICK_FRAUD_BOT_FAST_CLICK")
	ErrImpressionTokenExpired   = errors.New("ERR_IMPRESSION_TOKEN_EXPIRED")
	ErrDailyBudgetCapReached    = errors.New("ERR_DAILY_BUDGET_CAP_REACHED")
)

type CampaignStatus string

const (
	CampaignActive               CampaignStatus = "ACTIVE"
	CampaignPausedBudgetCap      CampaignStatus = "PAUSED_DAILY_BUDGET_CAP"
	CampaignPausedManual         CampaignStatus = "PAUSED_MANUAL"
)

type AdCampaign struct {
	ID             string         `json:"id"`
	OrgID          string         `json:"org_id"`
	AppID          string         `json:"app_id"`
	AppName        string         `json:"app_name"`
	TargetKeywords []string       `json:"target_keywords"`
	MaxBidCPC      int64          `json:"max_bid_cpc"`      // In minor units
	QualityScore   float64        `json:"quality_score"`    // 1.0 to 10.0
	DailyBudget    int64          `json:"daily_budget"`     // Max spend per day
	SpentToday     int64          `json:"spent_today"`
	Status         CampaignStatus `json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
}

type ImpressionToken struct {
	TokenID     string    `json:"token_id"`
	CampaignID  string    `json:"campaign_id"`
	SlotIndex   int       `json:"slot_index"`
	DeviceHash  string    `json:"device_hash"`
	ClearingCPC int64     `json:"clearing_cpc"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type ClickResult struct {
	ClickID     string `json:"click_id"`
	CampaignID  string `json:"campaign_id"`
	ChargedAmount int64 `json:"charged_amount"`
	IsFraud     bool   `json:"is_fraud"`
	FraudReason string `json:"fraud_reason,omitempty"`
}
