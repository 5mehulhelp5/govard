package tests

import (
	"path/filepath"
	"testing"
	"time"

	"govard/internal/engine"
)

func TestReadProjectRegistryReturnsEmptyWhenMissing(t *testing.T) {
	t.Setenv("GOVARD_PROJECT_REGISTRY_PATH", filepath.Join(t.TempDir(), "projects.json"))

	entries, err := engine.ReadProjectRegistryEntries()
	if err != nil {
		t.Fatalf("read project registry: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty registry, got %d entries", len(entries))
	}
}

func TestUpsertProjectRegistryEntryRequiresPath(t *testing.T) {
	t.Setenv("GOVARD_PROJECT_REGISTRY_PATH", filepath.Join(t.TempDir(), "projects.json"))

	err := engine.UpsertProjectRegistryEntry(engine.ProjectRegistryEntry{ProjectName: "demo"})
	if err == nil {
		t.Fatal("expected path validation error")
	}
}

func TestUpsertProjectRegistryEntryUpdatesExistingAndSortsByLastSeen(t *testing.T) {
	t.Setenv("GOVARD_PROJECT_REGISTRY_PATH", filepath.Join(t.TempDir(), "projects.json"))

	older := time.Date(2026, 2, 19, 10, 0, 0, 0, time.UTC)
	newer := older.Add(2 * time.Hour)
	newest := older.Add(4 * time.Hour)

	if err := engine.UpsertProjectRegistryEntry(engine.ProjectRegistryEntry{
		Path:        "/workspace/demo",
		ProjectName: "demo",
		Domain:      "demo.test",
		Recipe:      "magento2",
		LastSeenAt:  older,
		LastCommand: "init",
	}); err != nil {
		t.Fatalf("upsert demo: %v", err)
	}

	if err := engine.UpsertProjectRegistryEntry(engine.ProjectRegistryEntry{
		Path:        "/workspace/shop",
		ProjectName: "shop",
		Domain:      "shop.test",
		Recipe:      "laravel",
		LastSeenAt:  newer,
		LastCommand: "up",
	}); err != nil {
		t.Fatalf("upsert shop: %v", err)
	}

	if err := engine.UpsertProjectRegistryEntry(engine.ProjectRegistryEntry{
		Path:        "/workspace/demo",
		ProjectName: "demo",
		Domain:      "demo.test",
		Recipe:      "magento2",
		LastSeenAt:  newest,
		LastCommand: "bootstrap",
	}); err != nil {
		t.Fatalf("upsert demo update: %v", err)
	}

	entries, err := engine.ReadProjectRegistryEntries()
	if err != nil {
		t.Fatalf("read project registry: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Path != "/workspace/demo" {
		t.Fatalf("expected updated demo entry first, got %s", entries[0].Path)
	}
	if entries[0].LastCommand != "bootstrap" {
		t.Fatalf("expected updated command bootstrap, got %s", entries[0].LastCommand)
	}
	if !entries[0].LastSeenAt.Equal(newest) {
		t.Fatalf("expected newest timestamp %s, got %s", newest, entries[0].LastSeenAt)
	}
}
