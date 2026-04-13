package alert

import (
	"context"
	"fmt"
)

// NewCacheDispatcher wraps next with a ResponseCache so that duplicate
// notifications (same port + state) within the TTL are silently dropped.
// Panics if next or cache is nil.
func NewCacheDispatcher(cache *ResponseCache, next Dispatcher) Dispatcher {
	if cache == nil {
		panic("alert: NewCacheDispatcher: cache must not be nil")
	}
	if next == nil {
		panic("alert: NewCacheDispatcher: next must not be nil")
	}
	return &cacheDispatcher{cache: cache, next: next}
}

type cacheDispatcher struct {
	cache *ResponseCache
	next  Dispatcher
}

func (d *cacheDispatcher) Send(ctx context.Context, n Notification) error {
	key := cacheKey(n)
	if d.cache.Hit(key) {
		return nil
	}
	if err := d.next.Send(ctx, n); err != nil {
		return err
	}
	d.cache.Record(key)
	return nil
}

func cacheKey(n Notification) string {
	return fmt.Sprintf("%d:%s", n.Port, n.State)
}
