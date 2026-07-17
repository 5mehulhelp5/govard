package tests

import (
	"strings"
	"testing"
	"time"

	"govard/internal/cmd"
)

// TestFormatThroughputSuffixShowsRateAndETA covers the common case: some
// bytes moved over a non-trivial window, with an estimated total still ahead.
func TestFormatThroughputSuffixShowsRateAndETA(t *testing.T) {
	// 10 MiB in 1s => 10 MiB/s; 100 MiB remaining => ETA ~10s.
	deltaBytes := int64(10 * 1024 * 1024)
	elapsed := time.Second
	remaining := int64(100 * 1024 * 1024)

	suffix := cmd.FormatThroughputSuffixForTest(deltaBytes, elapsed, remaining)

	if !strings.Contains(suffix, "/s") {
		t.Fatalf("expected a rate figure in the suffix, got: %q", suffix)
	}
	if !strings.Contains(suffix, "ETA") {
		t.Fatalf("expected an ETA in the suffix, got: %q", suffix)
	}
}

// TestFormatThroughputSuffixHidesETAWhenTotalAlreadyReached covers the case
// where the actual transfer has caught up to (or passed) the estimated
// total - showing a negative or zero ETA would be nonsensical.
func TestFormatThroughputSuffixHidesETAWhenTotalAlreadyReached(t *testing.T) {
	suffix := cmd.FormatThroughputSuffixForTest(1024, time.Second, 0)

	if !strings.Contains(suffix, "/s") {
		t.Fatalf("expected the rate to still be shown, got: %q", suffix)
	}
	if strings.Contains(suffix, "ETA") {
		t.Fatalf("expected no ETA once the estimated total has been reached, got: %q", suffix)
	}
}

// TestFormatThroughputSuffixEmptyWithoutSignal covers the "not enough data
// yet" cases: no bytes moved since the last sample, or no elapsed time.
func TestFormatThroughputSuffixEmptyWithoutSignal(t *testing.T) {
	cases := []struct {
		name       string
		deltaBytes int64
		elapsed    time.Duration
		remaining  int64
	}{
		{"no bytes since last sample", 0, time.Second, 1024},
		{"no elapsed time", 1024, 0, 1024},
		{"negative delta", -5, time.Second, 1024},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			suffix := cmd.FormatThroughputSuffixForTest(c.deltaBytes, c.elapsed, c.remaining)
			if suffix != "" {
				t.Fatalf("expected an empty suffix, got: %q", suffix)
			}
		})
	}
}
