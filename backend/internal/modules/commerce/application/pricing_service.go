package application

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
)

// PricingService stores per-country list prices for apps.
type PricingService struct {
	mu    sync.RWMutex
	tiers map[string]domain.PriceTier // appID + ":" + country
}

func NewPricingService() *PricingService {
	return &PricingService{tiers: make(map[string]domain.PriceTier)}
}

func priceKey(appID, country string) string { return appID + ":" + strings.ToUpper(country) }

// SetTier creates or replaces one country price.
func (s *PricingService) SetTier(_ context.Context, appID, countryCode string, currency domain.Currency, listPriceMinor int64, taxInclusive bool) (domain.PriceTier, error) {
	if appID == "" || len(strings.TrimSpace(countryCode)) != 2 {
		return domain.PriceTier{}, fmt.Errorf("%w: app and 2-letter country are required", domain.ErrPriceTierInvalid)
	}
	if listPriceMinor < 0 {
		return domain.PriceTier{}, fmt.Errorf("%w: price cannot be negative", domain.ErrPriceTierInvalid)
	}
	if currency == "" {
		currency = domain.CurrencyKRW
	}
	tier := domain.PriceTier{
		AppID:          appID,
		CountryCode:    strings.ToUpper(strings.TrimSpace(countryCode)),
		Currency:       currency,
		ListPriceMinor: listPriceMinor,
		TaxInclusive:   taxInclusive,
		UpdatedAt:      time.Now().UTC(),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tiers[priceKey(appID, countryCode)] = tier
	return tier, nil
}

func (s *PricingService) ListTiers(_ context.Context, appID string) []domain.PriceTier {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.PriceTier
	for _, t := range s.tiers {
		if t.AppID == appID {
			out = append(out, t)
		}
	}
	return out
}
