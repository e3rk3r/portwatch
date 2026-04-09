// Package api provides a lightweight HTTP API for querying portwatch status and history.
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/yourorg/portwatch/internal/history"
	"github.com/yourorg/portwatch/internal/monitor"
)

// Server exposes a small REST API over HTTP.
type Server struct {
	addr    string
	ring    *history.Ring
	tracker *monitor.Tracker
	httpSrv *http.Server
}

// New creates a new API Server bound to addr.
func New(addr string, ring *history.Ring, tracker *monitor.Tracker) *Server {
	s := &Server{
		addr:    addr,
		ring:    ring,
		tracker: tracker,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/history", s.handleHistory)
	mux.HandleFunc("/healthz", s.handleHealthz)

	s.httpSrv = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return s
}

// Start begins serving HTTP requests. It blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.httpSrv.Shutdown(shutCtx)
	}
}
