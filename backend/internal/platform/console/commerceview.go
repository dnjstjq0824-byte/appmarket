package console

import (
	"context"
	"time"

	commerceApp "github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	commerceDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
)

// PromotionDTO is the console promotion shape with live code count.
type PromotionDTO struct {
	ID            string `json:"id"`
	AppID         string `json:"appId"`
	CampaignName  string `json:"campaignName"`
	DiscountType  string `json:"discountType"`
	DiscountValue int    `json:"discountValue"`
	StartsAt      time.Time `json:"startsAt"`
	EndsAt        time.Time `json:"endsAt"`
	MaxUses       int    `json:"maxUses"`
	UsedCount     int    `json:"usedCount"`
	CodeCount     int    `json:"codeCount"`
}

type PromoCodeDTO struct {
	Code      string `json:"code"`
	UsedCount int    `json:"usedCount"`
}

type PriceTierDTO struct {
	CountryCode    string `json:"countryCode"`
	Currency       string `json:"currency"`
	ListPriceMinor int64  `json:"listPriceMinor"`
	TaxInclusive   bool   `json:"taxInclusive"`
}

type OrderDTO struct {
	ID               string `json:"id"`
	UserID           string `json:"userId"`
	AppID            string `json:"appId"`
	Currency         string `json:"currency"`
	GrossAmount      int64  `json:"grossAmount"`
	CommissionAmount int64  `json:"commissionAmount"`
	DevPayoutAmount  int64  `json:"devPayoutAmount"`
	Status           string `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
}

// CommerceViewService scopes promos, pricing, and settlement reads to
// org-owned apps.
type CommerceViewService struct {
	resolver *OrgAppResolver
	promos   *commerceApp.PromoService
	pricing  *commerceApp.PricingService
	orders   *commerceApp.OrderService
}

func NewCommerceViewService(resolver *OrgAppResolver, promos *commerceApp.PromoService, pricing *commerceApp.PricingService, orders *commerceApp.OrderService) *CommerceViewService {
	return &CommerceViewService{resolver: resolver, promos: promos, pricing: pricing, orders: orders}
}

type CreatePromoArgs struct {
	CampaignName  string
	DiscountType  string
	DiscountValue int
	StartsAt      time.Time
	EndsAt        time.Time
	MaxUses       int
	CodeCount     int
	CodePrefix    string
}

func (s *CommerceViewService) CreatePromotion(ctx context.Context, userID, orgID, appID string, args CreatePromoArgs) (*PromotionDTO, []PromoCodeDTO, error) {
	_, appID, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID)
	if err != nil {
		return nil, nil, err
	}
	promo, codes, err := s.promos.CreatePromotion(ctx, commerceApp.CreatePromoInput{
		AppID: appID, CampaignName: args.CampaignName,
		DiscountType: commerceDomain.DiscountType(args.DiscountType),
		DiscountValue: args.DiscountValue, StartsAt: args.StartsAt, EndsAt: args.EndsAt,
		MaxUses: args.MaxUses, CodeCount: args.CodeCount, CodePrefix: args.CodePrefix,
	})
	if err != nil {
		return nil, nil, err
	}
	out := make([]PromoCodeDTO, 0, len(codes))
	for _, c := range codes {
		out = append(out, PromoCodeDTO{Code: c.Code, UsedCount: c.UsedCount})
	}
	return s.toPromotionDTO(ctx, promo), out, nil
}

func (s *CommerceViewService) ListPromotions(ctx context.Context, userID, orgID, appID string) ([]PromotionDTO, error) {
	_, appID, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID)
	if err != nil {
		return nil, err
	}
	list := s.promos.ListByApp(ctx, appID)
	out := make([]PromotionDTO, 0, len(list))
	for i := range list {
		out = append(out, *s.toPromotionDTO(ctx, &list[i]))
	}
	return out, nil
}

func (s *CommerceViewService) toPromotionDTO(ctx context.Context, p *commerceDomain.Promotion) *PromotionDTO {
	return &PromotionDTO{
		ID: p.ID, AppID: p.AppID, CampaignName: p.CampaignName,
		DiscountType: string(p.DiscountType), DiscountValue: p.DiscountValue,
		StartsAt: p.StartsAt, EndsAt: p.EndsAt,
		MaxUses: p.MaxUses, UsedCount: p.UsedCount,
		CodeCount: len(s.promos.ListCodes(ctx, p.ID)),
	}
}

func (s *CommerceViewService) SetPriceTier(ctx context.Context, userID, orgID, appID, country, currency string, priceMinor int64, taxInclusive bool) (*PriceTierDTO, error) {
	_, appID, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID)
	if err != nil {
		return nil, err
	}
	tier, err := s.pricing.SetTier(ctx, appID, country, commerceDomain.Currency(currency), priceMinor, taxInclusive)
	if err != nil {
		return nil, err
	}
	return &PriceTierDTO{CountryCode: tier.CountryCode, Currency: string(tier.Currency), ListPriceMinor: tier.ListPriceMinor, TaxInclusive: tier.TaxInclusive}, nil
}

func (s *CommerceViewService) ListPriceTiers(ctx context.Context, userID, orgID, appID string) ([]PriceTierDTO, error) {
	_, appID, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID)
	if err != nil {
		return nil, err
	}
	tiers := s.pricing.ListTiers(ctx, appID)
	out := make([]PriceTierDTO, 0, len(tiers))
	for _, t := range tiers {
		out = append(out, PriceTierDTO{CountryCode: t.CountryCode, Currency: string(t.Currency), ListPriceMinor: t.ListPriceMinor, TaxInclusive: t.TaxInclusive})
	}
	return out, nil
}

// OrgSettlement summarizes dues plus the underlying orders for one org.
type OrgSettlement struct {
	Orders      []OrderDTO `json:"orders"`
	GrossTotal  int64      `json:"grossTotal"`
	PayoutTotal int64      `json:"payoutTotal"`
}

func (s *CommerceViewService) OrgSettlement(ctx context.Context, userID, orgID string) (*OrgSettlement, error) {
	orgID, _, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, "")
	if err != nil && err != ErrNoApp {
		return nil, err
	}
	orders := s.orders.ListOrdersByOrg(orgID)
	out := &OrgSettlement{Orders: make([]OrderDTO, 0, len(orders))}
	for _, o := range orders {
		out.Orders = append(out.Orders, OrderDTO{
			ID: o.ID, UserID: o.UserID, AppID: o.AppID, Currency: string(o.Currency),
			GrossAmount: o.GrossAmount, CommissionAmount: o.CommissionAmount,
			DevPayoutAmount: o.DevPayoutAmount, Status: string(o.Status), CreatedAt: o.CreatedAt,
		})
		if o.Status == commerceDomain.OrderPaid {
			out.GrossTotal += o.GrossAmount
			out.PayoutTotal += o.DevPayoutAmount
		}
	}
	return out, nil
}
