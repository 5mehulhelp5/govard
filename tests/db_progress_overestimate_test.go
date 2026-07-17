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

// TestImportProgressDoesNotLieAboutCompletion reproduces a real-world case:
// GetDatabaseSize() only estimates the dump size (SUM(data_length) from
// information_schema), so the real, uncompressed dump stream can end up
// noticeably smaller than that estimate. When the stream reaches EOF short
// of the estimated total, the copy-phase percentage must not claim "100%" -
// that would produce a frozen line like "Syncing DB [2.892GB/5.826GB] 100%"
// where the visible byte count and the percentage contradict each other. The
// bar is only allowed to reach 100% afterwards, once the import process has
// actually finished (the "- 100% - Import finalized" text appended by
// waitForImportCompletion) - so this only checks the copy-phase portion of
// the output, i.e. everything before "finalizing".
func TestImportProgressDoesNotLieAboutCompletion(t *testing.T) {
	originalWriter := pterm.DefaultProgressbar.Writer
	defer func() { pterm.DefaultProgressbar.Writer = originalWriter }()

	var captured bytes.Buffer
	pterm.DefaultProgressbar.Writer = &captured

	actualData := strings.Repeat("x", 500)
	estimatedTotal := int64(len(actualData) * 2) // estimate is double the real size

	importCmd := exec.Command("sh", "-c", "cat > /dev/null")

	if err := cmd.RunImportFromReaderWithProgress(importCmd, strings.NewReader(actualData), estimatedTotal, false, os.Stdout, os.Stderr, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := captured.String()
	copyPhase, _, found := strings.Cut(output, "finalizing")
	if !found {
		t.Fatalf("expected the output to reach the finalizing phase, got: %q", output)
	}
	if strings.Contains(copyPhase, "100%") {
		t.Fatalf("progress bar falsely reported 100%% during the copy phase after transferring only half of the estimated total; copy-phase output: %q", copyPhase)
	}
}

// TestImportProgressReportsShortfallNote covers the follow-up UX issue: a bar
// that legitimately stops below 100% (because the pre-sync estimate
// overshot the real dump size) looks identical to a stalled/failed transfer
// unless something explicitly says the stream did complete. A one-line note
// should make that clear whenever the real byte count landed short of the
// estimate.
//
// This note is deliberately NOT written through the importCmd.Stdout writer
// (passed here as os.Stdout, a real *os.File): exec.Cmd can share a *os.File
// with the child process directly, but a non-file io.Writer (e.g. a
// bytes.Buffer) makes exec.Cmd spawn its own background copy goroutine,
// which would race against a direct write from this test/production code on
// the same writer. Capturing via pterm.SetDefaultOutput avoids that hazard.
func TestImportProgressReportsShortfallNote(t *testing.T) {
	originalWriter := pterm.DefaultProgressbar.Writer
	defer func() { pterm.DefaultProgressbar.Writer = originalWriter }()
	pterm.DefaultProgressbar.Writer = &bytes.Buffer{}

	var captured bytes.Buffer
	pterm.SetDefaultOutput(&captured)
	defer pterm.SetDefaultOutput(os.Stdout)

	actualData := strings.Repeat("x", 500)
	estimatedTotal := int64(len(actualData) * 2)

	importCmd := exec.Command("sh", "-c", "cat > /dev/null")

	if err := cmd.RunImportFromReaderWithProgress(importCmd, strings.NewReader(actualData), estimatedTotal, false, os.Stdout, os.Stderr, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(captured.String(), "Stream complete") {
		t.Fatalf("expected a note confirming the stream completed despite landing short of the estimate, got: %q", captured.String())
	}
}

// TestImportProgressNoShortfallNoteWhenEstimateMatched ensures the note is
// only shown when there actually was a shortfall - it shouldn't add noise to
// the common case where the transfer reaches (or exceeds) the estimate.
func TestImportProgressNoShortfallNoteWhenEstimateMatched(t *testing.T) {
	originalWriter := pterm.DefaultProgressbar.Writer
	defer func() { pterm.DefaultProgressbar.Writer = originalWriter }()
	pterm.DefaultProgressbar.Writer = &bytes.Buffer{}

	var captured bytes.Buffer
	pterm.SetDefaultOutput(&captured)
	defer pterm.SetDefaultOutput(os.Stdout)

	actualData := strings.Repeat("x", 500)
	estimatedTotal := int64(len(actualData)) // estimate matches reality exactly

	importCmd := exec.Command("sh", "-c", "cat > /dev/null")

	if err := cmd.RunImportFromReaderWithProgress(importCmd, strings.NewReader(actualData), estimatedTotal, false, os.Stdout, os.Stderr, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(captured.String(), "Stream complete") {
		t.Fatalf("did not expect a shortfall note when the transfer matched the estimate, got: %q", captured.String())
	}
}
