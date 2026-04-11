package alert

import (
	"fmt"
	"sync"
	"time"
)

// AuditEntry records a single dispatched notification and its outcome.
type AuditEntry struct {
	Timestamp time.Time
	Port      int
	State     string
	Channel   string
	Success   bool
	Err       string
}

// AuditLog is a bounded, thread-safe log of dispatch outcomes.
type AuditLog struct {
	mu      sync.Mutex
	entries []AuditEntry
	cap     int
}

// DefaultAuditCapacity is the default maximum number of audit entries retained.
const DefaultAuditCapacity = 256

// NewAuditLog creates an AuditLog with the given capacity.
// If cap <= 0, DefaultAuditCapacity is used.
func NewAuditLog(cap int) *AuditLog {
	if cap <= 0 {
		cap = DefaultAuditCapacity
	}
	return &AuditLog{cap: cap, entries: make([]AuditEntry, 0, cap)}
}

// Record appends an entry, evicting the oldest when at capacity.
func (a *AuditLog) Record(e AuditEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.entries) >= a.cap {
		a.entries = a.entries[1:]
	}
	a.entries = append(a.entries, e)
}

// Snapshot returns a copy of all current entries, oldest first.
func (a *AuditLog) Snapshot() []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]AuditEntry, len(a.entries))
	copy(out, a.entries)
	return out
}

// Len returns the current number of entries.
func (a *AuditLog) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.entries)
}

// auditKey returns a human-readable key for an entry (used in tests/logging).
func auditKey(port int, state, channel string) string {
	return fmt.Sprintf("%d:%s:%s", port, state, channel)
}
