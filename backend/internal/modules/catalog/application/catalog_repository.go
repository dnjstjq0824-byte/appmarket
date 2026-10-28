package application

import (
	"context"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
)

// CatalogRepository abstracts persistence for catalog data.
type CatalogRepository interface {
	ListPublishedApps(ctx context.Context, filter domain.AppListFilter) ([]domain.CatalogApp, string, error)
	GetAppBySlug(ctx context.Context, slug string) (*domain.CatalogApp, error)
	GetLocalizations(ctx context.Context, appID string) ([]domain.Localization, error)
	GetMediaAssets(ctx context.Context, appID, locale string) ([]domain.MediaAsset, error)
	GetDistributions(ctx context.Context, appID string) ([]domain.Distribution, error)
	GetLatestRelease(ctx context.Context, distributionID string) (*domain.Release, error)
	GetCategories(ctx context.Context) ([]domain.Category, error)
	GetTags(ctx context.Context, appID string) ([]domain.Tag, error)
}

// Cache abstracts a key-value cache (Redis) for read-through caching.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttlSeconds int) error
}
