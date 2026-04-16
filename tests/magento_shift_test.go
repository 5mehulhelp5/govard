package tests

import (
	"encoding/json"
	"govard/internal/engine"
	"os"
	"path/filepath"
	"testing"
)

func TestMagentoProfileShiftDetection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "govard-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(cwd) }()

	// Mock Registry
	registryPath := filepath.Join(tmpDir, "projects.json")
	if err := os.Setenv("GOVARD_PROJECT_REGISTRY_PATH", registryPath); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Unsetenv("GOVARD_PROJECT_REGISTRY_PATH") }()

	registry := struct {
		Version  int                           `json:"version"`
		Projects []engine.ProjectRegistryEntry `json:"projects"`
	}{
		Version: 1,
		Projects: []engine.ProjectRegistryEntry{
			{
				Path:        tmpDir,
				ProjectName: "test-project",
				PHPVersion:  "8.2",
				Profile:     "default",
			},
		},
	}
	regData, _ := json.Marshal(registry)
	_ = os.WriteFile(registryPath, regData, 0644)

	// Case 1: No change
	config := engine.Config{}
	config.Stack.PHPVersion = "8.2"
	config.Profile = "default"

	shifted, reason := engine.CheckMagentoProfileShiftCleanupForTest(config)
	if shifted {
		t.Errorf("Expected no shift, got %v: %s", shifted, reason)
	}

	// Case 2: PHP Version change
	config.Stack.PHPVersion = "8.4"
	shifted, reason = engine.CheckMagentoProfileShiftCleanupForTest(config)
	if !shifted || reason != "PHP version changed: 8.2 -> 8.4" {
		t.Errorf("Expected PHP shift, got shifted=%v, reason=%s", shifted, reason)
	}

	// Case 3: Profile change
	config.Stack.PHPVersion = "8.2"
	config.Profile = "upgrade"
	shifted, reason = engine.CheckMagentoProfileShiftCleanupForTest(config)
	if !shifted || reason != "Profile changed: \"default\" -> \"upgrade\"" {
		t.Errorf("Expected Profile shift, got shifted=%v, reason=%s", shifted, reason)
	}
}
