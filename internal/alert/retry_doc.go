// Package alert provides rate limiting, notification dispatching, and
// resilient HTTP delivery for portwatch state-change events.
//
// # Retry
//
// The Retryer type wraps an HTTP client and automatically retries webhook
// deliveries that encounter transient failures (network errors or HTTP 5xx
// responses). Retries use a configurable exponential back-off strategy:
//
//	 policy := alert.DefaultRetryPolicy() // 4 attempts, 500 ms initial wait, ×2 multiplier
//	 retryer := alert.NewRetryer(policy)
//	 resp, err := retryer.Do(ctx, req)
//
// DefaultRetryPolicy() returns a policy suitable for most webhook targets.
// Custom policies can be constructed directly via RetryPolicy{}.
//
// Context cancellation is honoured between retry attempts, so the daemon
// can shut down cleanly without blocking on in-flight retries.
package alert
