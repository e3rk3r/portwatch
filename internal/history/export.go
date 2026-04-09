package history

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Format represents a supported export format.
type Format string

const (
	FormatJSON Format = "json"
	FormatCSV  Format = "csv"
)

// ExportJSON writes the given events as a JSON array to w.
func ExportJSON(w io.Writer, events []Event) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(events); err != nil {
		return fmt.Errorf("history: json export: %w", err)
	}
	return nil
}

// ExportCSV writes the given events as CSV rows to w.
// The header row is always written, even when events is empty.
func ExportCSV(w io.Writer, events []Event) error {
	cw := csv.NewWriter(w)

	if err := cw.Write([]string{"timestamp", "port", "state"}); err != nil {
		return fmt.Errorf("history: csv header: %w", err)
	}

	for _, e := range events {
		row := []string{
			e.Timestamp.UTC().Format(time.RFC3339),
			fmt.Sprintf("%d", e.Port),
			e.State,
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("history: csv row: %w", err)
		}
	}

	cw.Flush()
	return cw.Error()
}

// Export dispatches to ExportJSON or ExportCSV based on fmt.
func Export(w io.Writer, events []Event, format Format) error {
	switch format {
	case FormatJSON:
		return ExportJSON(w, events)
	case FormatCSV:
		return ExportCSV(w, events)
	default:
		return fmt.Errorf("history: unsupported export format %q", format)
	}
}
