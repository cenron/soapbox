package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type entry struct {
	data      []byte
	expiresAt time.Time
}

func (e entry) isExpired() bool {
	return time.Now().After(e.expiresAt)
}

type memoryCache struct {
	mu      sync.RWMutex
	entries map[string]entry
}

func NewMemoryCache() Cache {
	return &memoryCache{
		entries: make(map[string]entry),
	}
}

func (c *memoryCache) Get(_ context.Context, key string, dest any) error {
	c.mu.RLock()
	e, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists || e.isExpired() {
		if exists {
			c.mu.Lock()
			delete(c.entries, key)
			c.mu.Unlock()
		}
		return ErrCacheMiss
	}

	return json.Unmarshal(e.data, dest)
}

func (c *memoryCache) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = entry{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

func (c *memoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	return nil
}
