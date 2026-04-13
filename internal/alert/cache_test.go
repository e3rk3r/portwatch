package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

func fixedCacheTime(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestResponseCache_MissOnFirstCall(t *testing.T) {
	now := time.Now()
	c := NewResponseCache(DefaultCachePolicy(), fixedCacheTime(now))
	if c.Hit("8080:open") {
		t.Fatal("expected miss on first call")
	}
}

func TestResponseCache_HitAfterRecord(t *testing.T) {
	now := time.Now()
	c := NewResponseCache(DefaultCachePolicy(), fixedCacheTime(now))
	c.Record("8080:open")
	if !c.Hit("8080:open") {
		t.Fatal("expected hit after record")
	}
}

func TestResponseCache_ExpiredEntry(t *testing.T) {
	now := time.Now()
	p := CachePolicy{TTL: 5 * time.Second, Capacity: 16}
	c := NewResponseCache(p, fixedCacheTime(now))
	c.Record("9090:closed")
	// Advance clock past TTL.
	c.clock = fixedCacheTime(now.Add(10 * time.Second))
	if c.Hit("9090:closed") {
		t.Fatal("expected miss after TTL expiry")
	}
}

func TestResponseCache_Reset(t *testing.T) {
	now := time.Now()
	c := NewResponseCache(DefaultCachePolicy(), fixedCacheTime(now))
	c.Record("1234:open")
	c.Reset()
	if c.Hit("1234:open") {
		t.Fatal("expected miss after reset")
	}
}

func TestCacheDispatcher_PanicsOnNilCache(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil cache")
		}
	}()
	NewCacheDispatcher(nil, &logDispatcher{})
}

func TestCacheDispatcher_DeduplicatesWithinTTL(t *testing.T) {
	now := time.Now()
	cache := NewResponseCache(CachePolicy{TTL: 30 * time.Second, Capacity: 8}, fixedCacheTime(now))
	calls := 0
	next := DispatcherFunc(func(_ context.Context, _ Notification) error {
		calls++
		return nil
	})
	d := NewCacheDispatcher(cache, next)
	n := Notification{Port: 8080, State: "open"}
	for i := 0; i < 3; i++ {
		if err := d.Send(context.Background(), n); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if calls != 1 {
		t.Fatalf("expected 1 downstream call, got %d", calls)
	}
}

func TestCacheDispatcher_PropagatesError(t *testing.T) {
	now := time.Now()
	cache := NewResponseCache(DefaultCachePolicy(), fixedCacheTime(now))
	expected := errors.New("downstream failure")
	next := DispatcherFunc(func(_ context.Context, _ Notification) error { return expected })
	d := NewCacheDispatcher(cache, next)
	err := d.Send(context.Background(), Notification{Port: 443, State: "closed"})
	if !errors.Is(err, expected) {
		t.Fatalf("expected downstream error, got %v", err)
	}
	// Error must NOT be cached — next call should still reach downstream.
	if cache.Hit("443:closed") {
		t.Fatal("failed dispatch must not populate cache")
	}
}
