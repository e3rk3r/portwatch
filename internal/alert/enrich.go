package alert

import (
	"context"
	"fmt"
	"time"
)

// EnrichPolicy controls how notifications are enriched before dispatch.
type EnrichPolicy struct {
	// StaticLabels are merged into every notification's Labels map.
	StaticLabels map[string]string
	// HostnameLabel, when non-empty, adds the given key with the current hostname.
	HostnameLabel string
	// TimestampLabel, when non-empty, adds the given key with an RFC3339 timestamp.
	TimestampLabel string
}

// DefaultEnrichPolicy returns a no-op policy.
func DefaultEnrichPolicy() EnrichPolicy {
	return EnrichPolicy{}
}

// Enricher mutates a Notification's Labels field with additional metadata.
type Enricher struct {
	policy   EnrichPolicy
	hostname string
	now      func() time.Time
}

// NewEnricher constructs an Enricher from the given policy.
func NewEnricher(p EnrichPolicy, hostname string) *Enricher {
	if hostname == "" {
		hostname = "unknown"
	}
	return &Enricher{policy: p, hostname: hostname, now: time.Now}
}

// Enrich returns a copy of n with additional labels applied.
func (e *Enricher) Enrich(n Notification) Notification {
	if n.Labels == nil {
		n.Labels = make(map[string]string)
	}
	for k, v := range e.policy.StaticLabels {
		n.Labels[k] = v
	}
	if e.policy.HostnameLabel != "" {
		n.Labels[e.policy.HostnameLabel] = e.hostname
	}
	if e.policy.TimestampLabel != "" {
		n.Labels[e.policy.TimestampLabel] = e.now().UTC().Format(time.RFC3339)
	}
	return n
}

// EnrichDispatcher wraps a Dispatcher and enriches each notification before
// forwarding it downstream.
type EnrichDispatcher struct {
	enricher *Enricher
	next     Dispatcher
}

// NewEnrichDispatcher panics if enricher or next are nil.
func NewEnrichDispatcher(enricher *Enricher, next Dispatcher) *EnrichDispatcher {
	if enricher == nil {
		panic("alert: NewEnrichDispatcher: enricher must not be nil")
	}
	if next == nil {
		panic("alert: NewEnrichDispatcher: next must not be nil")
	}
	return &EnrichDispatcher{enricher: enricher, next: next}
}

// Send enriches n and forwards it to the next dispatcher.
func (d *EnrichDispatcher) Send(ctx context.Context, n Notification) error {
	enriched := d.enricher.Enrich(n)
	if err := d.next.Send(ctx, enriched); err != nil {
		return fmt.Errorf("enrich dispatcher: %w", err)
	}
	return nil
}
