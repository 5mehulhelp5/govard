package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/engine"
)

func TestIsolationModeDefaultFalse(t *testing.T) {
	tempDir := t.TempDir()
	base := `project_name: iso-default
domain: iso-default.test
framework: magento2
`
	if err := os.WriteFile(filepath.Join(tempDir, ".govard.yml"), []byte(base), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, _, err := engine.LoadConfigFromDir(tempDir, true)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Stack.Features.Isolated {
		t.Fatal("expected isolated to default to false")
	}
}

func TestIsolationModeEnabledFromConfig(t *testing.T) {
	tempDir := t.TempDir()
	base := `project_name: iso-enabled
domain: iso-enabled.test
framework: magento2
stack:
  features:
    isolated: true
`
	if err := os.WriteFile(filepath.Join(tempDir, ".govard.yml"), []byte(base), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, _, err := engine.LoadConfigFromDir(tempDir, true)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if !cfg.Stack.Features.Isolated {
		t.Fatal("expected isolated to be true")
	}
}

func TestIsolationModeEnabledFromLocalOverride(t *testing.T) {
	tempDir := t.TempDir()
	base := `project_name: iso-local
domain: iso-local.test
framework: magento2
`
	local := `stack:
  features:
    isolated: true
`
	if err := os.WriteFile(filepath.Join(tempDir, ".govard.yml"), []byte(base), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, ".govard.local.yml"), []byte(local), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, _, err := engine.LoadConfigFromDir(tempDir, true)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if !cfg.Stack.Features.Isolated {
		t.Fatal("expected isolated to be true from local override")
	}
}

func TestRenderBlueprintIsolationModeDisabled(t *testing.T) {
	content := renderComposeWithConfig(t, engine.Config{
		ProjectName: "iso-off",
		Framework:   "magento2",
		Domain:      "iso-off.test",
		Stack: engine.Stack{
			Features: engine.Features{Isolated: false},
		},
	})

	if strings.Contains(content, "internal: true") {
		t.Fatal("expected no internal: true when isolation is disabled")
	}
}

func TestRenderBlueprintIsolationModeEnabled(t *testing.T) {
	content := renderComposeWithConfig(t, engine.Config{
		ProjectName: "iso-on",
		Framework:   "magento2",
		Domain:      "iso-on.test",
		Stack: engine.Stack{
			Features: engine.Features{Isolated: true},
		},
	})

	if !strings.Contains(content, "internal: true") {
		t.Fatal("expected internal: true when isolation is enabled")
	}
}
