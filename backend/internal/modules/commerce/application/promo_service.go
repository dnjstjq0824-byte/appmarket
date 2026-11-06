package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/sanitize"
)

// PromoService manages promotion campaigns, code batches, and redemption.
type PromoService struct {
	mu         sync.RWMutex
	promotions map[string]*domain.Promotion
	codes      map[string]*domain.PromoCode // upper(code) -> code
	byApp      map[string][]string          // appID -> promotion IDs
	used       map[string]map[string]bool   // codeID -> userID -> redeemed
	usedKeys   map[string]map[string]string // codeID -> userID -> idempotency key
}

func NewPromoService() *PromoService {
	return &PromoService{
		promotions: make(map[string]*domain.Promotion),
		codes:      make(map[string]*domain.PromoCode),
		byApp:      make(map[string][]string),
		used:       make(map[string]map[string]bool),
		usedKeys:   make(map[string]map[string]string),
	}
}

type CreatePromoInput struct {
	AppID         string
	CampaignName  string
	DiscountType  domain.DiscountType
	DiscountValue int
	StartsAt      time.Time
	EndsAt        time.Time
	MaxUses       int
	CodeCount     int
	CodePrefix    string
}

// CreatePromotion validates and stores a campaign with a fresh code batch.
func (s *PromoService) CreatePromotion(_ context.Context, in CreatePromoInput) (*domain.Promotion, []domain.PromoCode, error) {
	if in.AppID == "" || strings.TrimSpace(in.CampaignName) == "" {
		return nil, nil, fmt.Errorf("%w: app and campaign name are required", domain.ErrPromoInvalid)
	}
	switch in.DiscountType {
	case domain.DiscountPercentage:
		if in.DiscountValue <= 0 || in.DiscountValue > 100 {
			return nil, nil, fmt.Errorf("%w: percentage must be 1-100", domain.ErrPromoInvalid)
		}
	case domain.DiscountFixedMinor, domain.DiscountFreeRedeem:
		if in.DiscountValue <= 0 && in.DiscountType != domain.DiscountFreeRedeem {
			return nil, nil, fmt.Errorf("%w: fixed discount must be positive", domain.ErrPromoInvalid)
		}
	default:
		return nil, nil, fmt.Errorf("%w: unknown discount type", domain.ErrPromoInvalid)
	}
	if !in.EndsAt.After(in.StartsAt) {
		return nil, nil, fmt.Errorf("%w: ends_at must be after starts_at", domain.ErrPromoInvalid)
	}
	if in.MaxUses <= 0 {
		return nil, nil, fmt.Errorf("%w: max_uses must be positive", domain.ErrPromoInvalid)
	}
	if in.CodeCount <= 0 || in.CodeCount > 10000 {
		return nil, nil, fmt.Errorf("%w: code_count must be 1-10000", domain.ErrPromoInvalid)
	}
	now := time.Now().UTC()
	promo := &domain.Promotion{
		ID:            fmt.Sprintf("promo-%d", now.UnixNano()),
		AppID:         in.AppID,
		CampaignName:  sanitize.Strict(in.CampaignName),
		DiscountType:  in.DiscountType,
		DiscountValue: in.DiscountValue,
		StartsAt:      in.StartsAt,
		EndsAt:        in.EndsAt,
		MaxUses:       in.MaxUses,
		CreatedAt:     now,
	}
	prefix := strings.ToUpper(strings.TrimSpace(in.CodePrefix))
	codes := make([]domain.PromoCode, 0, in.CodeCount)
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := 0; i < in.CodeCount; i++ {
		code, err := promoCode(prefix)
		if err != nil {
			return nil, nil, err
		}
		for {
			if _, exists := s.codes[code]; !exists {
				break
			}
			code, err = promoCode(prefix)
			if err != nil {
				return nil, nil, err
			}
		}
		c := &domain.PromoCode{
			ID:          fmt.Sprintf("pc-%s-%d", promo.ID, i),
			PromotionID: promo.ID,
			Code:        code,
			MaxUses:     1,
			CreatedAt:   now,
		}
		s.codes[code] = c
		codes = append(codes, *c)
	}
	s.promotions[promo.ID] = promo
	s.byApp[in.AppID] = append(s.byApp[in.AppID], promo.ID)
	cp := *promo
	return &cp, codes, nil
}

func promoCode(prefix string) (string, error) {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	suffix := strings.ToUpper(hex.EncodeToString(b[:]))
	if prefix == "" {
		return suffix, nil
	}
	return prefix + "-" + suffix, nil
}

