package application

import (
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/domain"
)

type AntiFraudDetector struct {
	mu            sync.Mutex
	userHistory   map[string][]time.Time
	ipHistory     map[string][]time.Time
}

func NewAntiFraudDetector() *AntiFraudDetector {
	return &AntiFraudDetector{
		userHistory: make(map[string][]time.Time),
		ipHistory:   make(map[string][]time.Time),
	}
}

type FraudEvaluationParams struct {
	UserID          string
	ClientIP        string
	AccountAgeHours float64
	Rating          int
	Comment         string
}

// EvaluateReview determines fraud probability score and review publishing status
func (d *AntiFraudDetector) EvaluateReview(params FraudEvaluationParams) (float64, domain.ReviewStatus) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	var fraudScore float64 = 0.0

	// 1. New account suspicion: accounts created < 2 hours ago submitting extreme reviews
	if params.AccountAgeHours < 2.0 && (params.Rating == 1 || params.Rating == 5) {
		fraudScore += 0.4
	}

	// 2. Velocity check: rapid consecutive reviews from same user (> 5 within 2 minutes)
	timestamps := d.userHistory[params.UserID]
	recentCount := 0
	for _, t := range timestamps {
		if now.Sub(t) < 2*time.Minute {
			recentCount++
		}
	}
	if recentCount >= 5 {
		fraudScore += 0.5
	}

	// 3. IP cluster check (> 10 reviews from same IP in 5 minutes)
	ipTimes := d.ipHistory[params.ClientIP]
	ipRecentCount := 0
	for _, t := range ipTimes {
		if now.Sub(t) < 5*time.Minute {
			ipRecentCount++
		}
	}
	if ipRecentCount >= 10 {
		fraudScore += 0.4
	}

	// Update history
	d.userHistory[params.UserID] = append(d.userHistory[params.UserID], now)
	d.ipHistory[params.ClientIP] = append(d.ipHistory[params.ClientIP], now)

	if fraudScore >= 0.7 {
		return fraudScore, domain.ReviewQuarantined
	} else if fraudScore >= 0.4 {
		return fraudScore, domain.ReviewFlaggedSuspicious
	}

	return fraudScore, domain.ReviewPublished
}
