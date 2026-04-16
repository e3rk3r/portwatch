// Package alert — correlate.go
//
// # Event Correlation
//
// The Correlator groups repeated notifications for the same (port, state) pair
// within a sliding time window. A downstream dispatcher is only invoked once
// the configured MinEvents threshold is reached inside the window, suppressing
// transient flaps and noisy repeated alerts.
//
// # Usage
//
//	d := alert.NewCorrelateDispatcher(
//		alert.CorrelatePolicy{
//			WindowDuration: 2 * time.Minute,
//			MinEvents:      3,
//		},
//		nextDispatcher,
//	)
//
// # Behaviour
//
//   - The first (MinEvents-1) notifications within a window are silently dropped.
//   - The MinEvents-th notification is forwarded exactly once.
//   - Subsequent notifications within the same window are dropped.
//   - After the window expires the counter resets on the next notification.
package alert
