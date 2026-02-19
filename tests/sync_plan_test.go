package tests

import (
	"strings"
	"testing"

	"govard/internal/engine/remote"
)

func TestBuildSyncPlan(t *testing.T) {
	plan := remote.BuildSyncPlan(remote.SyncOptions{
		Source:      "staging",
		Destination: "local",
		Files:       true,
		Resume:      true,
		Include:     []string{"app/*"},
		Exclude:     []string{"vendor/"},
	})
	if plan.Source != "staging" {
		t.Fatal("source mismatch")
	}
	if !strings.Contains(plan.Command, "rsync") {
		t.Fatal("expected rsync command")
	}
	if !strings.Contains(plan.Command, `--include "app/*"`) {
		t.Fatalf("expected include pattern in command, got: %s", plan.Command)
	}
	if !strings.Contains(plan.Command, `--exclude "vendor/"`) {
		t.Fatalf("expected exclude pattern in command, got: %s", plan.Command)
	}
	if !strings.Contains(plan.Command, "--partial --append-verify") {
		t.Fatalf("expected resume args in command, got: %s", plan.Command)
	}
}