type Redemption struct {
	Promotion domain.Promotion `json:"promotion"`
	Code      string           `json:"code"`
}

// Redeem validates a code for an app/user and consumes one use.
func (s *PromoService) Redeem(ctx context.Context, code, appID, userID string) (*Redemption, error) {
	return s.RedeemWithKey(ctx, code, appID, userID, "")
}

// Validate runs the same checks as Redeem without consuming a use, so the
// storefront can preview a discount before the order consumes the code.
func (s *PromoService) Validate(ctx context.Context, code, appID, userID string) (*domain.Promotion, error) {
	return s.ValidateWithKey(ctx, code, appID, userID, "")
}

// ValidateWithKey additionally accepts an order retry: when the code was
// already consumed under the same idempotency key, the promotion is returned
// so the order handler can replay instead of failing.
func (s *PromoService) ValidateWithKey(_ context.Context, code, appID, userID, idempotencyKey string) (*domain.Promotion, error) {
	now := time.Now().UTC()
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.codes[strings.ToUpper(strings.TrimSpace(code))]
	if !ok {
		return nil, domain.ErrPromoNotFound
	}
	promo, ok := s.promotions[c.PromotionID]
	if !ok || promo.AppID != appID {
		return nil, domain.ErrPromoNotFound
	}
	if now.Before(promo.StartsAt) || now.After(promo.EndsAt) {
		return nil, domain.ErrPromoExpired
	}
	if s.used[c.ID][userID] {
		if idempotencyKey == "" || s.usedKeys[c.ID][userID] != idempotencyKey {
			return nil, domain.ErrPromoAlreadyUsed
		}
		cp := *promo
		return &cp, nil
	}
	if promo.UsedCount >= promo.MaxUses || c.UsedCount >= c.MaxUses {
		return nil, domain.ErrPromoExhausted
	}
	cp := *promo
	return &cp, nil
}

// RedeemWithKey consumes one use idempotently: a retry carrying the same
// idempotency key returns the original redemption without consuming again.
func (s *PromoService) RedeemWithKey(_ context.Context, code, appID, userID, idempotencyKey string) (*Redemption, error) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.codes[strings.ToUpper(strings.TrimSpace(code))]
	if !ok {
		return nil, domain.ErrPromoNotFound
	}
	promo, ok := s.promotions[c.PromotionID]
	if !ok || promo.AppID != appID {
		return nil, domain.ErrPromoNotFound
	}
	if now.Before(promo.StartsAt) || now.After(promo.EndsAt) {
		return nil, domain.ErrPromoExpired
	}
	if s.used[c.ID] == nil {
		s.used[c.ID] = make(map[string]bool)
	}
	if s.usedKeys[c.ID] == nil {
		s.usedKeys[c.ID] = make(map[string]string)
	}
	if s.used[c.ID][userID] {
		if idempotencyKey != "" && s.usedKeys[c.ID][userID] == idempotencyKey {
			return &Redemption{Promotion: *promo, Code: c.Code}, nil
		}
		return nil, domain.ErrPromoAlreadyUsed
	}
	if promo.UsedCount >= promo.MaxUses || c.UsedCount >= c.MaxUses {
		return nil, domain.ErrPromoExhausted
	}
	s.used[c.ID][userID] = true
	s.usedKeys[c.ID][userID] = idempotencyKey
	c.UsedCount++
	promo.UsedCount++
	return &Redemption{Promotion: *promo, Code: c.Code}, nil
}

// ApplyDiscount returns the payable amount after applying a promotion to a
// list price. It never returns a negative amount.
func ApplyDiscount(listPriceMinor int64, promo *domain.Promotion) int64 {
	if promo == nil {
		return listPriceMinor
	}
	switch promo.DiscountType {
	case domain.DiscountFreeRedeem:
		return 0
	case domain.DiscountPercentage:
		return listPriceMinor * int64(100-promo.DiscountValue) / 100
	case domain.DiscountFixedMinor:
		if int64(promo.DiscountValue) >= listPriceMinor {
			return 0
		}
		return listPriceMinor - int64(promo.DiscountValue)
	default:
		return listPriceMinor
	}
}

func (s *PromoService) ListByApp(_ context.Context, appID string) []domain.Promotion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Promotion
	for _, id := range s.byApp[appID] {
		if p, ok := s.promotions[id]; ok {
			out = append(out, *p)
		}
	}
	return out
}

func (s *PromoService) ListCodes(_ context.Context, promotionID string) []domain.PromoCode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.PromoCode
	for _, c := range s.codes {
		if c.PromotionID == promotionID {
			out = append(out, *c)
		}
	}
	return out
}
