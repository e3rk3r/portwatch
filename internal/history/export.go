package history

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ExportFormat enumerates supported export formats.
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
)

// ExportJSON writes events as a JSON array to w.
func ExportJSON(w io.Writer, events []Event) error {
	type row struct {
		Port      int    `json:"port"`
		State     string `json:"state"`
		Timestamp string `json:"timestamp"`
	}
	rows := make([]row, len(events))
	for i, e := range events {
		rows[i] = row{
			Port:      e.Port,
			State:     e.State.String(),
			Timestamp: e.Timestamp.UTC().Format(time.RFC3339),
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}

// ExportCSV writes events as CSV (with header) to w.
func ExportCSV(w io.Writer, events []Event) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"port", "state", "timestamp"}); err != nil {
		return err
	}
	for _, e := range events {
		rec := []string{
			fmt.Sprintf("%d", e.Port),
			e.State.String(),
			e.Timestamp.UTC().Format(time.RFC3339),
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// Export dispatches to the appropriate format writer.
func Export(w io.Writer, events []Event, format ExportFormat) error {
	switch format {
	case FormatJSON:
		return ExportJSON(w, events)
	case FormatCSV:
		return ExportCSV(w, events)
	default:
		return fmt.Errorf("unsupported export format: %q", format)
	}
}
