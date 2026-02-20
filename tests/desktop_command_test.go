package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/cmd"
)

func TestFindDesktopBinaryForTestMissingBinaryOutsideRepo(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to read working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	workDir := filepath.Join(t.TempDir(), "outside-repo")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("failed to create working directory: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	_, err = cmd.FindDesktopBinaryForTest()
	if err == nil {
		t.Fatal("expected missing desktop binary error")
	}

	message := err.Error()
	if !strings.Contains(message, "govard-desktop binary not found in PATH") {
		t.Fatalf("expected PATH guidance in error, got: %q", message)
	}
	if strings.Contains(message, "could not locate repository root") {
		t.Fatalf("error should not leak repo root lookup details: %q", message)
	}
}
