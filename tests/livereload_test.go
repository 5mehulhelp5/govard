package tests

import (
	"strings"
	"testing"

	"govard/internal/engine"
)

func TestRenderBlueprintLiveReloadDisabledNoWatcher(t *testing.T) {
	content := renderComposeWithConfig(t, engine.Config{
		ProjectName: "lr-off",
		Framework:   "magento2",
		Domain:      "lr-off.test",
		Stack: engine.Stack{
			Features: engine.Features{LiveReload: false},
		},
	})

	if strings.Contains(content, "watcher:") {
		t.Fatal("expected no watcher service when LiveReload is disabled")
	}
}

func TestRenderBlueprintLiveReloadEnabledHasWatcher(t *testing.T) {
	content := renderComposeWithConfig(t, engine.Config{
		ProjectName: "lr-on",
		Framework:   "magento2",
		Domain:      "lr-on.test",
		Stack: engine.Stack{
			NodeVersion: "20",
			Features:    engine.Features{LiveReload: true},
		},
	})

	if !strings.Contains(content, "watcher:") {
		t.Fatal("expected watcher service when LiveReload is enabled")
	}
	if !strings.Contains(content, "node:20-alpine") {
		t.Fatal("expected node image with configured version")
	}
}

func TestLiveReloadConfigParseDefault(t *testing.T) {
	cfg := engine.Config{}
	if cfg.Stack.Features.LiveReload {
		t.Fatal("expected LiveReload to default to false")
	}
}
