// Package api implements a lightweight HTTP server that exposes portwatch
// runtime data over a small REST interface.
//
// Endpoints:
//
//	GET /healthz          — liveness probe, always returns {"status":"ok"}
//	GET /status           — current port states tracked by the monitor
//	GET /history          — event history from the ring buffer
//	                        optional query params: port, state, since (RFC3339)
//
// The server is started by passing a context; cancelling the context triggers
// a graceful shutdown with a 5-second drain window.
package api
