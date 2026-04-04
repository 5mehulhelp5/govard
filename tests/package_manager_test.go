package tests

import (
	"os"
	"path/filepath"
	"testing"

	"govard/internal/engine"
)

func TestResolveNodePackageManagerPrefersPackageJSONDeclaration(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package-lock.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"packageManager":"pnpm@10.11.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	pm := engine.ResolveNodePackageManager(root)

	if pm != "pnpm" {
		t.Fatalf("expected pnpm from packageManager declaration, got %s", pm)
	}
}

func TestResolveNodePackageManagerUsesWorkspaceDetection(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "pnpm-workspace.yaml"), []byte("packages:\n  - ."), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package-lock.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	pm := engine.ResolveNodePackageManager(root)
	if pm != "pnpm" {
		t.Fatalf("expected pnpm from workspace detection, got %s", pm)
	}
}

func TestResolveNodePackageManagerUsesLockfileDetection(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "pnpm-lock.yaml"), []byte("lockfileVersion: '9.0'"), 0644); err != nil {
		t.Fatal(err)
	}

	pm := engine.ResolveNodePackageManager(root)
	if pm != "pnpm" {
		t.Fatalf("expected pnpm from lockfile detection, got %s", pm)
	}
}
