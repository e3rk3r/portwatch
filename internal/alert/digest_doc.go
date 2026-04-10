// Package alert provides alerting primitives for portwatch.
//
// # Digest
//
// The Digester and DigestDispatcher types implement event batching: instead of
// forwarding every state-change notification immediately, they accumulate
// events within a configurable time window and flush them as a single
// (optionally annotated) batch.
//
// This is useful in scenarios where a port flaps rapidly — e.g. during a
// rolling restart — and operators want a summary rather than a flood of
// individual webhook calls.
//
// Usage:
//
//	policy := alert.DigestPolicy{Window: 2 * time.Minute}
//	dd := alert.NewDigestDispatcher(policy, downstream)
//
//	// Accumulate events from the monitor loop:
//	_ = dd.Send(notification)
//
//	// On shutdown, flush any remaining events:
//	dd.Flush()
//
// The DigestDispatcher implements the Dispatcher interface and can be composed
// with SuppressDispatcher or EscalationDispatcher in the middleware chain.
package alert
