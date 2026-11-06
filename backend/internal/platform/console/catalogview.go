package console

import (
	"context"
	"time"

	catalogApp "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
)

// AppDTO is the developer-console app summary.
type AppDTO struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	PackageName  string    `json:"packageName"`
	Status       string    `json:"status"`
	Rating       float64   `json:"rating"`
	Installs     int64     `json:"installs"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// LocalizationDTO is one locale's editable listing fields.
type LocalizationDTO struct {
	ID               string `json:"id"`
	AppID            string `json:"appId"`
	Locale           string `json:"locale"`
	Title            string `json:"title"`
	ShortDescription string `json:"shortDescription"`
	FullDescription  string `json:"fullDescription"`
	WhatsNew         string `json:"whatsNew"`
}

// CatalogViewService serves org-scoped catalog reads/writes for consoles.
type CatalogViewService struct {
	resolver *OrgAppResolver
	catalog  *catalogApp.CatalogService
}

func NewCatalogViewService(resolver *OrgAppResolver, catalog *catalogApp.CatalogService) *CatalogViewService {
	return &CatalogViewService{resolver: resolver, catalog: catalog}
}

func (s *CatalogViewService) ListOrgApps(ctx context.Context, userID, orgID string) ([]AppDTO, error) {
	orgID, _, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, "")
	if err != nil && err != ErrNoApp {
		return nil, err
	}
	apps, err := s.catalog.ListAppsByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make([]AppDTO, 0, len(apps))
	for _, a := range apps {
		out = append(out, AppDTO{
			ID: a.ID, Slug: a.CanonicalSlug, PackageName: a.PrimaryPackageName,
			Status: string(a.CatalogStatus), Rating: a.AverageRating, Installs: a.InstallCount,
			UpdatedAt: a.UpdatedAt,
		})
	}
	return out, nil
}

// SubmitForReview moves an org-owned DRAFT/UNLISTED app into the review queue.
func (s *CatalogViewService) SubmitForReview(ctx context.Context, userID, orgID, appID string) error {
	if _, resolvedApp, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID); err != nil {
		return err
	} else {
		appID = resolvedApp
	}
	return s.catalog.SubmitForReview(ctx, appID)
}

func (s *CatalogViewService) ListLocalizations(ctx context.Context, userID, orgID, appID string) ([]LocalizationDTO, error) {
	if _, resolvedApp, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID); err != nil {
		return nil, err
	} else {
		appID = resolvedApp
	}
	locs, err := s.catalog.ListLocalizations(ctx, appID)
	if err != nil {
		return nil, err
	}
	out := make([]LocalizationDTO, 0, len(locs))
	for _, l := range locs {
		out = append(out, LocalizationDTO{
			ID: l.ID, AppID: l.AppID, Locale: l.Locale, Title: l.Title,
			ShortDescription: l.ShortDescription, FullDescription: l.FullDescription, WhatsNew: l.WhatsNew,
		})
	}
	return out, nil
}
