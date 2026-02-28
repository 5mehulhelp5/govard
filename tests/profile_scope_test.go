package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/engine"
)

// --- Config Layer Tests ---

func TestProfileConfigLayerMerge(t *testing.T) {
	tempDir := t.TempDir()

	base := `project_name: myproj
domain: myproj.test
framework: magento2
stack:
  php_version: "8.1"
  services:
    search: elasticsearch
`
	profile := `stack:
  php_version: "8.2"
  services:
    search: opensearch
`
	if err := os.WriteFile(filepath.Join(tempDir, ".govard.yml"), []byte(base), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, ".govard.upgrade.yml"), []byte(profile), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, loaded, err := engine.LoadConfigFromDirWithProfile(tempDir, true, "upgrade")
	if err != nil {
		t.Fatalf("load config with profile: %v", err)
	}

	if cfg.Stack.PHPVersion != "8.2" {
		t.Fatalf("expected profile php 8.2, got %s", cfg.Stack.PHPVersion)
	}
	if cfg.Stack.Services.Search != "opensearch" {
		t.Fatalf("expected profile search opensearch, got %s", cfg.Stack.Services.Search)
	}

	// Check that profile layer was loaded
	foundProfile := false
	for _, path := range loaded {
		if strings.Contains(path, ".govard.upgrade.yml") {
			foundProfile = true
			break
		}
	}
	if !foundProfile {
		t.Fatal("expected profile layer in loaded paths")
	}
}

func TestProfileLocalOverridesProfile(t *testing.T) {
	tempDir := t.TempDir()

	base := `project_name: myproj
domain: myproj.test
framework: magento2
stack:
  php_version: "8.1"
`
	profile := `stack:
  php_version: "8.2"
`
	local := `stack:
  php_version: "8.3"
`

	if err := os.WriteFile(filepath.Join(tempDir, ".govard.yml"), []byte(base), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, ".govard.upgrade.yml"), []byte(profile), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, ".govard.local.yml"), []byte(local), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, _, err := engine.LoadConfigFromDirWithProfile(tempDir, true, "upgrade")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	// Local override should win over profile
	if cfg.Stack.PHPVersion != "8.3" {
		t.Fatalf("expected local override 8.3 to win over profile 8.2, got %s", cfg.Stack.PHPVersion)
	}
}

func TestProfileEmptyStringFallsBackToDefault(t *testing.T) {
	tempDir := t.TempDir()
	base := `project_name: fallback
domain: fallback.test
framework: magento2
stack:
  php_version: "8.1"
`
	if err := os.WriteFile(filepath.Join(tempDir, ".govard.yml"), []byte(base), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, _, err := engine.LoadConfigFromDirWithProfile(tempDir, true, "")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Stack.PHPVersion != "8.1" {
		t.Fatalf("expected base php 8.1, got %s", cfg.Stack.PHPVersion)
	}
}

// --- Compose Path Tests ---

func TestComposeFilePathWithProfile(t *testing.T) {
	path1 := engine.ComposeFilePathWithProfile("/tmp/myproj", "myproj", "")
	path2 := engine.ComposeFilePathWithProfile("/tmp/myproj", "myproj", "upgrade")

	if path1 == path2 {
		t.Fatal("expected different compose paths for different profiles")
	}
	if !strings.Contains(path2, "upgrade") {
		t.Fatal("expected profile name in compose path")
	}
}

func TestComposeFilePathWithProfileDefaultMatchesOriginal(t *testing.T) {
	original := engine.ComposeFilePath("/tmp/myproj", "myproj")
	withEmpty := engine.ComposeFilePathWithProfile("/tmp/myproj", "myproj", "")

	if original != withEmpty {
		t.Fatalf("expected empty profile to produce same path as original.\noriginal: %s\nwithEmpty: %s", original, withEmpty)
	}
}
func TestRenderBlueprintProfileVolumeIsolation(t *testing.T) {
	content1 := renderComposeWithConfig(t, engine.Config{
		ProjectName: "vol-iso",
		Framework:   "magento2",
		Domain:      "vol-iso.test",
	})
	if !strings.Contains(content1, "db-data:") {
		t.Fatal("expected default db-data volume")
	}

	content2 := renderComposeWithConfig(t, engine.Config{
		ProjectName: "vol-iso",
		Framework:   "magento2",
		Domain:      "vol-iso.test",
		Profile:     "upgrade",
	})

	if !strings.Contains(content2, "db-data-upgrade:") {
		t.Fatal("expected suffixed db-data-upgrade volume")
	}
	if strings.Contains(content2, " db-data:") {
		t.Fatal("expected no original db-data volume when profile is active")
	}
}
