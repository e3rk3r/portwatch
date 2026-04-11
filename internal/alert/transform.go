package alert

import (
	"fmt"
	"strings"
	"time"
)

// TransformPolicy defines how a Notification is mutated before delivery.
type TransformPolicy struct {
	// TitleTemplate overrides the notification title using simple placeholders:
	// {port}, {state}, {host}
	TitleTemplate string

	// AddLabels are merged into Notification.Labels.
	AddLabels map[string]string

	// StripSensitiveHeaders removes Authorization / Cookie headers from the
	// notification metadata so they are not forwarded to downstream channels.
	StripSensitiveHeaders bool
}

// DefaultTransformPolicy returns a no-op policy.
func DefaultTransformPolicy() TransformPolicy {
	return TransformPolicy{}
}

// Transformer applies a TransformPolicy to every Notification that passes
// through it before delegating to the next Dispatcher.
type Transformer struct {
	policy TransformPolicy
	next   Dispatcher
}

// NewTransformer constructs a Transformer. next must not be nil.
func NewTransformer(p TransformPolicy, next Dispatcher) *Transformer {
	if next == nil {
		panic("alert: NewTransformer: next dispatcher must not be nil")
	}
	return &Transformer{policy: p, next: next}
}

// Send transforms n according to the policy and forwards it.
func (t *Transformer) Send(n Notification) error {
	n = t.apply(n)
	return t.next.Send(n)
}

func (t *Transformer) apply(n Notification) Notification {
	// Render title template.
	if tpl := t.policy.TitleTemplate; tpl != "" {
		n.Title = renderTemplate(tpl, n)
	}

	// Merge extra labels.
	if len(t.policy.AddLabels) > 0 {
		if n.Labels == nil {
			n.Labels = make(map[string]string, len(t.policy.AddLabels))
		}
		for k, v := range t.policy.AddLabels {
			n.Labels[k] = v
		}
	}

	// Strip sensitive headers stored in Labels.
	if t.policy.StripSensitiveHeaders && n.Labels != nil {
		sensitive := []string{"authorization", "cookie", "set-cookie"}
		for _, key := range sensitive {
			for k := range n.Labels {
				if strings.EqualFold(k, key) {
					delete(n.Labels, k)
				}
			}
		}
	}

	return n
}

// renderTemplate replaces {port}, {state}, {host}, {time} in tpl.
func renderTemplate(tpl string, n Notification) string {
	r := strings.NewReplacer(
		"{port}", fmt.Sprintf("%d", n.Port),
		"{state}", n.State,
		"{host}", n.Host,
		"{time}", n.Timestamp.Format(time.RFC3339),
	)
	return r.Replace(tpl)
}
