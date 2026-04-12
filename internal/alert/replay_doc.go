// Package alert – replay module.
//
// # Replay
//
// The Replayer provides a bounded in-memory buffer of recent [Notification]
// values so that they can be re-delivered to a downstream [Dispatcher] after
// a transient outage.
//
// # Usage
//
//	replayer := alert.NewReplayer(alert.DefaultReplayPolicy())
//
//	// Wrap any dispatcher to automatically buffer every notification.
//	rd := alert.NewReplayDispatcher(myDispatcher, replayer)
//
//	// When the downstream recovers, drain the buffer:
//	if err := replayer.Replay(ctx, myDispatcher); err != nil {
//	    log.Printf("some notifications could not be replayed: %v", err)
//	}
//
// # Policy
//
// [ReplayPolicy] controls two dimensions:
//
//   - MaxEvents – maximum number of entries kept in the ring buffer.
//     Oldest entries are evicted when the buffer is full.
//   - MaxAge – notifications older than this duration are silently
//     dropped during [Replayer.Replay] and never re-delivered.
//
// The zero value of [ReplayPolicy] is not valid; use [DefaultReplayPolicy]
// to obtain safe defaults (64 events, 10-minute max age).
package alert
