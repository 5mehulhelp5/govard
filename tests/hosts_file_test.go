package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/engine"
)

func TestAddHostsEntryIsIdempotent(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "hosts")
	initial := "127.0.0.1 localhost\n"
	if err := os.WriteFile(tempFile, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	restore := engine.SetHostsFilePathForTest(tempFile)
	defer restore()

	if err := engine.AddHostsEntry("demo.test"); err != nil {
		t.Fatalf("add hosts entry: %v", err)
	}
	if err := engine.AddHostsEntry("demo.test"); err != nil {
		t.Fatalf("add hosts entry second time: %v", err)
	}

	data, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Count(content, "demo.test") != 1 {
		t.Fatalf("expected single hosts entry for demo.test, got: %s", content)
	}
}

func TestRemoveHostsEntryPreservesOtherHosts(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "hosts")
	initial := "127.0.0.1 localhost demo.test\n127.0.0.1 keep.test\n"
	if err := os.WriteFile(tempFile, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	restore := engine.SetHostsFilePathForTest(tempFile)
	defer restore()

	if err := engine.RemoveHostsEntry("demo.test"); err != nil {
		t.Fatalf("remove hosts entry: %v", err)
	}

	data, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "demo.test") {
		t.Fatalf("did not expect demo.test after removal, got: %s", content)
	}
	if !strings.Contains(content, "keep.test") {
		t.Fatalf("expected keep.test to remain, got: %s", content)
	}
}
