// Package alert implements a simple rate-limiter for portwatch alerts.
//
// When a monitored port flaps rapidly the same webhook or script would
// otherwise be invoked on every poll tick. The Limiter in this package
// enforces a per-key cooldown so that downstream systems are not flooded.
//
// Typical usage:
//
//	limiter := alert.NewLimiter(alert.DefaultPolicy())
//
//	// Inside your event loop:
//	key := fmt.Sprintf("%s:%d:%s", host, port, state)
//	if limiter.Allow(key) {
//	    executor.Run(port, state)
//	}
//
// The cooldown duration is configurable via alert.Policy. Use
// alert.DefaultPolicy() for a sensible 30-second default.
package alert
