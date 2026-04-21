// Package alert – tap dispatcher
//
// # Overview
//
// A [Tap] is a passive recording buffer that can be inserted anywhere in a
// dispatcher pipeline without changing its observable behaviour.  Every
// notification that flows through the tap is stored in an in-memory ring
// buffer; the downstream dispatcher is always called and its result is
// returned unchanged.
//
// # Typical use-cases
//
//   - Unit-testing pipeline stages by inspecting what notifications were
//     delivered.
//   - Live debugging: attach a tap in production and query its snapshot via
//     the HTTP API without modifying business logic.
//   - Lightweight audit trail when a full [AuditLog] is not required.
//
// # Usage
//
//	tap := alert.NewTap(nil) // nil → default policy (256 entries)
//	pipeline := alert.NewTapDispatcher(tap, downstream)
//
//	// later…
//	for _, n := range tap.Snapshot() {
//		fmt.Println(n.Port, n.State)
//	}
//
// # Thread safety
//
// All methods on [Tap] are safe for concurrent use.
package alert
