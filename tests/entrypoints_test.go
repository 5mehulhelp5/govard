package tests

import (
	"os/exec"
	"strings"
	"testing"
)

func TestEntrypointDesktopStubs(t *testing.T) {
	for _, target := range []string{"../cmd/govard-desktop", "../desktop"} {
		cmd := exec.Command("go", "run", target)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("go run %s failed: %v\n%s", target, err, string(output))
		}
		if !strings.Contains(string(output), "not built yet") {
			t.Fatalf("expected not-built message for %s, got %q", target, string(output))
		}
	}
}
