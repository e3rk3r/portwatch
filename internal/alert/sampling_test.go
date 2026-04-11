package alert

import (
	"errors"
	"testing"
)

// captureDispatcher records whether Send was called.
type captureDispatcher struct {
	called bool
	err    error
}

func (c *captureDispatcher) Send(_ Notification) error {
	c.called = true
	return c.err
}

func sampleNotif() Notification {
	return BuildNotification(8080, "open", "test")
}

func TestSampler_RateOne_AlwaysPasses(t *testing.T) {
	s := NewSampler(DefaultSamplePolicy())
	for i := 0; i < 20; i++ {
		if !s.Allow(sampleNotif()) {
			t.Fatal("expected Allow=true for rate 1.0")
		}
	}
}

func TestSampler_RateZero_AlwaysBlocks(t *testing.T) {
	s := NewSampler(SamplePolicy{Rate: 0.0})
	for i := 0; i < 20; i++ {
		if s.Allow(sampleNotif()) {
			t.Fatal("expected Allow=false for rate 0.0")
		}
	}
}

func TestSampler_Clamp_NegativeRate(t *testing.T) {
	s := NewSampler(SamplePolicy{Rate: -5.0})
	if s.policy.Rate != 0 {
		t.Fatalf("expected clamped rate 0, got %v", s.policy.Rate)
	}
}

func TestSampler_Clamp_OverRate(t *testing.T) {
	s := NewSampler(SamplePolicy{Rate: 3.5})
	if s.policy.Rate != 1.0 {
		t.Fatalf("expected clamped rate 1.0, got %v", s.policy.Rate)
	}
}

func TestSampleDispatcher_Passes_WhenAllowed(t *testing.T) {
	cap := &captureDispatcher{}
	sd := NewSampleDispatcher(NewSampler(DefaultSamplePolicy()), cap)
	if err := sd.Send(sampleNotif()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cap.called {
		t.Fatal("expected next dispatcher to be called")
	}
}

func TestSampleDispatcher_Drops_WhenBlocked(t *testing.T) {
	cap := &captureDispatcher{}
	sd := NewSampleDispatcher(NewSampler(SamplePolicy{Rate: 0.0}), cap)
	if err := sd.Send(sampleNotif()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.called {
		t.Fatal("expected next dispatcher NOT to be called")
	}
}

func TestSampleDispatcher_PropagatesError(t *testing.T) {
	sentinel := errors.New("downstream error")
	cap := &captureDispatcher{err: sentinel}
	sd := NewSampleDispatcher(NewSampler(DefaultSamplePolicy()), cap)
	if err := sd.Send(sampleNotif()); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestNewSampleDispatcher_NilSamplerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil sampler")
		}
	}()
	NewSampleDispatcher(nil, &captureDispatcher{})
}

func TestNewSampleDispatcher_NilNextPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next")
		}
	}()
	NewSampleDispatcher(NewSampler(DefaultSamplePolicy()), nil)
}
