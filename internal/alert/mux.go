package alert

// Mux is a content-based multiplexer that routes a Notification to exactly one
// downstream Dispatcher by evaluating a prioritised list of predicates.
//
// Unlike Router (which can match many routes) Mux stops at the first matching
// branch, making it suitable for mutually-exclusive routing logic such as
// severity tiers or environment splits.
//
// If no branch matches and a default Dispatcher has been registered it is used;
// otherwise Send returns nil without forwarding the notification.
type Mux struct {
	branches []muxBranch
	defaultD Dispatcher
}

type muxBranch struct {
	predicate func(Notification) bool
	target    Dispatcher
}

// DefaultMuxPolicy returns a zero-value Mux ready for branch registration.
func DefaultMuxPolicy() *Mux {
	return &Mux{}
}

// NewMux constructs a Mux with an optional default Dispatcher that is used
// when no registered branch matches the incoming Notification.
// Pass nil to silently drop unmatched notifications.
func NewMux(defaultDispatcher Dispatcher) *Mux {
	return &Mux{defaultD: defaultDispatcher}
}

// Handle registers a branch: when predicate(n) returns true the notification
// is forwarded to target. Branches are evaluated in registration order.
// Panics if predicate or target is nil.
func (m *Mux) Handle(predicate func(Notification) bool, target Dispatcher) {
	if predicate == nil {
		panic("alert: Mux.Handle called with nil predicate")
	}
	if target == nil {
		panic("alert: Mux.Handle called with nil target")
	}
	m.branches = append(m.branches, muxBranch{predicate: predicate, target: target})
}

// Send evaluates each branch predicate in order and forwards n to the first
// matching Dispatcher. If no branch matches, the default Dispatcher is used
// (if set). Returns nil when the notification is intentionally dropped.
func (m *Mux) Send(n Notification) error {
	for _, b := range m.branches {
		if b.predicate(n) {
			return b.target.Send(n)
		}
	}
	if m.defaultD != nil {
		return m.defaultD.Send(n)
	}
	return nil
}

// Len returns the number of registered branches, excluding the default.
func (m *Mux) Len() int {
	return len(m.branches)
}
