package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
)

// fireWebhookCtx posts a JSON-encoded Notification to the given URL.
func fireWebhookCtx(ctx context.Context, url string, n Notification) error {
	body, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("channel: marshal notification: %w",
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
Errorf("channel: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "portwatch/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("channel: webhook post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("channel: webhook non-2xx status %d", resp.StatusCode)
	}
	return nil
}

// fireScriptCtx executes the script at path, passing notification fields
// as environment variables: PW_PORT, PW_STATE, PW_TIMESTAMP.
func fireScriptCtx(ctx context.Context, path string, n Notification) error {
	if path == "" {
		return fmt.Errorf("channel: script path is empty")
	}

	cmd := exec.CommandContext(ctx, path)
	cmd.Env = append(cmd.Environ(),
		"PW_PORT="+strconv.Itoa(n.Port),
		"PW_STATE="+n.State,
		"PW_TIMESTAMP="+n.Timestamp.UTC().String(),
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("channel: script %q failed: %w (output: %s)", path, err, out)
	}
	return nil
}
