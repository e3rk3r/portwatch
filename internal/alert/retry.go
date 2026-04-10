package alert

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// RetryPolicy defines how webhook delivery retries behave.
type RetryPolicy struct {
	MaxAttempts int
	InitialWait time.Duration
	Multiplier  float64
	MaxWait     time.Duration
}

// DefaultRetryPolicy returns a sensible retry policy for webhook delivery.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 4,
		InitialWait: 500 * time.Millisecond,
		Multiplier:  2.0,
		MaxWait:     10 * time.Second,
	}
}

// Retryer wraps an HTTP client and retries transient failures.
type Retryer struct {
	client *http.Client
	policy RetryPolicy
	cfunc() time.Time
	sleep  func(time.Duration)
}

// NewRetryer creates a Retryer with the given policy and a real HTTP client.
func NewRetryer(policy RetryPolicy) *Retryer {
	return &Retryer{
		client: &http.Client{Timeout: 10 * time.Second},
		policy: policy,
		clock:  time.Now,
		sleep:  time.Sleep,
	}
}

// Do performs the request, retrying on transient (5xx / network) errors.
// It respects context cancellation between attempts.
func (r *Retryer) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	wait := r.policy.InitialWait
	var lastErr error

	for attempt := 1; attempt <= r.policy.MaxAttempts; attempt++ {
		resp, err := r.client.Do(req.WithContext(ctx))
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}
		if err == nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("server returned %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		if attempt == r.policy.MaxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}

		wait = time.Duration(float64(wait) * r.policy.Multiplier)
		if wait > r.policy.MaxWait {
			wait = r.policy.MaxWait
		}
	}

	return nil, fmt.Errorf("all %d attempts failed: %w", r.policy.MaxAttempts, errors.Unwrap(lastErr))
}
