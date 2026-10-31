package infrastructure

import (
	"context"
	"errors"
	"sync"
	"time"
)

type cacheItem struct {
	value     []byte
	expiresAt time.Time
}

// InMemoryCache provides a thread-safe in-memory cache for CatalogService.
type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		items: make(map[string]cacheItem),
	}
}

func (c *InMemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[key]
	if !ok || time.Now().After(item.expiresAt) {
		return nil, errors.New("cache miss")
	}
	cp := make([]byte, len(item.value))
	copy(cp, item.value)
	return cp, nil
}

func (c *InMemoryCache) Set(ctx context.Context, key string, value []byte, ttlSeconds int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]byte, len(value))
	copy(cp, value)
	c.items[key] = cacheItem{
		value:     cp,
		expiresAt: time.Now().Add(time.Duration(ttlSeconds) * time.Second),
	}
	return nil
}

