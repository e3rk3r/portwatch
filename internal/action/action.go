package action

import (
	"context"
	"fmt"
	"log"

	"github.com/user/portwatch/internal/config"
)

// Executor dispatches actions when port state changes occur.
type Executor struct {
	cfg *config.Config
}

// NewExecutor creates an Executor backed by the provided config.
func NewExecutor(cfg *config.Config) *Executor {
	return &Executor{cfg: cfg}
}

// Run finds the matching port config and fires every configured action.
func (e *Executor) Run(ctx context.Context, port int, state string) error {
	for _, pc := range e.cfg.Ports {
		if pc.Port != port {
			continue
		}
		for _, a := range pc.Actions {
			if a.On != state && a.On != "any" {
				continue
			}
			var err error
			switch a.Type {
			case "webhook":
				err = fireWebhook(ctx, a, port, state)
			case "script":
				err = fireScript(ctx, a, port, state)
			default:
				err = fmt.Errorf("unknown action type: %s", a.Type)
			}
			if err != nil {
				log.Printf("action error (port=%d type=%s): %v", port, a.Type, err)
			}
		}
		return nil
	}
	return fmt.Errorf("no config found for port %d", port)
}
