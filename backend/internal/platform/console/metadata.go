package console

import (
	"context"

	catalogApp "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	catalogDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
	distApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
)

// MetadataService writes store-listing metadata after verifying the caller
// org owns the app.
type MetadataService struct {
	resolver    *OrgAppResolver
	catalog     *catalogApp.CatalogService
	mediaStore  distApp.ObjectStorage
	mediaBucket string
}

func NewMetadataService(resolver *OrgAppResolver, catalog *catalogApp.CatalogService) *MetadataService {
	return &MetadataService{resolver: resolver, catalog: catalog}
}

// SaveMetadata persists one locale's listing fields for an org-owned app.
func (s *MetadataService) SaveMetadata(ctx context.Context, userID, orgID, appID, locale, title, shortDescription, fullDescription, whatsNew string) (*catalogDomain.Localization, error) {
	_, resolvedApp, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID)
	if err != nil {
		return nil, err
	}
	appID = resolvedApp
	return s.catalog.SaveLocalization(ctx, catalogDomain.Localization{
		AppID:            appID,
		Locale:           locale,
		Title:            title,
		ShortDescription: shortDescription,
		FullDescription:  fullDescription,
		WhatsNew:         whatsNew,
	})
}
