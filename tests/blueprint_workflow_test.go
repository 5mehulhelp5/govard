package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"govard/internal/engine"

	"gopkg.in/yaml.v3"
)

func TestFullSetupLogic(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "full-setup-*")
	defer os.RemoveAll(tempDir)

	projectName := filepath.Base(tempDir)

	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..")
	blueprintsDir := filepath.Join(projectRoot, "internal", "blueprints", "files")

	destBlueprintsDir := filepath.Join(tempDir, "blueprints")
	if err := copyDir(blueprintsDir, destBlueprintsDir); err != nil {
		t.Fatalf("Failed to copy blueprints: %v", err)
	}

	config := engine.Config{
		ProjectName: projectName,
		Framework:   "magento2",
		Domain:      projectName + ".test",
		Stack: engine.Stack{
			PHPVersion: "8.1",
			WebServer:  "nginx",
			Features: engine.Features{
				Varnish: true,
			},
		},
	}

	data, _ := yaml.Marshal(&config)
	_ = os.WriteFile(filepath.Join(tempDir, "govard.yml"), data, 0644)

	err := engine.RenderBlueprint(tempDir, config)
	if err != nil {
		t.Fatalf("Failed to render blueprint: %v", err)
	}

	renderPath := engine.ComposeFilePath(tempDir, config.ProjectName)
	rendered, _ := os.ReadFile(renderPath)

	if !strings.Contains(string(rendered), "govard-proxy") {
		t.Error("Rendered compose file missing govard-proxy network")
	}

	if !strings.Contains(string(rendered), "external: true") {
		t.Error("govard-proxy network should be marked as external")
	}
}

// TestRenderBlueprintReRendersWhenComposeFileMissing is a regression test for the bug where
// RenderBlueprintWithProfile would skip rendering (due to a matching hash) even when the
// rendered compose file had been deleted from disk — causing `govard env up` to fail with
// "no such file or directory" in the Start stage.
func TestRenderBlueprintReRendersWhenComposeFileMissing(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "render-missing-compose-*")
	defer os.RemoveAll(tempDir)

	config := engine.Config{
		ProjectName: "sample-project",
		Framework:   "magento2",
		Domain:      "sample-project.test",
		Stack: engine.Stack{
			PHPVersion: "8.3",
		},
	}

	// First render — produces compose file + hash.
	if err := engine.RenderBlueprint(tempDir, config); err != nil {
		t.Fatalf("first render failed: %v", err)
	}

	composePath := engine.ComposeFilePath(tempDir, config.ProjectName)
	if _, err := os.Stat(composePath); err != nil {
		t.Fatalf("compose file missing after first render: %v", err)
	}

	// Simulate the compose file being deleted (e.g. manual cleanup, tmp-dir wipe).
	if err := os.Remove(composePath); err != nil {
		t.Fatalf("could not remove compose file: %v", err)
	}

	// Second render — config unchanged, so hash would normally cause a skip.
	// This must NOT skip, because the compose file is gone.
	if err := engine.RenderBlueprint(tempDir, config); err != nil {
		t.Fatalf("second render failed: %v", err)
	}

	if _, err := os.Stat(composePath); err != nil {
		t.Errorf("compose file still missing after second render (hash-skip regression): %v", err)
	}
}

func TestRenderBlueprintReRendersWhenBlueprintContentsChange(t *testing.T) {
	tempDir := t.TempDir()
	setTestGovardHome(t, tempDir)

	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..")
	blueprintsDir := filepath.Join(projectRoot, "internal", "blueprints", "files")

	destBlueprintsDir := filepath.Join(tempDir, "blueprints")
	if err := copyDir(blueprintsDir, destBlueprintsDir); err != nil {
		t.Fatalf("Failed to copy blueprints: %v", err)
	}

	config := engine.Config{
		ProjectName: "sample-project",
		Framework:   "custom",
		Domain:      "sample-project.test",
		Stack: engine.Stack{
			PHPVersion: "8.4",
			Services: engine.Services{
				WebServer: "nginx",
				Search:    "none",
				Cache:     "none",
				Queue:     "none",
			},
		},
	}

	if err := engine.RenderBlueprint(tempDir, config); err != nil {
		t.Fatalf("first render failed: %v", err)
	}

	composePath := engine.ComposeFilePath(tempDir, config.ProjectName)
	before, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("read first compose file: %v", err)
	}
	if !strings.Contains(string(before), "govard-net:") {
		t.Fatalf("expected initial compose output to contain govard-net network, got:\n%s", string(before))
	}

	basePath := filepath.Join(destBlueprintsDir, "includes", "base.yml")
	baseContent, err := os.ReadFile(basePath)
	if err != nil {
		t.Fatalf("read copied base blueprint: %v", err)
	}
	updated := strings.Replace(string(baseContent), "govard-net", "govard-net-reloaded", 1)
	if updated == string(baseContent) {
		t.Fatal("expected blueprint content replacement to change base.yml")
	}
	if err := os.WriteFile(basePath, []byte(updated), 0o644); err != nil {
		t.Fatalf("write modified base blueprint: %v", err)
	}

	if err := engine.RenderBlueprint(tempDir, config); err != nil {
		t.Fatalf("second render failed: %v", err)
	}

	after, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("read second compose file: %v", err)
	}
	if !strings.Contains(string(after), "- govard-net-reloaded") {
		t.Fatalf("expected compose output to re-render after blueprint change, got:\n%s", string(after))
	}
}
