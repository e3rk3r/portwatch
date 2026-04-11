package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

func fixedCircuitClock(t time.Time) circuitClock { return func() time.Time { return t } }

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitPolicy())
	if cb.State() != CircuitClosed {
		t.Fatalf("expected Closed, got %v", cb.State())
	}
	if !cb.Allow() {
		t.Fatal("expected Allow() true on closed circuit")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	p := CircuitPolicy{FailureThreshold: 2, SuccessThreshold: 1, OpenDuration: time.Minute}
	cb := NewCircuitBreaker(p)
	cb.RecordFailure()
	if cb.State() != CircuitClosed {
		t.Fatal("should still be closed after 1 failure")
	}
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("expected Open after threshold, got %v", cb.State())
	}
	if cb.Allow() {
		t.Fatal("expected Allow() false when open")
	}
}

func TestCircuitBreaker_HalfOpenAfterDuration(t *testing.T) {
	base := time.Now()
	p := CircuitPolicy{FailureThreshold: 1, SuccessThreshold: 1, OpenDuration: 10 * time.Second}
	cb := NewCircuitBreaker(p)
	cb.now = fixedCircuitClock(base)
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatal("expected Open")
	}
	cb.now = fixedCircuitClock(base.Add(11 * time.Second))
	if !cb.Allow() {
		t.Fatal("expected Allow() true after open duration")
	}
	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected HalfOpen, got %v", cb.State())
	}
}

func TestCircuitBreaker_ClosesAfterSuccessThreshold(t *testing.T) {
	p := CircuitPolicy{FailureThreshold: 1, SuccessThreshold: 2, OpenDuration: time.Millisecond}
	cb := NewCircuitBreaker(p)
	cb.RecordFailure()
	time.Sleep(2 * time.Millisecond)
	cb.Allow() // transitions to half-open
	cb.RecordSuccess()
	if cb.State() != CircuitHalfOpen {
		t.Fatal("still half-open after 1 success")
	}
	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Fatalf("expected Closed, got %v", cb.State())
	}
}

func TestCircuitDispatcher_DropsWhenOpen(t *testing.T) {
	p := CircuitPolicy{FailureThreshold: 1, SuccessThreshold: 1, OpenDuration: time.Minute}
	var called bool
	next := DispatcherFunc(func(_ context.Context, _ Notification) error {
		called = true
		return nil
	})
	cd := NewCircuitDispatcher(next, p)
	// force open by recording a failure directly
	cd.breaker.RecordFailure()
	err := cd.Send(context.Background(), Notification{Port: 8080})
	if err == nil {
		t.Fatal("expected error when circuit open")
	}
	if called {
		t.Fatal("downstream should not be called when circuit open")
	}
}

func TestCircuitDispatcher_RecordsFailure(t *testing.T) {
	p := CircuitPolicy{FailureThreshold: 2, SuccessThreshold: 1, OpenDuration: time.Minute}
	next := DispatcherFunc(func(_ context.Context, _ Notification) error {
		return errors.New("downstream error")
	})
	cd := NewCircuitDispatcher(next, p)
	cd.Send(context.Background(), Notification{Port: 9090}) //nolint:errcheck
	if cd.breaker.failures != 1 {
		t.Fatalf("expected 1 failure recorded, got %d", cd.breaker.failures)
	}
}

func TestCircuitDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewCircuitDispatcher(nil, DefaultCircuitPolicy())
}
