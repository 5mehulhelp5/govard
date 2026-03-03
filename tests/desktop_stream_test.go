package tests

import (
	"strings"
	"testing"

	"govard/internal/desktop"
)

func TestDesktopPkgSanitizeStreamLineForTestStripsANSIAndControlChars(t *testing.T) {
	raw := []byte("\x1b[32mBootstrap\x1b[0m step \x07ok")

	got := desktop.SanitizeStreamLineForTest(raw)
	if strings.Contains(got, "\x1b") {
		t.Fatalf("expected ANSI escape codes stripped, got %q", got)
	}
	if strings.Contains(got, "\x07") {
		t.Fatalf("expected control characters stripped, got %q", got)
	}
	if got != "Bootstrap step ok" {
		t.Fatalf("unexpected sanitized output: %q", got)
	}
}

func TestDesktopPkgSanitizeStreamLineForTestDropsInvalidUTF8(t *testing.T) {
	raw := []byte{0xff, 0xfe, 'A', 'B'}

	got := desktop.SanitizeStreamLineForTest(raw)
	if got != "AB" {
		t.Fatalf("expected invalid bytes removed, got %q", got)
	}
}
