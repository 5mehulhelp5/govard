package tests

import (
	"context"
	"strings"
	"testing"

	"govard/internal/desktop"
)

func TestDesktopPkgTerminalSessionLifecycle(t *testing.T) {
	t.Run("terminate unknown session is idempotent", func(t *testing.T) {
		svc := desktop.NewLogService()
		svc.Setup(context.Background())

		message, err := svc.TerminateTerminal("missing-session")
		if err != nil {
			t.Fatalf("terminate missing session: %v", err)
		}
		if !strings.Contains(strings.ToLower(message), "not found") {
			t.Fatalf("expected not found message, got %q", message)
		}
	})

	t.Run("terminate active session closes and removes it", func(t *testing.T) {
		svc := desktop.NewLogService()
		svc.Setup(context.Background())

		sessionID := "demo-session"
		cleanup, wasClosed := desktop.InjectTerminalSessionForTest(sessionID)
		defer cleanup()

		if !desktop.HasTerminalSessionForTest(sessionID) {
			t.Fatalf("expected injected session %q", sessionID)
		}

		message, err := svc.TerminateTerminal(sessionID)
		if err != nil {
			t.Fatalf("terminate session: %v", err)
		}
		if !strings.Contains(strings.ToLower(message), "terminated") {
			t.Fatalf("expected terminated message, got %q", message)
		}

		if desktop.HasTerminalSessionForTest(sessionID) {
			t.Fatalf("expected session %q to be removed", sessionID)
		}
		if !wasClosed() {
			t.Fatalf("expected session %q pty to be closed", sessionID)
		}
	})
}
