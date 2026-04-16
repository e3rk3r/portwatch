// Package alert – load shedding dispatcher.
//
// # Overview
//
// LoadShedder provides back-pressure by capping the number of concurrent
// Dispatch calls that may be in-flight at any one time.  When the limit is
// reached, new notifications are immediately rejected with ErrLoadShed rather
// than queued, preserving system stability under burst traffic.
//
// # Usage
//
//	policy := alert.ShedPolicy{MaxInFlight: 16}
//	d := alert.NewShedDispatcher(policy, downstream)
//
// # Behaviour
//
//   - The first MaxInFlight concurrent calls proceed normally.
//   - Any additional concurrent call receives ErrLoadShed instantly.
//   - Slots are released via defer, so errors in downstream do not leak slots.
//   - Zero or negative MaxInFlight falls back to DefaultShedPolicy (32).
package alert
