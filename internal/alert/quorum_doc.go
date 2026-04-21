// Package alert provides composable dispatcher middleware for portwatch.
//
// # Quorum Dispatcher
//
// NewQuorumDispatcher wraps a slice of Dispatcher targets and sends a
// Notification to all of them concurrently. It returns nil (success) only
// when at least Policy.Required targets respond without error.
//
// This is useful when you have redundant notification channels (e.g. two
// webhook endpoints and a log sink) and want to guarantee that a minimum
// number of them actually received the alert before considering the send
// successful.
//
// Example — require 2 out of 3 targets:
//
//	q := alert.NewQuorumDispatcher(
//		alert.QuorumPolicy{Total: 3, Required: 2},
//		[]alert.Dispatcher{webhookA, webhookB, logSink},
//	)
//
// Use DefaultQuorumPolicy(n) to get a simple-majority policy (⌊n/2⌋ + 1).
package alert
