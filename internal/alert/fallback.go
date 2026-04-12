package alert

import (
	"context"
	"log"
)

// FallbackDispatcher tries the primary dispatcher first; if it returns an
// error it delegates to the secondary (fallback) dispatcher instead.
// This is useful for pairing a fast webhook with a slower but more reliable
// channel such as a local script or log sink.
type FallbackDispatcher struct {
	primary  Dispatcher
	fallback Dispatcher
	logger   *log.Logger
}

// NewFallbackDispatcher constructs a FallbackDispatcher.
// Both primary and fallback must be non-nil.
func NewFallbackDispatcher(primary, fallback Dispatcher, logger *log.Logger) *FallbackDispatcher {
	if primary == nil {
		panic("fallback: primary dispatcher must not be nil")
	}
	if fallback == nil {
		panic("fallback: fallback dispatcher must not be nil")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &FallbackDispatcher{
		primary:  primary,
		fallback: fallback,
		logger:   logger,
	}
}

// Send attempts delivery via the primary dispatcher. On error it logs the
// failure and retries via the fallback dispatcher, returning the fallback
// result to the caller.
func (f *FallbackDispatcher) Send(ctx context.Context, n Notification) error {
	if err := f.primary.Send(ctx, n); err != nil {
		f.logger.Printf("fallback: primary dispatcher failed (port=%d state=%s): %v; trying fallback",
			n.Port, n.State, err)
		return f.fallback.Send(ctx, n)
	}
	return nil
}
