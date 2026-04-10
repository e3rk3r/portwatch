package alert

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryer_SuccessOnFirstAttempt(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	r := NewRetryer(DefaultRetryPolicy())
	req, _ := http.NewRequest(http.MethodPost, ts.URL, nil)
	resp, err := r.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRetryer_RetriesOn500(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	policy := RetryPolicy{MaxAttempts: 4, InitialWait: 1 * time.Millisecond, Multiplier: 1.0, MaxWait: 5 * time.Millisecond}
	r := NewRetryer(policy)
	req, _ := http.NewRequest(http.MethodPost, ts.URL, nil)
	resp, err := r.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if calls.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", calls.Load())
	}
}

func TestRetryer_ExhaustsAttempts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	policy := RetryPolicy{MaxAttempts: 3, InitialWait: 1 * time.Millisecond, Multiplier: 1.0, MaxWait: 5 * time.Millisecond}
	r := NewRetryer(policy)
	req, _ := http.NewRequest(http.MethodPost, ts.URL, nil)
	_, err := r.Do(context.Background(), req)
	if err == nil {
		t.Fatal("expected error after exhausting attempts")
	}
}

func TestRetryer_RespectsContextCancel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	policy := RetryPolicy{MaxAttempts: 10, InitialWait: 50 * time.Millisecond, Multiplier: 1.0, MaxWait: 100 * time.Millisecond}
	r := NewRetryer(policy)
	req, _ := http.NewRequest(http.MethodPost, ts.URL, nil)

	go func() {
		time.Sleep(60 * time.Millisecond)
		cancel()
	}()

	_, err := r.Do(ctx, req)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
