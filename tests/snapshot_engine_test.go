package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
	"govard/internal/engine"
)

func TestListSnapshotsReadsMetadata(t *testing.T) {
	root := t.TempDir()
	snapshotRoot := filepath.Join(root, ".govard", "snapshots")
	if err := os.MkdirAll(snapshotRoot, 0755); err != nil {
		t.Fatal(err)
	}

	meta1 := engine.SnapshotMetadata{
		Name:      "older",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		DB:        true,
		Media:     false,
	}
	meta2 := engine.SnapshotMetadata{
		Name:      "newer",
		CreatedAt: time.Now().Add(-1 * time.Hour),
		DB:        true,
		Media:     true,
	}

	writeSnapshotMeta(t, snapshotRoot, meta1)
	writeSnapshotMeta(t, snapshotRoot, meta2)

	list, err := engine.ListSnapshots(root)
	if err != nil {
		t.Fatalf("list snapshots: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(list))
	}
	if list[0].Name != "newer" {
		t.Fatalf("expected first snapshot to be newer, got %s", list[0].Name)
	}
}

func TestRestoreSnapshotMissing(t *testing.T) {
	err := engine.RestoreSnapshot(t.TempDir(), engine.Config{ProjectName: "demo"}, "missing", false, false)
	if err == nil {
		t.Fatal("expected restore missing snapshot to fail")
	}
}

func writeSnapshotMeta(t *testing.T, root string, meta engine.SnapshotMetadata) {
	t.Helper()
	dir := filepath.Join(root, meta.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	payload, err := yaml.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.yml"), payload, 0644); err != nil {
		t.Fatal(err)
	}
}
