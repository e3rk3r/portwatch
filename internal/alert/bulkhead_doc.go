// Package alert provides a composable alerting pipeline for portwatch.
//
// # Bulkhead Dispatcher
//
// The bulkhead pattern limits the number of concurrent in-flight dispatches
// to a downstream Dispatcher. This prevents a slow or overloaded downstream
// (e.g. a webhook endpoint) from exhausting goroutines or file descriptors
// across the entire daemon.
//
// Unlike a queue-based approach, the bulkhead rejects excess calls immediately
// with ErrBulkheadFull, preserving low latency and making back-pressure
// explicit to the caller.
//
// Usage:
//
//	policy := alert.BulkheadPolicy{MaxConcurrent: 4}
//	d := alert.NewBulkheadDispatcher(policy, myWebhookDispatcher)
//
// The bulkhead is safe for concurrent use. It composes naturally with other
// pipeline stages such as Retry, Timeout, and CircuitBreaker.
package alert
