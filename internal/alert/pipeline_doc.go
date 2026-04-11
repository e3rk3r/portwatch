// Package alert — pipeline.go
//
// Pipeline provides a composable, ordered chain of Dispatcher stages.
//
// Usage:
//
//	// Build individual dispatchers.
//	 dedup  := alert.NewDedupDispatcher(deduplicator, next)
//	 limiter := alert.NewSuppressDispatcher(suppressor, dedup)
//	 digest  := alert.NewDigestDispatcher(digester, limiter)
//
//	// Combine them into a single entry-point.
//	 pipe, err := alert.NewPipeline(digest, limiter, dedup)
//	 if err != nil {
//	     log.Fatal(err)
//	 }
//
//	 // Send a notification through the whole chain.
//	 if err := pipe.Send(ctx, notification); err != nil {
//	     log.Printf("alert pipeline error: %v", err)
//	 }
//
// If any stage returns an error the chain is aborted immediately and
// subsequent stages are skipped. Stages are executed synchronously in
// the order they were passed to NewPipeline.
package alert
