package alert

import "context"

// ReplayDispatcher wraps a downstream Dispatcher and records every
// successfully sent notification in a Replayer buffer. On failure it stores
// the notification so it can be retried via Replayer.Replay.
type ReplayDispatcher struct {
	next     Dispatcher
	replayer *Replayer
}

// NewReplayDispatcher creates a ReplayDispatcher.
// next must not be nil; replayer must not be nil.
func NewReplayDispatcher(next Dispatcher, replayer *Replayer) *ReplayDispatcher {
	if next == nil {
		panic("alert: NewReplayDispatcher: next Dispatcher must not be nil")
	}
	if replayer == nil {
		panic("alert: NewReplayDispatcher: Replayer must not be nil")
	}
	return &ReplayDispatcher{next: next, replayer: replayer}
}

// Send forwards the notification to the next Dispatcher. On success the
// notification is recorded in the replay buffer. On failure the notification
// is also stored so Replay can retry it later.
func (d *ReplayDispatcher) Send(ctx context.Context, n Notification) error {
	err := d.next.Send(ctx, n)
	if err != nil {
		// Store for later replay.
		d.replayer.Record(n)
		return err
	}
	// Record successful deliveries so they can be replayed to a secondary
	// target if needed (e.g. audit or shadow).
	d.replayer.Record(n)
	return nil
}
