package tests

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"govard/internal/cmd"

	"github.com/pterm/pterm"
)

// newTestProgressBar builds a bar the same way production code does
// (pterm.DefaultProgressbar...Start()), redirecting output to a captured
// buffer, and returns a restore func to undo the writer override.
func newTestProgressBar(t *testing.T, total int, title string) (*pterm.ProgressbarPrinter, *bytes.Buffer) {
	t.Helper()
	originalWriter := pterm.DefaultProgressbar.Writer
	t.Cleanup(func() { pterm.DefaultProgressbar.Writer = originalWriter })

	var captured bytes.Buffer
	pterm.DefaultProgressbar.Writer = &captured

	bar, _ := pterm.DefaultProgressbar.WithTotal(total).
		WithTitle(title).
		WithShowCount(false).
		WithShowPercentage(false).
		WithShowElapsedTime(false).
		Start()
	return bar, &captured
}

// TestWaitForImportCompletionShowsFinalizingIndicator reproduces the reported
// bug: the progress bar reports 100% and disappears as soon as bytes have
// been streamed into the import command's stdin, even though the import
// process (e.g. mysql committing a large transaction) can keep running for
// a long time afterwards with zero UI feedback. The fix must keep animating
// the SAME bar with a "finalizing" indicator, and only mark 100% once the
// import process has actually exited.
func TestWaitForImportCompletionShowsFinalizingIndicator(t *testing.T) {
	bar, out := newTestProgressBar(t, 100, "Importing DB (~1kB estimated)")

	// Simulates an import command that has already consumed all its input
	// (stdin already closed by the caller) but keeps doing work afterwards,
	// analogous to mysql running its final COMMIT after all statements have
	// been read from the pipe.
	importCmd := exec.Command("sh", "-c", "sleep 0.2")
	if err := importCmd.Start(); err != nil {
		t.Fatalf("failed to start fake import command: %v", err)
	}

	start := time.Now()
	err := cmd.WaitForImportCompletionForTest(bar, "Importing DB (~1kB estimated)", time.Time{}, importCmd, nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed < 150*time.Millisecond {
		t.Fatalf("expected to block until the import process actually exited, only waited %v", elapsed)
	}
	if !strings.Contains(out.String(), "finalizing") {
		t.Fatalf("expected a finalizing indicator to be shown while waiting for import completion, got: %q", out.String())
	}
	if !strings.Contains(out.String(), "100%") {
		t.Fatalf("expected the true completion signal (100%%) once the import process has actually exited, got: %q", out.String())
	}
}

func TestWaitForImportCompletionPropagatesFailure(t *testing.T) {
	bar, out := newTestProgressBar(t, 100, "Importing DB (~1kB estimated)")

	importCmd := exec.Command("sh", "-c", "exit 1")
	if err := importCmd.Start(); err != nil {
		t.Fatalf("failed to start fake import command: %v", err)
	}

	err := cmd.WaitForImportCompletionForTest(bar, "Importing DB (~1kB estimated)", time.Time{}, importCmd, nil)
	if err == nil {
		t.Fatal("expected an error when the import command exits non-zero")
	}
	if strings.Contains(out.String(), "100%") {
		t.Fatalf("did not expect a 100%% completion signal when the import command failed, got: %q", out.String())
	}
}

// TestWaitForImportCompletionPollsTargetSize reproduces the real-world
// complaint that a multi-GB import can sit at "finalizing..." for tens of
// minutes with zero indication it's still alive. When a poll func is
// supplied, the bar's title must be updated with the polled size at least
// once while waiting - a stub stands in for GetDatabaseSize so the test
// stays hermetic (no Docker/mysql dependency).
func TestWaitForImportCompletionPollsTargetSize(t *testing.T) {
	restore := cmd.SetFinalizePollIntervalForTest(10 * time.Millisecond)
	defer restore()

	bar, out := newTestProgressBar(t, 100, "Importing DB (~1kB estimated)")

	importCmd := exec.Command("sh", "-c", "sleep 0.2")
	if err := importCmd.Start(); err != nil {
		t.Fatalf("failed to start fake import command: %v", err)
	}

	polls := 0
	poll := func() (int64, error) {
		polls++
		return int64(polls) * 1024 * 1024 * 1024, nil
	}

	if err := cmd.WaitForImportCompletionForTest(bar, "Importing DB (~1kB estimated)", time.Time{}, importCmd, poll); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if polls == 0 {
		t.Fatal("expected the poll function to be called at least once while waiting")
	}
	if !strings.Contains(out.String(), "written so far") {
		t.Fatalf("expected the bar to report the polled size, got: %q", out.String())
	}
}

// TestWaitForImportCompletionPollErrorsAreIgnored ensures a transient poll
// failure (e.g. the target briefly unreachable during a heavy commit) never
// aborts the wait or surfaces as the operation's error.
func TestWaitForImportCompletionPollErrorsAreIgnored(t *testing.T) {
	restore := cmd.SetFinalizePollIntervalForTest(10 * time.Millisecond)
	defer restore()

	bar, _ := newTestProgressBar(t, 100, "Importing DB (~1kB estimated)")

	importCmd := exec.Command("sh", "-c", "sleep 0.1")
	if err := importCmd.Start(); err != nil {
		t.Fatalf("failed to start fake import command: %v", err)
	}

	poll := func() (int64, error) {
		return 0, errors.New("target briefly unreachable")
	}

	if err := cmd.WaitForImportCompletionForTest(bar, "Importing DB (~1kB estimated)", time.Time{}, importCmd, poll); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
