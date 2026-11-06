package console

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	catalogDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/domain"
	distApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
)

const maxMediaBytes = 20 * 1024 * 1024

var (
	ErrMediaInvalid     = errors.New("ERR_MEDIA_INVALID")
	ErrMediaNotFound    = errors.New("ERR_MEDIA_NOT_FOUND")
	ErrMediaStoreUnset  = errors.New("ERR_MEDIA_STORE_UNSET")
	ErrMediaUnsupported = errors.New("ERR_MEDIA_UNSUPPORTED_FORMAT")
)

// WithMediaStorage attaches the object store used for listing images.
func (s *MetadataService) WithMediaStorage(store distApp.ObjectStorage, bucket string) *MetadataService {
	s.mediaStore = store
	s.mediaBucket = bucket
	return s
}

// SaveMediaImage decodes, hashes, stores, and registers a listing image.
func (s *MetadataService) SaveMediaImage(ctx context.Context, userID, orgID, appID, locale, formFactor, assetType, fileName string, data []byte) (*catalogDomain.MediaAsset, error) {
	_, resolvedApp, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID)
	if err != nil {
		return nil, err
	}
	if s.mediaStore == nil || s.mediaBucket == "" {
		return nil, ErrMediaStoreUnset
	}
	if len(data) == 0 || len(data) > maxMediaBytes {
		return nil, fmt.Errorf("%w: image must be 1..20MiB", ErrMediaInvalid)
	}
	format, dims, err := decodeImageDims(data)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	sha := hex.EncodeToString(sum[:])
	key := fmt.Sprintf("media/%s/%s.%s", resolvedApp, sha, format)
	if err := s.mediaStore.PutObject(ctx, s.mediaBucket, &distApp.StorageObject{
		Key:          key,
		Content:      data,
		SHA256:       sha,
		ContentType:  "image/" + format,
		CacheControl: "public, max-age=31536000, immutable",
	}); err != nil {
		return nil, err
	}
	order := 0
	if existing, err := s.catalog.ListMediaAssets(ctx, resolvedApp, locale); err == nil {
		order = len(existing)
	}
	return s.catalog.SaveMediaAsset(ctx, catalogDomain.MediaAsset{
		AppID: resolvedApp, Locale: locale, DeviceFormFactor: formFactor,
		AssetType: assetType, URL: key,
		Width: dims.X, Height: dims.Y, SHA256: sha, DisplayOrder: order,
	})
}

// ListMedia returns listing images for an org-owned app.
func (s *MetadataService) ListMedia(ctx context.Context, userID, orgID, appID, locale string) ([]catalogDomain.MediaAsset, error) {
	_, resolvedApp, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID)
	if err != nil {
		return nil, err
	}
	return s.catalog.ListMediaAssets(ctx, resolvedApp, locale)
}

// DeleteMedia removes a listing image and its stored bytes.
func (s *MetadataService) DeleteMedia(ctx context.Context, userID, orgID, appID, assetID string) error {
	_, resolvedApp, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID)
	if err != nil {
		return err
	}
	listed, err := s.catalog.ListMediaAssets(ctx, resolvedApp, "")
	if err != nil {
		return err
	}
	var key string
	for _, m := range listed {
		if m.ID == assetID {
			key = m.URL
			break
		}
	}
	if key == "" {
		return ErrMediaNotFound
	}
	if err := s.catalog.DeleteMediaAsset(ctx, resolvedApp, assetID); err != nil {
		return err
	}
	if s.mediaStore != nil && s.mediaBucket != "" {
		_ = s.mediaStore.DeleteObject(ctx, s.mediaBucket, key)
	}
	return nil
}

// MediaContent streams stored image bytes for an org-owned app asset.
func (s *MetadataService) MediaContent(ctx context.Context, userID, orgID, appID, assetID string) (*distApp.StorageObject, error) {
	_, resolvedApp, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID)
	if err != nil {
		return nil, err
	}
	if s.mediaStore == nil || s.mediaBucket == "" {
		return nil, ErrMediaStoreUnset
	}
	listed, err := s.catalog.ListMediaAssets(ctx, resolvedApp, "")
	if err != nil {
		return nil, err
	}
	for _, m := range listed {
		if m.ID == assetID {
			return s.mediaStore.GetObject(ctx, s.mediaBucket, m.URL)
		}
	}
	return nil, ErrMediaNotFound
}

func decodeImageDims(data []byte) (string, image.Point, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", image.Point{}, fmt.Errorf("%w: only PNG and JPEG are accepted", ErrMediaUnsupported)
	}
	if strings.ToLower(format) != "png" && strings.ToLower(format) != "jpeg" {
		return "", image.Point{}, fmt.Errorf("%w: only PNG and JPEG are accepted", ErrMediaUnsupported)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return "", image.Point{}, fmt.Errorf("%w: image has no dimensions", ErrMediaInvalid)
	}
	ext := strings.ToLower(format)
	if ext == "jpeg" {
		ext = "jpg"
	}
	return ext, image.Point{X: cfg.Width, Y: cfg.Height}, nil
}
