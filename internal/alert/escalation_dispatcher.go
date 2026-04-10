package alert

import (
	"context"
	"fmt"
	"log"
)

// EscalationDispatcher wraps a primary Dispatcher and an escalation channel.
// When the Escalator signals a threshold breach, the notification is also
// forwarded to the escalation channel.
type EscalationDispatcher struct {
	primary    *Dispatcher
	escalation *Channel
	escalator  *Escalator
}

// NewEscalationDispatcher constructs an EscalationDispatcher.
func NewEscalationDispatcher(primary *Dispatcher, escalation *Channel, policy EscalationPolicy) *EscalationDispatcher {
	return &EscalationDispatcher{
		primary:    primary,
		escalation: escalation,
		escalator:  NewEscalator(policy),
	}
}

// Send delivers the notification via the primary dispatcher and, if the
// escalation threshold is met, also via the escalation channel.
func (ed *EscalationDispatcher) Send(ctx context.Context, n Notification) error {
	if err := ed.primary.Send(ctx, n); err != nil {
		return fmt.Errorf("primary dispatch: %w", err)
	}

	key := escalationKey(n)
	if n.State == "closed" {
		if ed.escalator.Record(key) {
			log.Printf("[escalation] threshold reached for %s, forwarding to escalation channel", key)
			if err := ed.escalation.Send(ctx, n); err != nil {
				log.Printf("[escalation] escalation channel error: %v", err)
			}
		}
	} else {
		ed.escalator.Reset(key)
	}
	return nil
}

func escalationKey(n Notification) string {
	return fmt.Sprintf("port:%d:%s", n.Port, n.State)
}
