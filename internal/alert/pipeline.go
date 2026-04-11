// Package alert provides alerting primitives for portwatch.
package alert

import (
	"context"
	"fmt"
)

// Stage represents a single middleware step in an alert pipeline.
type Stage func(ctx context.Context, n Notification) error

// Pipeline chains multiple Dispatcher implementations into a single
// Dispatcher. Each stage is called in order; if any stage returns an
// error the chain is aborted and the error is returned to the caller.
type Pipeline struct {
	stages []Dispatcher
}

// NewPipeline constructs a Pipeline from an ordered list of Dispatchers.
// At least one dispatcher must be provided.
func NewPipeline(first Dispatcher, rest ...Dispatcher) (*Pipeline, error) {
	if first == nil {
		return nil, fmt.Errorf("pipeline: first dispatcher must not be nil")
	}
	all := make([]Dispatcher, 0, 1+len(rest))
	all = append(all, first)
	for i, d := range rest {
		if d == nil {
			return nil, fmt.Errorf("pipeline: dispatcher at index %d is nil", i+1)
		}
		all = append(all, d)
	}
	return &Pipeline{stages: all}, nil
}

// Send passes the notification through every stage in order.
// The first non-nil error short-circuits the remaining stages.
func (p *Pipeline) Send(ctx context.Context, n Notification) error {
	for _, d := range p.stages {
		if err := d.Send(ctx, n); err != nil {
			return err
		}
	}
	return nil
}

// Len returns the number of stages in the pipeline.
func (p *Pipeline) Len() int { return len(p.stages) }
