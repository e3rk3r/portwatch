// Package alert — hedge dispatcher
//
// # Hedge Dispatcher
//
// The hedge dispatcher implements the "hedged request" pattern: it sends the
// primary notification immediately and, if no success is observed within a
// configurable [HedgePolicy.Delay], fires one or more additional concurrent
// attempts against the same [Dispatcher].
//
// The first attempt that returns nil is treated as success and its result is
// returned to the caller. Remaining in-flight goroutines are allowed to finish
// naturally (they share the parent context) so that connections are not
// abandoned mid-flight.
//
// # When to use
//
// Use the hedge dispatcher when your downstream webhook or script has a
// long-tail latency distribution and you want to bound p99 delivery latency
// without sacrificing reliability. It trades a small increase in duplicate
// deliveries (at-least-once semantics) for a significant reduction in tail
// latency.
//
// # Configuration
//
//	policy := alert.HedgePolicy{
//	    Delay:     150 * time.Millisecond,
//	    MaxHedges: 1,
//	}
//	dispatcher := alert.NewHedgeDispatcher(primary, policy)
package alert
