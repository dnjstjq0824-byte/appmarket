package application

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/ads/domain"
)

type GSPWinner struct {
	Campaign    domain.AdCampaign
	ClearingCPC int64
	RankScore   float64
}

type AuctionEngine struct {
	mu        sync.RWMutex
	campaigns map[string]*domain.AdCampaign
	floorCPC  int64
}

func NewAuctionEngine(floorCPC int64) *AuctionEngine {
	if floorCPC <= 0 {
		floorCPC = 50 // Minimum ₩50 floor
	}
	return &AuctionEngine{
		campaigns: make(map[string]*domain.AdCampaign),
		floorCPC:  floorCPC,
	}
}

func (e *AuctionEngine) AddCampaign(campaign domain.AdCampaign) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.campaigns[campaign.ID] = &campaign
}

func (e *AuctionEngine) GetCampaign(id string) *domain.AdCampaign {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.campaigns[id]
}

// RunGSPAuction runs Generalized Second-Price auction for a keyword
func (e *AuctionEngine) RunGSPAuction(keyword, deviceHash string, slotCount int) ([]GSPWinner, []domain.ImpressionToken) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	normKw := strings.ToLower(strings.TrimSpace(keyword))
	type Candidate struct {
		Campaign  *domain.AdCampaign
		RankScore float64
	}

	var eligible []Candidate

	for _, camp := range e.campaigns {
		if camp.Status != domain.CampaignActive {
			continue
		}
		if camp.SpentToday >= camp.DailyBudget {
			continue
		}

		matched := false
		for _, kw := range camp.TargetKeywords {
			if strings.EqualFold(strings.TrimSpace(kw), normKw) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		rankScore := float64(camp.MaxBidCPC) * camp.QualityScore
		eligible = append(eligible, Candidate{
			Campaign:  camp,
			RankScore: rankScore,
		})
	}

	// Sort descending by RankScore
	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].RankScore > eligible[j].RankScore
	})

	var winners []GSPWinner
	var tokens []domain.ImpressionToken
	now := time.Now().UTC()

	limit := slotCount
	if limit > len(eligible) {
		limit = len(eligible)
	}

	for i := 0; i < limit; i++ {
		winner := eligible[i]
		clearing := e.floorCPC

		// Generalized Second Price: price determined by next competitor
		if i+1 < len(eligible) {
			nextScore := eligible[i+1].RankScore
			clearing = int64(math.Ceil(nextScore/winner.Campaign.QualityScore)) + 1
		}

		// Cap clearing price at winner's MaxBidCPC
		if clearing > winner.Campaign.MaxBidCPC {
			clearing = winner.Campaign.MaxBidCPC
		}
		if clearing < e.floorCPC {
			clearing = e.floorCPC
		}

		tokenID := fmt.Sprintf("tok-%d-%s-%d", now.UnixNano(), winner.Campaign.ID, i)
		token := domain.ImpressionToken{
			TokenID:     tokenID,
			CampaignID:  winner.Campaign.ID,
			SlotIndex:   i,
			DeviceHash:  deviceHash,
			ClearingCPC: clearing,
			IssuedAt:    now,
			ExpiresAt:   now.Add(5 * time.Minute), // 5-minute impression validity
		}

		winners = append(winners, GSPWinner{
			Campaign:    *winner.Campaign,
			ClearingCPC: clearing,
			RankScore:   winner.RankScore,
		})
		tokens = append(tokens, token)
	}

	return winners, tokens
}
