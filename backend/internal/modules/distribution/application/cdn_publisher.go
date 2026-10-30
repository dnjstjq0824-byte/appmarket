package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
)

var (
	ErrCDNImmutablePathConflict = errors.New("ERR_CDN_IMMUTABLE_OVERWRITE_FORBIDDEN")
	ErrObjectNotFound           = errors.New("ERR_OBJECT_NOT_FOUND")
	ErrInvalidChecksumMismatch  = errors.New("ERR_CHECKSUM_MISMATCH")
)

type StorageObject struct {
	Key                string
	Content            []byte
	SHA256             string
	ContentType        string
	CacheControl       string
	ContentDisposition string
	Tags               map[string]string
	CreatedAt          time.Time
}

type ObjectStorage interface {
	GetObject(ctx context.Context, bucket, key string) (*StorageObject, error)
	PutObject(ctx context.Context, bucket string, obj *StorageObject) error
	DeleteObject(ctx context.Context, bucket, key string) error
	TagObject(ctx context.Context, bucket, key string, tags map[string]string) error
}

type InMemoryObjectStorage struct {
	mu      sync.RWMutex
	buckets map[string]map[string]*StorageObject
}

func NewInMemoryObjectStorage() *InMemoryObjectStorage {
	return &InMemoryObjectStorage{
		buckets: make(map[string]map[string]*StorageObject),
	}
}

func (s *InMemoryObjectStorage) GetObject(ctx context.Context, bucket, key string) (*StorageObject, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return nil, ErrObjectNotFound
	}
	obj, exists := b[key]
	if !exists {
		return nil, ErrObjectNotFound
	}
	return obj, nil
}

func (s *InMemoryObjectStorage) PutObject(ctx context.Context, bucket string, obj *StorageObject) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.buckets[bucket]; !exists {
		s.buckets[bucket] = make(map[string]*StorageObject)
	}
	s.buckets[bucket][obj.Key] = obj
	return nil
}

func (s *InMemoryObjectStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if b, exists := s.buckets[bucket]; exists {
		delete(b, key)
	}
	return nil
}

func (s *InMemoryObjectStorage) TagObject(ctx context.Context, bucket, key string, tags map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, exists := s.buckets[bucket]
	if !exists {
		return ErrObjectNotFound
	}
	obj, exists := b[key]
	if !exists {
		return ErrObjectNotFound
	}
	if obj.Tags == nil {
		obj.Tags = make(map[string]string)
	}
	for k, v := range tags {
		obj.Tags[k] = v
	}
	return nil
}

type CDNPublisher struct {
	storage          ObjectStorage
	quarantineBucket string
	publicBucket     string
	cdnOriginDomain  string
	purgedKeys       map[string]time.Time
	mu               sync.Mutex
}

func NewCDNPublisher(storage ObjectStorage, quarantineBucket, publicBucket, cdnOriginDomain string) *CDNPublisher {
	return &CDNPublisher{
		storage:          storage,
		quarantineBucket: quarantineBucket,
		publicBucket:     publicBucket,
		cdnOriginDomain:  cdnOriginDomain,
		purgedKeys:       make(map[string]time.Time),
	}
}

// FormatCDNKey generates content-addressable immutable key: packages/{app_id}/{release_id}/{sha256}.apk
func FormatCDNKey(appID, releaseID, sha256Checksum string) string {
	return fmt.Sprintf("packages/%s/%s/%s.apk", appID, releaseID, sha256Checksum)
}

// PromoteToPublicCDN copies verified artifact from quarantine to public CDN with immutable caching
func (p *CDNPublisher) PromoteToPublicCDN(ctx context.Context, artifact *domain.ReleaseArtifact) (string, error) {
	// 1. Retrieve quarantined binary
	quarantineObj, err := p.storage.GetObject(ctx, p.quarantineBucket, artifact.S3QuarantineKey)
	if err != nil {
		return "", fmt.Errorf("failed to get quarantined artifact: %w", err)
	}

	// 2. Validate SHA-256 integrity
	hasher := sha256.New()
	hasher.Write(quarantineObj.Content)
	actualSHA256 := hex.EncodeToString(hasher.Sum(nil))

	if actualSHA256 != artifact.SHA256Hash {
		return "", ErrInvalidChecksumMismatch
	}

	cdnKey := FormatCDNKey(artifact.PackageName, artifact.ReleaseID, artifact.SHA256Hash)

	// 3. Immutability Check: ensure existing file cannot be overwritten with different binary
	existingObj, err := p.storage.GetObject(ctx, p.publicBucket, cdnKey)
	if err == nil && existingObj != nil {
		if existingObj.SHA256 != actualSHA256 {
			return "", ErrCDNImmutablePathConflict
		}
		// If identical content already exists, idempotent return
		return fmt.Sprintf("https://%s/%s", p.cdnOriginDomain, cdnKey), nil
	}

	// 4. Construct immutable public CDN object
	publicObj := &StorageObject{
		Key:                cdnKey,
		Content:            quarantineObj.Content,
		SHA256:             actualSHA256,
		ContentType:        "application/vnd.android.package-archive",
		CacheControl:       "public, max-age=31536000, immutable",
		ContentDisposition: fmt.Sprintf("attachment; filename=\"%s-%d.apk\"", artifact.PackageName, artifact.VersionCode),
		Tags: map[string]string{
			"status":      "published",
			"app_id":      artifact.PackageName,
			"release_id":  artifact.ReleaseID,
			"promoted_at": time.Now().UTC().Format(time.RFC3339),
		},
		CreatedAt: time.Now().UTC(),
	}

	if err := p.storage.PutObject(ctx, p.publicBucket, publicObj); err != nil {
		return "", fmt.Errorf("failed to put public cdn object: %w", err)
	}

	return fmt.Sprintf("https://%s/%s", p.cdnOriginDomain, cdnKey), nil
}

// PurgeCDNRelease performs edge purge and marks storage object as quarantined_revoked
func (p *CDNPublisher) PurgeCDNRelease(ctx context.Context, appID, releaseID, sha256Checksum string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	cdnKey := FormatCDNKey(appID, releaseID, sha256Checksum)
	p.purgedKeys[cdnKey] = time.Now().UTC()

	// Update object tag to block origin serving
	_ = p.storage.TagObject(ctx, p.publicBucket, cdnKey, map[string]string{
		"status":      "quarantined_revoked",
		"revoked_at":  time.Now().UTC().Format(time.RFC3339),
	})

	return nil
}

func (p *CDNPublisher) IsKeyPurged(cdnKey string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, exists := p.purgedKeys[cdnKey]
	return exists
}
