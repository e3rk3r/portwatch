package alert

// FilterPolicy defines criteria for deciding whether a notification
// should be forwarded to the next dispatcher in the pipeline.
type FilterPolicy struct {
	// Ports is an optional allowlist of port numbers. If non-empty, only
	// notifications whose Port appears in the list are forwarded.
	Ports []int

	// States is an optional allowlist of states (e.g. "open", "closed").
	// If non-empty, only notifications whose State appears in the list are
	// forwarded.
	States []string
}

// DefaultFilterPolicy returns a FilterPolicy that allows everything.
func DefaultFilterPolicy() FilterPolicy {
	return FilterPolicy{}
}

// NewFilter returns a Dispatcher that forwards notifications to next only when
// they satisfy the given FilterPolicy.
func NewFilter(policy FilterPolicy, next Dispatcher) *Filter {
	if next == nil {
		panic("alert: NewFilter requires a non-nil next dispatcher")
	}
	return &Filter{policy: policy, next: next}
}

// Filter is a Dispatcher that conditionally forwards to another Dispatcher.
type Filter struct {
	policy FilterPolicy
	next   Dispatcher
}

// Send forwards n to the next Dispatcher only if it passes the filter policy.
func (f *Filter) Send(n Notification) error {
	if !f.allow(n) {
		return nil
	}
	return f.next.Send(n)
}

func (f *Filter) allow(n Notification) bool {
	if len(f.policy.Ports) > 0 {
		matched := false
		for _, p := range f.policy.Ports {
			if p == n.Port {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(f.policy.States) > 0 {
		matched := false
		for _, s := range f.policy.States {
			if s == n.State {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}
