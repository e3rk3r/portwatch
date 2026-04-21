package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

func enrichNotif() Notification {
	return Notification{Port: 8080, State: "open", Title: "test"}
}

func TestEnricher_NoOpPolicy(t *testing.T) {
	e := NewEnricher(DefaultEnrichPolicy(), "")
	out := e.Enrich(enrichNotif())
	if len(out.Labels) != 0 {
		t.Fatalf("expected no labels, got %v", out.Labels)
	}
}

func TestEnricher_StaticLabels(t *testing.T) {
	p := EnrichPolicy{StaticLabels: map[string]string{"env": "prod", "team": "ops"}}
	e := NewEnricher(p, "host1")
	out := e.Enrich(enrichNotif())
	if out.Labels["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", out.Labels["env"])
	}
	if out.Labels["team"] != "ops" {
		t.Errorf("expected team=ops, got %q", out.Labels["team"])
	}
}

func TestEnricher_HostnameLabel(t *testing.T) {
	p := EnrichPolicy{HostnameLabel: "host"}
	e := NewEnricher(p, "myserver")
	out := e.Enrich(enrichNotif())
	if out.Labels["host"] != "myserver" {
		t.Errorf("expected host=myserver, got %q", out.Labels["host"])
	}
}

func TestEnricher_TimestampLabel(t *testing.T) {
	fixed := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	p := EnrichPolicy{TimestampLabel: "ts"}
	e := NewEnricher(p, "h")
	e.now = func() time.Time { return fixed }
	out := e.Enrich(enrichNotif())
	expected := fixed.Format(time.RFC3339)
	if out.Labels["ts"] != expected {
		t.Errorf("expected ts=%, got %q", expected, out.Labels["ts"])
	}
}

func TestEnricher_PreservesExistingLabels(t *testing.T) {
	p := EnrichPolicy{StaticLabels: map[string]string{"env": "prod"}}
	e := NewEnricher(p, "h")
	n := enrichNotif()
	n.Labels = map[string]string{"existing": "value"}
	out := e.Enrich(n)
	if out.Labels["existing"] != "value" {
		t.Error("existing label was overwritten")
	}
	if out.Labels["env"] != "prod" {
		t.Error("static label missing")
	}
}

func TestEnrichricherPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil enricher")
		}
	}()
	NewEnrichDispatcher(nil, &captureDispatcher{})
}

func TestEnrichDispatcher_NilNextPanics(t *testing.T) {

		if r := recover(); r == nil {
			t.Error("expected panic for nil next")
		}
	}()
	NewEnrichDispatcher(NewEnricherh"), nil)
}

func TestEnrichDispatcher_ForwardsEnrichedNotification(t *testing.T) {
	p := EnrichPolicy{StaticLabels: map[string]string{"env": "staging"}}
	e := NewEnricher(p, "srv")
	cap := &captureDispatcher{}
	d := NewEnrichDispatcher(e, cap)

	if err := d.Send(context.Background(), enrichNotif()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.last.Labels["env"] != "staging" {
		t.Errorf("expected enriched label, got %v", cap.last.Labels)
	}
}

func TestEnrichDispatcher_PropagatesError(t *testing.T) {
	e := NewEnricher(DefaultEnrichPolicy(), "h")
	fail := &errorDispatcher{err: errors.New("downstream failure")}
	d := NewEnrichDispatcher(e, fail)
	err := d.Send(context.Background(), enrichNotif())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
