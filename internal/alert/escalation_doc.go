// Package alert provides rate-limiting, retry, backoff, and notification
// dispatch primitives for portwatch alerts.
//
// # Escalation
//
// The escalation subsystem allows portwatch to forward notifications to a
// secondary channel (e.g., PagerDuty webhook or on-call script) when a port
// remains in a degraded state beyond a configurable threshold.
//
// Usage:
//
//	policy := alert.EscalationPolicy{
//		Threshold:   3,
//		MinDuration: 5 * time.Minute,
//	}
//	escCh, _ := alert.NewChannel(alert.ChannelConfig{
//		Type:     "webhook",
//		Endpoint: "https://hooks.example.com/oncall",
//	})
//	ed := alert.NewEscalationDispatcher(primary, escCh, policy)
//	// ed.Send(ctx, notification) — escalates automatically when thresholds are met.
//
// The Escalator resets its counter for a key whenever an "open" (recovered)
// notification is received, preventing stale escalation state.
package alert
