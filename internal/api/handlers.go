package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/yourorg/portwatch/internal/history"
)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// handleHealthz returns a simple liveness probe.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleStatus returns the current tracked state for all ports.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, s.tracker.Snapshot())
}

// handleHistory returns recorded events, optionally filtered by port or since.
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := history.Query{}

	if p := r.URL.Query().Get("port"); p != "" {
		port, err := strconv.Atoi(p)
		if err != nil || port < 1 || port > 65535 {
			http.Error(w, "invalid port", http.StatusBadRequest)
			return
		}
		q.Port = port
	}

	if since := r.URL.Query().Get("since"); since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err != nil {
			http.Error(w, "invalid since: use RFC3339", http.StatusBadRequest)
			return
		}
		q.Since = t
	}

	if st := r.URL.Query().Get("state"); st != "" {
		q.State = st
	}

	events := history.Filter(s.ring.Snapshot(), q)
	writeJSON(w, http.StatusOK, events)
}
