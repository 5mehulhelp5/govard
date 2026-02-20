package tests

import (
	"path/filepath"
	"testing"
	"time"

	"govard/internal/cmd"
	"govard/internal/engine"
)

func TestTrackProjectRegistryForTest(t *testing.T) {
	t.Setenv("GOVARD_PROJECT_REGISTRY_PATH", filepath.Join(t.TempDir(), "projects.json"))

	cfg := engine.Config{
		ProjectName: "demo",
		Domain:      "demo.test",
		Recipe:      "magento2",
	}
	if err := cmd.TrackProjectRegistryForTest(cfg, "/workspace/demo", "up"); err != nil {
		t.Fatalf("track project registry: %v", err)
	}

	entries, err := engine.ReadProjectRegistryEntries()
	if err != nil {
		t.Fatalf("read project registry: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 registry entry, got %d", len(entries))
	}
	if entries[0].ProjectName != "demo" {
		t.Fatalf("expected project name demo, got %s", entries[0].ProjectName)
	}
	if entries[0].LastCommand != "up" {
		t.Fatalf("expected last command up, got %s", entries[0].LastCommand)
	}
}

func TestWriteOperationEventForTest(t *testing.T) {
	t.Setenv("GOVARD_OPERATIONS_LOG_PATH", filepath.Join(t.TempDir(), "operations.log"))

	cfg := engine.Config{ProjectName: "demo"}
	if err := cmd.WriteOperationEventForTest(
		"up.run",
		engine.OperationStatusSuccess,
		cfg,
		"",
		"",
		"completed",
		"",
		1500*time.Millisecond,
	); err != nil {
		t.Fatalf("write operation event: %v", err)
	}

	events, err := engine.ReadOperationEvents(10)
	if err != nil {
		t.Fatalf("read operation events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 operation event, got %d", len(events))
	}
	if events[0].Operation != "up.run" {
		t.Fatalf("expected operation up.run, got %s", events[0].Operation)
	}
	if events[0].Project != "demo" {
		t.Fatalf("expected project demo, got %s", events[0].Project)
	}
	if events[0].DurationMS != 1500 {
		t.Fatalf("expected duration 1500ms, got %d", events[0].DurationMS)
	}
}
