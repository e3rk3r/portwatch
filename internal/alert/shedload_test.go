package alert

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestLoadShedder_AllowsWithinLimit(t *testing.T) {
	s := NewLoadShedder(ShedPolicy{MaxInFlight: 2})
	if err := s.Acquire(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.InFlight() != 1 {
		t.Fatalf("expected 1 in-flight, got %d", s.InFlight())
	}
	s.Release()
	if s.InFlight() != 0 {
		t.Fatalf("expected 0 in-flight after release")
	}
}

func TestLoadShedder_RejectsAtLimit(t *testing.T) {
	s := NewLoadShedder(ShedPolicy{MaxInFlight: 1})
	if err := s.Acquire(); err != nil {
		t.Fatal(err)
	}
	defer s.Release()
	if err := s.Acquire(); err != ErrLoadShed {
		t.Fatalf("expected ErrLoadShed, got %v", err)
	}
	if s.InFlight() != 1 {
		t.Fatalf("in-flight should remain 1 after shed")
	}
}

func TestLoadShedder_DefaultPolicy(t *testing.T) {
	s := NewLoadShedder(ShedPolicy{MaxInFlight: 0})
	if s.policy.MaxInFlight != DefaultShedPolicy().MaxInFlight {
		t.Fatalf("expected default max in-flight")
	}
}

func TestShedDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewShedDispatcher(DefaultShedPolicy(), nil)
}

func TestShedDispatcher_ForwardsWhenCapacityAvailable(t *testing.T) {
	called := false
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		called = true
		return nil
	})
	d := NewShedDispatcher(ShedPolicy{MaxInFlight: 4}, next)
	if err := d.Dispatch(context.Background(), Notification{Port: 8080}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("next not called")
	}
}

func TestShedDispatcher_ConcurrentLimit(t *testing.T) {
	const limit = 5
	var active int64
	var mu sync.Mutex
	var shed int

	block := make(chan struct{})
	next := dispatcherFunc(func(_ context.Context, _ Notification) error {
		<-block
		return nil
	})
	d := NewShedDispatcher(ShedPolicy{MaxInFlight: limit}, next)

	var wg sync.WaitGroup
	for i := 0; i < limit+3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := d.Dispatch(context.Background(), Notification{Port: 9000})
			if err == ErrLoadShed {
				mu.Lock()
				shed++
				mu.Unlock()
			} else {
				_ = active
			}
		}()
	}
	time.Sleep(20 * time.Millisecond)
	close(block)
	wg.Wait()
	if shed < 1 {
		t.Fatalf("expected at least 1 shed, got %d", shed)
	}
}
