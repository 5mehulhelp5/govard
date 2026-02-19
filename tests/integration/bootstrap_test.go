//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestBootstrapValidationMatrix(t *testing.T) {
	env := NewTestEnvironment(t)

	t.Run("FreshAndCloneMutuallyExclusive", func(t *testing.T) {
		projectDir := env.CreateProjectFromFixture(t, "magento2/options-dev", "bootstrap-validate-fresh-clone")
		result := env.RunGovard(t, projectDir, "bootstrap", "--fresh", "--clone", "--skip-up")
		if result.Success() {
			t.Fatal("expected failure for --fresh + --clone")
		}
		assertBootstrapContains(t, result.Stderr, "--fresh and --clone cannot be used together")
	})

	t.Run("CodeOnlyRequiresClone", func(t *testing.T) {
		projectDir := env.CreateProjectFromFixture(t, "magento2/options-dev", "bootstrap-validate-code-only")
		result := env.RunGovard(t, projectDir, "bootstrap", "--clone=false", "--code-only", "--skip-up")
		if result.Success() {
			t.Fatal("expected failure for --code-only without --clone")
		}
		assertBootstrapContains(t, result.Stderr, "--code-only requires --clone")
	})

	t.Run("InvalidVersionRejected", func(t *testing.T) {
		projectDir := env.CreateProjectFromFixture(t, "magento2/options-dev", "bootstrap-validate-version")
		result := env.RunGovard(t, projectDir, "bootstrap", "--fresh", "--version", "1.0.0", "--skip-up")
		if result.Success() {
			t.Fatal("expected failure for invalid --version")
		}
		assertBootstrapContains(t, result.Stderr, "invalid --version value")
	})
}

func TestBootstrapCloneRequiresConfiguredRemote(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-dev", "bootstrap-no-remote")

	result := env.RunGovard(t, projectDir, "bootstrap", "--clone", "--environment", "dev", "--skip-up")
	if result.Success() {
		t.Fatal("expected missing remote error")
	}
	assertBootstrapContains(t, result.Stderr, "remote 'dev' is not configured")
}

func TestBootstrapCloneOrchestrationWithShims(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "bootstrap-clone-shims")
	shim := env.SetupRuntimeShims(t, map[string]int{
		"docker": 0,
		"ssh":    0,
		"rsync":  0,
	})

	result := env.RunGovardWithEnv(
		t,
		projectDir,
		shim.Env(),
		"bootstrap",
		"--clone",
		"--environment", "dev",
		"--skip-up",
		"--no-composer",
		"--no-db",
		"--no-media",
		"--no-admin",
	)
	result.AssertSuccess(t)

	logs := shim.ReadLog(t)
	assertBootstrapContains(t, logs, "ssh|")
	assertBootstrapContains(t, logs, "govard-remote-ok")
	assertBootstrapContains(t, logs, "rsync|")
	assertBootstrapContains(t, logs, "deploy@dev.example.com:/var/www/html/")
	assertBootstrapContains(t, logs, "docker|")
}

func TestBootstrapFreshOrchestrationWithShims(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-dev", "bootstrap-fresh-shims")
	shim := env.SetupRuntimeShims(t, map[string]int{
		"docker": 0,
		"ssh":    0,
		"rsync":  0,
	})

	result := env.RunGovardWithEnv(
		t,
		projectDir,
		shim.Env(),
		"bootstrap",
		"--fresh",
		"--skip-up",
		"--no-admin",
	)
	result.AssertSuccess(t)

	logs := shim.ReadLog(t)
	assertBootstrapContains(t, logs, "docker|exec")
	assertBootstrapContains(t, logs, "/tmp/govard-create-project")
	assertBootstrapContains(t, logs, "command -v rsync")
	assertBootstrapContains(t, logs, "setup:install")
}

func TestBootstrapCloneMediaSyncExcludesProductByDefault(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "bootstrap-media-default")
	shim := env.SetupRuntimeShims(t, map[string]int{
		"docker": 0,
		"ssh":    0,
		"rsync":  0,
	})

	result := env.RunGovardWithEnv(
		t,
		projectDir,
		shim.Env(),
		"bootstrap",
		"--clone",
		"--environment", "dev",
		"--skip-up",
		"--no-composer",
		"--no-db",
		"--no-admin",
	)
	result.AssertSuccess(t)

	logs := shim.ReadLog(t)
	assertBootstrapContains(t, logs, "--exclude catalog/product")
}

func TestBootstrapCloneMediaSyncIncludeProductFlag(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "bootstrap-media-include-product")
	shim := env.SetupRuntimeShims(t, map[string]int{
		"docker": 0,
		"ssh":    0,
		"rsync":  0,
	})

	result := env.RunGovardWithEnv(
		t,
		projectDir,
		shim.Env(),
		"bootstrap",
		"--clone",
		"--environment", "dev",
		"--skip-up",
		"--no-composer",
		"--no-db",
		"--no-admin",
		"--include-product",
	)
	result.AssertSuccess(t)

	logs := shim.ReadLog(t)
	if strings.Contains(logs, "--exclude catalog/product ") {
		t.Fatalf("did not expect catalog/product exclusion when --include-product is set, got:\n%s", logs)
	}
	assertBootstrapContains(t, logs, "--exclude catalog/product/cache")
}

func assertBootstrapContains(t *testing.T, haystack string, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected %q in output, got:\n%s", needle, haystack)
	}
}
