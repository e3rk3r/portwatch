package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/action"
	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/monitor"
	"github.com/user/portwatch/internal/poller"
)

// Daemon orchestrates polling, state tracking, and action execution.
type Daemon struct {
	cfg      *config.Config
	tracker  *monitor.Tracker
	executor *action.Executor
}

// New creates a Daemon from the given config.
func New(cfg *config.Config) *Daemon {
	return &Daemon{
		cfg:      cfg,
		tracker:  monitor.NewTracker(),
		executor: action.NewExecutor(cfg),
	}
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	interval := time.Duration(d.cfg.Interval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("portwatch daemon started (interval=%s, ports=%d)", interval, len(d.cfg.Ports))

	for {
		select {
		case <-ctx.Done():
			log.Println("portwatch daemon stopped")
			return ctx.Err()
		case <-ticker.C:
			d.tick(ctx)
		}
	}
}

func (d *Daemon) tick(ctx context.Context) {
	results := poller.PollAll(ctx, d.cfg.Ports)
	for _, r := range results {
		changed, prev, curr := d.tracker.Update(r.Host, r.Port, r.State)
		if !changed {
			continue
		}
		log.Printf("port %s:%d state change: %s -> %s", r.Host, r.Port, prev, curr)
		if err := d.executor.Run(ctx, r.Port, curr); err != nil {
			log.Printf("action error for port %d: %v", r.Port, err)
		}
	}
}
