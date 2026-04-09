// Package daemon provides the top-level orchestration loop for portwatch.
//
// The Daemon ties together the poller, state tracker, and action executor:
//
//  1. On each tick (configured interval), it polls all configured ports via
//     the poller package.
//  2. Each result is passed to the monitor.Tracker to detect state changes
//     (open <-> closed).
//  3. When a change is detected, the action.Executor fires any configured
//     webhooks or scripts for that port and new state.
//
// Usage:
//
//	cfg, _ := config.Load("portwatch.yaml")
//	d := daemon.New(cfg)
//	d.Run(ctx)
package daemon
