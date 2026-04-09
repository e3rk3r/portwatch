package action

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/user/portwatch/internal/config"
)

func fireScript(ctx context.Context, a config.Action, port int, state string) error {
	if a.Path == "" {
		return fmt.Errorf("script action missing path")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, a.Path, strconv.Itoa(port), state)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script %q failed: %w (output: %s)", a.Path, err, string(out))
	}
	return nil
}
