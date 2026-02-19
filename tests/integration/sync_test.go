//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestSyncPlanMagento2CurrentFlags(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "sync-plan-current")

	result := env.RunGovard(
		t,
		projectDir,
		"sync",
		"--source", "dev",
		"--destination", "local",
		"--plan",
		"--file",
		"--include", "app/*",
		"--exclude", "vendor/",
	)
	result.AssertSuccess(t)

	out := result.Stdout
	assertContains(t, out, "Sync Plan Summary")
	assertContains(t, out, "scopes: files")
	assertContains(t, out, "include patterns: app/*")
	assertContains(t, out, "exclude patterns: vendor/")
	assertContains(t, out, "resume mode: enabled")
	assertContains(t, out, "--partial --append-verify")
	assertContains(t, out, "--include app/*")
	assertContains(t, out, "--exclude vendor/")
}

func TestSyncPlanNoResume(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "sync-plan-no-resume")

	result := env.RunGovard(
		t,
		projectDir,
		"sync",
		"--source", "dev",
		"--destination", "local",
		"--plan",
		"--file",
		"--no-resume",
	)
	result.AssertSuccess(t)

	out := result.Stdout
	assertContains(t, out, "resume mode: disabled")
	if strings.Contains(out, "--append-verify") {
		t.Fatalf("did not expect resume rsync flags when --no-resume is set, got: %s", out)
	}
}

func TestSyncPolicyProtectedDestinationBlocked(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-staging", "sync-policy-protected")

	result := env.RunGovard(t, projectDir, "sync", "--source", "local", "--destination", "prod", "--file")
	if result.Success() {
		t.Fatal("expected protected destination error")
	}
	assertContains(t, result.Stderr, "write-protected")
}

func TestSyncPolicyRemoteToRemoteBlocked(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-staging", "sync-policy-remote-remote")

	result := env.RunGovard(t, projectDir, "sync", "--source", "staging", "--destination", "prod", "--file")
	if result.Success() {
		t.Fatal("expected local<->remote validation error")
	}
	assertContains(t, result.Stderr, "local<->remote")
}

func TestSyncRuntimeRsyncInvocationWithShims(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "sync-runtime-rsync")
	shim := env.SetupRuntimeShims(t, map[string]int{
		"docker": 0,
		"ssh":    0,
		"rsync":  0,
	})

	result := env.RunGovardWithEnv(
		t,
		projectDir,
		shim.Env(),
		"sync",
		"--source", "dev",
		"--destination", "local",
		"--file",
	)
	result.AssertSuccess(t)

	logs := shim.ReadLog(t)
	assertContains(t, logs, "rsync|")
	assertContains(t, logs, "deploy@dev.example.com:/var/www/html/")
	assertContains(t, logs, "--partial --append-verify")
}

func TestSyncRuntimeDBPipelineWithShims(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "sync-runtime-db")
	shim := env.SetupRuntimeShims(t, map[string]int{
		"docker": 0,
		"ssh":    0,
		"rsync":  0,
	})

	result := env.RunGovardWithEnv(
		t,
		projectDir,
		shim.Env(),
		"sync",
		"--source", "dev",
		"--destination", "local",
		"--db",
	)
	result.AssertSuccess(t)

	logs := shim.ReadLog(t)
	assertContains(t, logs, "ssh|")
	assertContains(t, logs, "docker|exec -i -e MYSQL_PWD=magento m2-clone-basic-db-1 sh -lc if command -v mysql >/dev/null 2>&1; then DB_CLI=mysql; elif command -v mariadb >/dev/null 2>&1; then DB_CLI=mariadb;")
}

func assertContains(t *testing.T, haystack string, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected %q in output, got:\n%s", needle, haystack)
	}
}
