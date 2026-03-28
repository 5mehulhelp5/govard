package tests

import (
	"context"
	"os"
	"testing"

	"govard/internal/engine"
)

func TestGhostProjectCleanup(t *testing.T) {
	projectPath := "/home/kai/Work/htdocs/ghost-test"

	// Ensure project directory and container are NOT deleted yet
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Skipf("skipping: project directory %s does not exist", projectPath)
	}

	t.Logf("Running resilient cleanup for ghost project at: %s", projectPath)

	err := engine.DeleteProject(context.Background(), projectPath, os.Stdout, os.Stderr)
	if err != nil {
		t.Errorf("Resilient cleanup failed: %v", err)
	}

	// Verify registry removal (indirectly by trying to find it)
	_, err = engine.FindProjectByQuery("ghost-test")
	if err == nil {
		t.Error("Project 'ghost-test' still exists in registry after deletion")
	}
}
