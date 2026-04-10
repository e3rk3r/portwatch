package api

import (
	"net/http"
	"strconv"

	"github.com/user/portwatch/internal/history"
)

// handleExport serves GET /export?format=json|csv&port=N&limit=N
func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	format := history.ExportFormat(q.Get("format"))
	if format == "" {
		format = history.FormatJSON
	}

	filter := history.QueryFilter{}

	if portStr := q.Get("port"); portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil || p < 1 || p > 65535 {
			http.Error(w, "invalid port", http.StatusBadRequest)
			return
		}
		filter.Port = p
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		filter.Limit = l
	}

	events := history.Query(s.ring.Snapshot(), filter)

	switch format {
	case history.FormatCSV:
		w.Header().Set("Content-Type", "text/csv")
	default:
		w.Header().Set("Content-Type", "application/json")
	}

	if err := history.Export(w, events, format); err != nil {
		// Headers already sent; log only.
		_ = err
	}
}
