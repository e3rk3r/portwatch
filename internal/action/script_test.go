package action

import (
	"context"
	"testing"

	"github.com/user/portwatch/internal/config"
)

func TestFireScript_MissingPath(t *testing.T) {
	a := config.Action{Type: "script", On: "open", Path: ""}
	if err := fireScript(context.Background(), a, 8080, "open"); err == nil {
		t.Error("expected error when path is empty")
	}
}

func TestFireScript_EchoScript(t *testing.T) {
	a := config.Action{Type: "script", On: "open", Path: "/bin/echo"}
	if err := fireScript(context.Background(), a, 8080, "open"); err != nil {
		t.Errorf("unexpected error running echo: %v", err)
	}
}

func TestFireScript_BadScript(t *testing.T) {
	a := config.Action{Type: "script", On: "open", Path: "/nonexistent/script.sh"}
	if err := fireScript(context.Background(), a, 8080, "open"); err == nil {
		t.Error("expected error for nonexistent script")
	}
}
