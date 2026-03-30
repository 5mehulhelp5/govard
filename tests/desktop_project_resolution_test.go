package tests

import (
	"govard/internal/desktop"
	"govard/internal/engine"
	"os"
	"path/filepath"
	"testing"
)

func TestDesktopProjectResolution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "govard-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy project directory
	projectPath := filepath.Join(tmpDir, "sample-project")
	err = os.MkdirAll(projectPath, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(projectPath, engine.BaseConfigFile), []byte("project_name: sample-project\ndomain: sample-project.test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Set up mock registry
	registryPath := filepath.Join(tmpDir, "projects.json")
	err = os.Setenv(engine.ProjectRegistryPathEnvVar, registryPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Unsetenv(engine.ProjectRegistryPathEnvVar)
	}()

	entry := engine.ProjectRegistryEntry{
		Path:        projectPath,
		ProjectName: "sample-project",
		Domain:      "sample-project.test",
	}
	err = engine.UpsertProjectRegistryEntry(entry)
	if err != nil {
		t.Fatalf("failed to upsert registry entry: %v", err)
	}

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"Exact Domain Match", "sample-project.test", projectPath},
		{"Exact Name Match", "sample-project", projectPath},
		{"Fuzzy Search", "sample", projectPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := desktop.ResolveProjectRootForRemotesForTest(tt.query)
			if err != nil {
				t.Fatalf("failed to resolve %q: %v", tt.query, err)
			}
			if filepath.Clean(got) != filepath.Clean(tt.want) {
				t.Errorf("ResolveProjectRootForRemotes(%q) = %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}
