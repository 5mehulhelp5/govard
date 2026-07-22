//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtensionsAndCustomCommandsLifecycle(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "extensions-custom-m2")

	initResult := env.RunGovard(t, projectDir, "extensions", "init")
	initResult.AssertSuccess(t)
	assertContains(t, initResult.Stdout+initResult.Stderr, "Extension contract scaffolded")

	requiredFiles := []string{
		filepath.Join(projectDir, ".govard", ".govard.local.yml"),
		filepath.Join(projectDir, ".govard", "docker-compose.override.yml"),
		filepath.Join(projectDir, ".govard", "hooks", "pre_up.sh"),
		filepath.Join(projectDir, ".govard", "commands", "hello"),
	}
	for _, path := range requiredFiles {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected scaffold file %s: %v", path, err)
		}
	}

	listResult := env.RunGovard(t, projectDir, "custom", "list")
	listResult.AssertSuccess(t)
	assertContains(t, listResult.Stdout+listResult.Stderr, "hello")

	helloResult := env.RunGovard(t, projectDir, "custom", "hello", "one", "two")
	helloResult.AssertSuccess(t)
	assertContains(t, helloResult.Stdout, "Hello from .govard/commands/hello")
	assertContains(t, helloResult.Stdout, "Args: one two")

	fallbackPath := filepath.Join(projectDir, ".govard", "commands", "fallback")
	fallbackScript := "#!/usr/bin/env bash\nset -euo pipefail\necho \"fallback:$*\"\n"
	if err := os.WriteFile(fallbackPath, []byte(fallbackScript), 0o644); err != nil {
		t.Fatalf("failed to write fallback script: %v", err)
	}
	fallbackResult := env.RunGovard(t, projectDir, "custom", "fallback", "alpha")
	fallbackResult.AssertSuccess(t)
	assertContains(t, fallbackResult.Stdout, "fallback:alpha")

	helloPath := filepath.Join(projectDir, ".govard", "commands", "hello")
	customHello := "#!/usr/bin/env bash\nset -euo pipefail\necho 'custom hello preserved'\n"
	if err := os.WriteFile(helloPath, []byte(customHello), 0o755); err != nil {
		t.Fatalf("failed to customize hello command: %v", err)
	}

	withoutForce := env.RunGovard(t, projectDir, "extensions", "init")
	withoutForce.AssertSuccess(t)
	assertContains(t, withoutForce.Stdout+withoutForce.Stderr, "No files changed")

	helloContent, err := os.ReadFile(helloPath)
	if err != nil {
		t.Fatalf("failed to read hello script: %v", err)
	}
	if !strings.Contains(string(helloContent), "custom hello preserved") {
		t.Fatalf("expected custom hello script to be preserved without --force, got:\n%s", string(helloContent))
	}

	withForce := env.RunGovard(t, projectDir, "extensions", "init", "--force")
	withForce.AssertSuccess(t)
	assertContains(t, withForce.Stdout+withForce.Stderr, "Extension contract scaffolded")

	forcedHelloContent, err := os.ReadFile(helloPath)
	if err != nil {
		t.Fatalf("failed to read hello script after --force: %v", err)
	}
	assertContains(t, string(forcedHelloContent), "Hello from .govard/commands/hello")
}
