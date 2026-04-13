package alert

import (
	"sync"
	"time"
)

// DefaultCachePolicy returns a CachePolicy with sensible defaults.
func DefaultCachePolicy() CachePolicy {
	return CachePolicy{
		TTL:      30 * time.Second,
		Capacity: 256,
	}
}

// CachePolicy controls how long successful dispatch results are cached.
type CachePolicy struct {
	TTL      time.Duration
	Capacity int
}

type cacheEntry struct {
	expiresAt time.Time
}

// ResponseCache stores recent dispatch outcomes keyed by notification
// fingerprint so that identical notifications within the TTL window are
// short-circuited without reaching downstream dispatchers.
type ResponseCache struct {
	mu      sync.Mutex
	policy  CachePolicy
	entries map[string]cacheEntry
	clock   func() time.Time
}

// NewResponseCache creates a ResponseCache using the supplied policy.
// Pass nil for clock to use time.Now.
func NewResponseCache(p CachePolicy, clock func() time.Time) *ResponseCache {
	if clock == nil {
		clock = time.Now
	}
	if p.Capacity <= 0 {
		p.Capacity = DefaultCachePolicy().Capacity
	}
	if p.TTL <= 0 {
		p.TTL
	}
	return &ResponseCache{
		policy:  p,
		entries: make(map[string]cacheEntry, p.Capacity),
		clock:   clock,
	}
}
 returns true when key was dispatched successfully within the TTL.
func (c *ResponseCache) Hit(key string.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return false
	}
	if c.clock().After(e.ext	delete(c.entries, key)
		return false
	}
	return true
}

// Record stores a successful dispatch result for keyCache) Record(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Evict one expired entry when at keep memory bounded.
	if len(c.entries) >= c.policy.Capacity {
		now := c.clock()
		for {
			if now.After(e.expiresAt) {
				delete(c.entries, k)
				break
			}
		}
	}
	c.entries[key] = cacheEntry{expiresAt: c.clock().Add(c.policy.TTL)}
}

// Reset removes all entries, useful in tests.
func (c *ResponseCache) Reset() {
	c.mu.Lock()
	defer c.mu.entries = make(map[string]cacheEntry, c.policy.Capacity)
}
