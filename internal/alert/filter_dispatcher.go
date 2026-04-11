package alert

// NewFilterDispatcher is a convenience constructor that wraps next with a
// Filter using the provided ports and states allowlists. Pass nil or empty
// slices to skip filtering on that dimension.
//
// Example:
//
//	d := NewFilterDispatcher([]int{8080, 9090}, []string{"closed"}, downstream)
//	// Only closed-state events on port 8080 or 9090 reach downstream.
func NewFilterDispatcher(ports []int, states []string, next Dispatcher) Dispatcher {
	policy := FilterPolicy{
		Ports:  ports,
		States: states,
	}
	return NewFilter(policy, next)
}

// filterKey returns a string key used for logging / tracing purposes.
func filterKey(n Notification) string {
	return dedupKey(n) // reuse port:state format
}
