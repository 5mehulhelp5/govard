//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"

	"govard/internal/engine"
)

func TestConfigCommandGetSetAndUnknownKey(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "config-cmd-m2")

	getProject := env.RunGovard(t, projectDir, "config", "get", "project_name")
	getProject.AssertSuccess(t)
	if strings.TrimSpace(getProject.Stdout) != "m2-clone-basic" {
		t.Fatalf("expected project_name m2-clone-basic, got %q", strings.TrimSpace(getProject.Stdout))
	}

	setDomain := env.RunGovard(t, projectDir, "config", "set", "domain", "m2-config-updated.test")
	setDomain.AssertSuccess(t)
	assertContains(t, setDomain.Stdout+setDomain.Stderr, "Config updated: domain = m2-config-updated.test")

	getDomain := env.RunGovard(t, projectDir, "config", "get", "domain")
	getDomain.AssertSuccess(t)
	if strings.TrimSpace(getDomain.Stdout) != "m2-config-updated.test" {
		t.Fatalf("expected updated domain in config get, got %q", strings.TrimSpace(getDomain.Stdout))
	}

	setCache := env.RunGovard(t, projectDir, "config", "set", "services.cache", "redis")
	setCache.AssertSuccess(t)
	assertContains(t, setCache.Stdout+setCache.Stderr, "Config updated: services.cache = redis")

	config, _, err := engine.LoadConfigFromDir(projectDir, true)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}
	if config.Domain != "m2-config-updated.test" {
		t.Fatalf("expected config domain m2-config-updated.test, got %q", config.Domain)
	}
	if config.Stack.Services.Cache != "redis" {
		t.Fatalf("expected cache service redis, got %q", config.Stack.Services.Cache)
	}

	unknown := env.RunGovard(t, projectDir, "config", "set", "unknown.key", "x")
	unknown.AssertExitCode(t, 1)
	assertContains(t, strings.ToLower(unknown.Stdout+unknown.Stderr), "unknown config key: unknown.key")
}

func TestConfigureMagentoWithShims(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "configure-m2-shims")
	shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})

	result := env.RunGovardWithEnv(t, projectDir, shim.Env(), "config", "auto")
	result.AssertSuccess(t)

	output := strings.ToLower(result.Stdout + result.Stderr)
	assertContains(t, output, "auto-configuration")
}
