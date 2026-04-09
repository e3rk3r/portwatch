package history

import (
	"time"

	"github.com/user/portwatch/internal/config"
)

// Event represents a recorded port state change.
type Event struct {
	Port      int
	State     string
	Timestamp time.Time
}

// Filter holds criteria for querying history events.
type Filter struct {
	// Port filters events to a specific port; 0 means all ports.
	Port int
	// State filters events by state ("open" or "closed"); empty means all.
	State string
	// Since filters events that occurred at or after this time.
	Since time.Time
	// Limit caps the number of results; 0 means no limit.
	Limit int
}

// Query returns events from the ring buffer that match the given filter.
// Results are returned in chronological order (oldest first).
func (r *Ring) Query(f Filter) []Event {
	snapshot := r.Snapshot()

	var results []Event
	for _, e := range snapshot {
		if f.Port != 0 && e.Port != f.Port {
			continue
		}
		if f.State != "" && e.State != f.State {
			continue
		}
		if !f.Since.IsZero() && e.Timestamp.Before(f.Since) {
			continue
		}
		results = append(results, e)
	}

	if f.Limit > 0 && len(results) > f.Limit {
		results = results[len(results)-f.Limit:]
	}

	return results
}

// Summary returns a map of port -> last known state derived from history.
func (r *Ring) Summary(ports []config.PortConfig) map[int]string {
	snapshot := r.Snapshot()

	last := make(map[int]string, len(ports))
	// initialise with unknown so every configured port appears
	for _, p := range ports {
		last[p.Port] = "unknown"
	}

	for _, e := range snapshot {
		last[e.Port] = e.State
	}

	return last
}
