package console

import (
	"context"
	"time"

	adsApp "github.com/bookingnetwork/appmarket/backend/internal/modules/ads/application"
	adsDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/ads/domain"
)

// CampaignDTO is the console campaign shape.
type CampaignDTO struct {
	ID             string   `json:"id"`
	AppID          string   `json:"appId"`
	AppName        string   `json:"appName"`
	TargetKeywords []string `json:"targetKeywords"`
	MaxBidCPC      int64    `json:"maxBidCpc"`
	DailyBudget    int64    `json:"dailyBudget"`
	Status         string   `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

func toCampaignDTO(c *adsDomain.AdCampaign) *CampaignDTO {
	return &CampaignDTO{
		ID: c.ID, AppID: c.AppID, AppName: c.AppName, TargetKeywords: c.TargetKeywords,
		MaxBidCPC: c.MaxBidCPC, DailyBudget: c.DailyBudget,
		Status: string(c.Status), CreatedAt: c.CreatedAt,
	}
}

// AdsViewService scopes campaign management to org-owned apps.
type AdsViewService struct {
	resolver  *OrgAppResolver
	campaigns *adsApp.CampaignService
}

func NewAdsViewService(resolver *OrgAppResolver, campaigns *adsApp.CampaignService) *AdsViewService {
	return &AdsViewService{resolver: resolver, campaigns: campaigns}
}

func (s *AdsViewService) CreateCampaign(ctx context.Context, userID, orgID, appID, appName string, keywords []string, maxBidCPC, dailyBudget int64) (*CampaignDTO, error) {
	orgID, appID, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID)
	if err != nil {
		return nil, err
	}
	c, err := s.campaigns.CreateCampaign(ctx, adsApp.CreateCampaignInput{
		OrgID: orgID, AppID: appID, AppName: appName,
		TargetKeywords: keywords, MaxBidCPC: maxBidCPC, DailyBudget: dailyBudget,
	})
	if err != nil {
		return nil, err
	}
	return toCampaignDTO(c), nil
}

func (s *AdsViewService) ListCampaigns(ctx context.Context, userID, orgID string) ([]CampaignDTO, error) {
	orgID, _, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, "")
	if err != nil && err != ErrNoApp {
		return nil, err
	}
	list := s.campaigns.ListCampaignsByOrg(ctx, orgID)
	out := make([]CampaignDTO, 0, len(list))
	for i := range list {
		out = append(out, *toCampaignDTO(&list[i]))
	}
	return out, nil
}

func (s *AdsViewService) SetPaused(ctx context.Context, userID, orgID, campaignID string, paused bool) (*CampaignDTO, error) {
	orgID, _, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, "")
	if err != nil && err != ErrNoApp {
		return nil, err
	}
	c, err := s.campaigns.SetPaused(ctx, orgID, campaignID, paused)
	if err != nil {
		return nil, err
	}
	return toCampaignDTO(c), nil
}
