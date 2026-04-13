// Package alert provides composable notification dispatchers for portwatch.
//
// # Sliding-Window Rate Limiter
//
// WindowCounter implements a sliding-window event counter that tracks the
// number of events recorded for a given key within a configurable time window.
// Unlike a fixed-window (token-bucket) approach, the sliding window avoids
// burst spikes at window boundaries by pruning only timestamps that have
// fallen outside the rolling duration.
//
// # Usage
//
//	policy := alert.WindowPolicy{Size: 30 * time.Second, MaxEvents: 5}
//	counter := alert.NewWindowCounter(policy)
//
//	// Wrap any Dispatcher:
//	d := alert.NewWindowDispatcher(downstream, counter)
//
// Each call to Send derives a key from the notification's port and state.
// If the number of events recorded for that key within the window is at or
// above MaxEvents, Send returns an error and the notification is dropped.
// Once timestamps age out of the window, new events are permitted again.
//
// Reset(key) can be called to clear the history for a specific key, for
// example when a port transitions back to an open state and the operator
// wants to restart rate-limiting from zero.
package alert
