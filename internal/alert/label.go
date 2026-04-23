package alert

import (
	"fmt"
	"strings"
	"sync"
)

// DefaultLabelPolicy returns a LabelPolicy with no static labels.
func DefaultLabelPolicy() LabelPolicy {
	return LabelPolicy{Labels: map[string]string{}}
}

// LabelPolicy defines static key/value labels to attach to every notification.
type LabelPolicy struct {
	// Labels is the set of key=value pairs to merge into Notification.Labels.
	Labels map[string]string
	// Prefix is prepended to every label key (optional).
	Prefix string
}

// Labeler attaches a fixed set of labels to notifications before forwarding.
type Labeler struct {
	mu     sync.RWMutex
	policy LabelPolicy
}

// NewLabeler constructs a Labeler from the given policy.
func NewLabeler(p LabelPolicy) *Labeler {
	if p.Labels == nil {
		p.Labels = map[string]string{}
	}
	return &Labeler{policy: p}
}

// Apply merges the policy labels into n.Labels, prefixing keys when configured.
// Existing keys in n.Labels are NOT overwritten.
func (l *Labeler) Apply(n *Notification) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if n.Labels == nil {
		n.Labels = map[string]string{}
	}
	for k, v := range l.policy.Labels {
		key := k
		if l.policy.Prefix != "" {
			key = fmt.Sprintf("%s%s", l.policy.Prefix, k)
		}
		if _, exists := n.Labels[key]; !exists {
			n.Labels[key] = v
		}
	}
}

// Set updates a single label at runtime (thread-safe).
func (l *Labeler) Set(key, value string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.policy.Labels[strings.TrimSpace(key)] = value
}

// Snapshot returns a copy of the current label map.
func (l *Labeler) Snapshot() map[string]string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make(map[string]string, len(l.policy.Labels))
	for k, v := range l.policy.Labels {
		out[k] = v
	}
	return out
}
