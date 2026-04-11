package alert

import (
	"fmt"
	"sync"
)

// RoutePolicy defines rules for directing notifications to named pipelines.
type RoutePolicy struct {
	// Routes maps a pipeline name to a list of port numbers it should handle.
	// An empty Ports slice means the pipeline receives all notifications.
	Routes []Route
}

// Route associates a named dispatcher with an optional set of port filters.
type Route struct {
	Name  string
	Ports []int // empty = match all
	States []string // empty = match all; values: "open", "closed"
}

// DefaultRoutePolicy returns a policy with no routes (pass-through).
func DefaultRoutePolicy() RoutePolicy {
	return RoutePolicy{}
}

// Router dispatches a notification to one or more downstream dispatchers
// based on port/state routing rules.
type Router struct {
	mu      sync.RWMutex
	routes  []routeEntry
}

type routeEntry struct {
	route      Route
	dispatcher Dispatcher
}

// NewRouter creates a Router. Each (Route, Dispatcher) pair is registered in
// order; the first matching route wins unless multicast is desired — callers
// may register the same dispatcher under multiple routes.
func NewRouter() *Router {
	return &Router{}
}

// Register adds a route → dispatcher mapping to the router.
func (r *Router) Register(route Route, d Dispatcher) error {
	if d == nil {
		return fmt.Errorf("router: dispatcher for route %q must not be nil", route.Name)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = append(r.routes, routeEntry{route: route, dispatcher: d})
	return nil
}

// Send delivers n to every dispatcher whose route matches the notification.
// All matching dispatchers are called; errors are collected and the last
// non-nil error is returned.
func (r *Router) Send(n Notification) error {
	r.mu.RLock()
	entries := make([]routeEntry, len(r.routes))
	copy(entries, r.routes)
	r.mu.RUnlock()

	var lastErr error
	matched := false
	for _, e := range entries {
		if routeMatches(e.route, n) {
			matched = true
			if err := e.dispatcher.Send(n); err != nil {
				lastErr = err
			}
		}
	}
	if !matched && len(entries) > 0 {
		// no route matched — silently drop
		return nil
	}
	return lastErr
}

func routeMatches(route Route, n Notification) bool {
	if len(route.Ports) > 0 {
		found := false
		for _, p := range route.Ports {
			if p == n.Port {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(route.States) > 0 {
		found := false
		for _, s := range route.States {
			if s == n.State {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
