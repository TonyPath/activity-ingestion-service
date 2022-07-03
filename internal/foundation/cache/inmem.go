// Package cache represents a hashset(redis-like) in-memory storage.
package cache

import (
	"context"
	"sync"
	"time"
)

type ResultStatus int

const (
	Found ResultStatus = iota
	Created
)

// InMemory holds required configuration
type InMemory struct {
	mu        sync.RWMutex
	store     map[string]int64
	ttl       int64
	ctx       context.Context
	ctxCancel context.CancelFunc
}

// New initiates the cache storage and setup a periodic cleanup mechanism
func New(ttl int64) *InMemory {
	c := InMemory{
		store: make(map[string]int64),
		ttl:   ttl,
	}
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())

	c.setupCleanup(ttl)
	return &c
}

// FindOrAdd can be used to decide if a message will be ignored in case of Found result
// or proceed to processing in case of Created result.
func (c *InMemory) FindOrAdd(kid string) ResultStatus {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, found := c.store[kid]; found {
		return Found
	}

	c.store[kid] = time.Now().Unix()
	return Created
}

// KeyExists checks if an entry already exists based on it's key identifier.
func (c *InMemory) KeyExists(kid string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, found := c.store[kid]
	return found
}

// Close gracefully by cleaning up the storage.
func (c *InMemory) Close() {
	c.ctxCancel()
}

func (c *InMemory) setupCleanup(ttl int64) {
	ticker := time.NewTicker(time.Duration(ttl) * time.Second)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				c.cleanup()
			}
		}
	}(c.ctx)
}

func (c *InMemory) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.store {
		if v < time.Now().Unix()-c.ttl {
			delete(c.store, k)
		}
	}
}
