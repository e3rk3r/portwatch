package alert

import (
	"context"
	"fmt"
)

// MuxDispatcher routes a notification to one of several named dispatchers
// based on the result of the Mux selector. It panics on construction if any
// required argument is nil.
type MuxDispatcher struct {
	mux     *Mux
	routes  map[string]Dispatcher
	fallback Dispatcher // optional; if nil, unmatched notifications are dropped
}

// NewMuxDispatcher creates a MuxDispatcher.
//
//   - mux       – the Mux used to select a route key from the notification.
//   - routes    – map of key → Dispatcher; must be non-nil and non-empty.
//   - fallback  – dispatcher used when no route key matches; may be nil.
func NewMuxDispatcher(mux *Mux, routes map[string]Dispatcher, fallback Dispatcher) *MuxDispatcher {
	if mux == nil {
		panic("alert: NewMuxDispatcher: mux must not be nil")
	}
	if len(routes) == 0 {
		panic("alert: NewMuxDispatcher: routes must not be empty")
	}
	copy := make(map[string]Dispatcher, len(routes))
	for k, v := range routes {
		if v == nil {
			panic(fmt.Sprintf("alert: NewMuxDispatcher: route %q has nil dispatcher", k))
		}
		copy[k] = v
	}
	return &MuxDispatcher{mux: mux, routes: copy, fallback: fallback}
}

// Send selects a route key via the Mux, dispatches to the matching Dispatcher,
// and falls back to the fallback Dispatcher (if set) when no key matches.
func (d *MuxDispatcher) Send(ctx context.Context, n Notification) error {
	key := muxKey(n)
	if target, ok := d.routes[key]; ok {
		return target.Send(ctx, n)
	}
	if d.fallback != nil {
		return d.fallback.Send(ctx, n)
	}
	return nil
}

func muxKey(n Notification) string {
	return fmt.Sprintf("%d:%s", n.Port, n.State)
}
