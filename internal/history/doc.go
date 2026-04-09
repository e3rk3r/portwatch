// Package history provides a thread-safe, fixed-capacity ring buffer for
// recording port state-change events observed by the portwatch daemon.
//
// Usage:
//
//	ring := history.NewRing(200)   // keep last 200 events
//	ring.Record(history.Event{
//		Port:  8080,
//		Host:  "localhost",
//		State: "open",
//	})
//
//	events := ring.Snapshot()       // chronological slice, safe to read
//
// The ring is designed to be embedded in the daemon and queried by future
// status endpoints (e.g. HTTP /status or CLI --history flag) without
// blocking the main poll loop.
package history
