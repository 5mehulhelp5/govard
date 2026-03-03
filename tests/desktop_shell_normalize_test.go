package tests

import (
	"testing"

	"govard/internal/desktop"
)

func TestDesktopPkgNormalizeShellForTest_DefaultsToSh(t *testing.T) {
	if got := desktop.NormalizeShellForTest(""); got != "sh" {
		t.Fatalf("expected sh, got %q", got)
	}
}

func TestDesktopPkgNormalizeShellForTest_RespectsKnownShells(t *testing.T) {
	if got := desktop.NormalizeShellForTest("bash"); got != "bash" {
		t.Fatalf("expected bash, got %q", got)
	}
	if got := desktop.NormalizeShellForTest("sh"); got != "sh" {
		t.Fatalf("expected sh, got %q", got)
	}
	if got := desktop.NormalizeShellForTest("  BASH  "); got != "bash" {
		t.Fatalf("expected normalized bash, got %q", got)
	}
}
