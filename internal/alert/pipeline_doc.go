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
//
// Stage ordering:
//
//	The stages passed to NewPipeline should be ordered from outermost to
//	innermost — i.e. the first stage is called first and is responsible for
//	passing control to the next stage via its internal next reference.
//	Passing stages in the wrong order will not cause an error, but may
//	result in deduplication or suppression logic being bypassed.
package alert
