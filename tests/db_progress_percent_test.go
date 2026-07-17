package tests

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"govard/internal/cmd"

	"github.com/pterm/pterm"
)

// TestDisplayPercentCapsBelow100UntilDone covers the core contract: the
// percentage shown during the copy phase must never claim 100% - the real
// "done" signal is the stream-complete note plus the finalize spinner
// afterwards, not this number.
func TestDisplayPercentCapsBelow100UntilDone(t *testing.T) {
	cases := []struct {
		name    string
		current int64
		total   int64
		want    int
	}{
		{"under estimate", 50, 100, 50},
		{"exactly at estimate", 100, 100, 99},
		{"past estimate", 500, 100, 99},
		{"zero total", 50, 0, 0},
		{"zero current", 0, 100, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := cmd.DisplayPercentForTest(c.current, c.total)
			if got != c.want {
				t.Fatalf("DisplayPercentForTest(%d, %d) = %d, want %d", c.current, c.total, got, c.want)
			}
			if got >= 100 {
				t.Fatalf("DisplayPercentForTest(%d, %d) = %d, must never reach 100", c.current, c.total, got)
			}
		})
	}
}

// TestImportProgressSurvivesUnderestimate reproduces the more serious bug
// found while investigating this: pterm's ProgressbarPrinter.Add snaps
// Total=Current and permanently deactivates the bar (all further UpdateTitle
// calls become no-ops) the instant Current reaches Total. If the pre-sync
// estimate undershoots the real stream size, that would freeze the entire
// display - byte count, throughput, ETA - while a possibly large remainder
// keeps copying silently in the background. The fix must keep the bar alive
// and its title updating for the whole transfer, no matter how far past the
// estimate it goes.
func TestImportProgressSurvivesUnderestimate(t *testing.T) {
	originalWriter := pterm.DefaultProgressbar.Writer
	defer func() { pterm.DefaultProgressbar.Writer = originalWriter }()
	var captured bytes.Buffer
	pterm.DefaultProgressbar.Writer = &captured

	// Real data is 5x the estimate.
	actualData := strings.Repeat("x", 1000)
	estimatedTotal := int64(200)

	importCmd := exec.Command("sh", "-c", "cat > /dev/null")
	if err := cmd.RunImportFromReaderWithProgress(importCmd, strings.NewReader(actualData), estimatedTotal, false, os.Stdout, os.Stderr, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := captured.String()
	if strings.Contains(output, "500%") {
		t.Fatalf("expected the percentage to never overshoot past the cap, got: %q", output)
	}
	copyPhase, _, found := strings.Cut(output, "finalizing")
	if !found {
		t.Fatalf("expected the output to reach the finalizing phase, got: %q", output)
	}
	if strings.Contains(copyPhase, "100%") {
		t.Fatalf("expected the bar to never claim 100%% mid-copy when the estimate undershoots, copy-phase output: %q", copyPhase)
	}
	if output == "" {
		t.Fatal("expected the bar to keep rendering for the entire transfer, got no output at all")
	}
}

// TestImportProgressTitleOmitsByteCount ensures the title only shows the
// estimate once (as static context) plus percentage/rate/ETA - not a
// continuously-updating "current/total" byte pair, which read as more
// precise than the underlying estimate actually is.
func TestImportProgressTitleOmitsByteCount(t *testing.T) {
	originalWriter := pterm.DefaultProgressbar.Writer
	defer func() { pterm.DefaultProgressbar.Writer = originalWriter }()
	var captured bytes.Buffer
	pterm.DefaultProgressbar.Writer = &captured

	actualData := strings.Repeat("x", 500)
	estimatedTotal := int64(len(actualData))

	importCmd := exec.Command("sh", "-c", "cat > /dev/null")
	if err := cmd.RunImportFromReaderWithProgress(importCmd, strings.NewReader(actualData), estimatedTotal, false, os.Stdout, os.Stderr, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := captured.String()
	if !strings.Contains(output, "estimated") {
		t.Fatalf("expected the estimated total to be shown as static context, got: %q", output)
	}
	if strings.Contains(output, "/500B") || strings.Contains(output, "[0B") {
		t.Fatalf("did not expect a current/total byte pair in the title, got: %q", output)
	}
}
