//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
	"time"
)

func TestOpenCommandGuidanceTargets(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "open-m2")
	shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})

	shellResult := env.RunGovardWithEnv(t, projectDir, shim.Env(), "open", "shell")
	shellResult.AssertSuccess(t)
	assertContains(t, shim.ReadLog(t), "docker|exec")

	sftpResult := env.RunGovard(t, projectDir, "open", "sftp")
	sftpResult.AssertSuccess(t)
	assertContains(t, sftpResult.Stdout+sftpResult.Stderr, "SFTP is not supported for local target")

	unknownResult := env.RunGovard(t, projectDir, "open", "unknown-target")
	unknownResult.AssertExitCode(t, 1)
	assertContains(t, strings.ToLower(unknownResult.Stdout+unknownResult.Stderr), "unknown target")
}

func TestOpenAndShortcutBrowserCommands(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "open-shortcuts-m2")
	shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})
	browserCommand := browserOpenCommandNameForTest(t)
	installRuntimeCommandShim(t, shim, browserCommand, 0)

	commands := [][]string{
		{"open", "admin"},
		{"open", "db"},
		{"open", "db", "--client"},
		{"open", "elasticsearch"},
		{"open", "opensearch"},
		{"open", "mail"},
	}

	for _, args := range commands {
		result := env.RunGovardWithEnv(t, projectDir, shim.Env(), args...)
		result.AssertSuccess(t)
	}

	ok := WaitForCondition(t, 2*time.Second, 20*time.Millisecond, func() bool {
		logs := shim.ReadLog(t)
		return strings.Count(logs, browserCommand+"|") >= len(commands)
	})
	if !ok {
		t.Fatalf("expected %d %s invocations, got logs:\n%s", len(commands), browserCommand, shim.ReadLog(t))
	}

	logs := shim.ReadLog(t)
	assertContains(t, logs, browserCommand+"|https://m2-clone-basic.test/admin")
	assertContains(t, logs, "project=m2-clone-basic")
	assertContains(t, logs, "db=magento")
	assertContains(t, logs, browserCommand+"|mysql://magento:magento@127.0.0.1:3306/magento")
	assertContains(t, logs, browserCommand+"|https://elasticsearch.govard.test")
	assertContains(t, logs, browserCommand+"|https://opensearch.govard.test")
	assertContains(t, logs, browserCommand+"|https://mail.govard.test")
	if strings.Contains(logs, "ssh|") {
		t.Fatalf("did not expect ssh tunnel for default open targets, got logs:\n%s", logs)
	}
}

func TestOpenDBEnvironmentSelection(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "open-db-env-selection-m2")
	shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})
	browserCommand := browserOpenCommandNameForTest(t)
	installRuntimeCommandShim(t, shim, browserCommand, 0)

	localResult := env.RunGovardWithEnv(t, projectDir, shim.Env(), "open", "db", "-e", "local")
	localResult.AssertSuccess(t)

	logs := shim.ReadLog(t)
	assertContains(t, logs, "project=m2-clone-basic")
	assertContains(t, logs, "db=magento")
	if strings.Contains(logs, "ssh|") {
		t.Fatalf("did not expect ssh tunnel for local db open, got logs:\n%s", logs)
	}

	clientResult := env.RunGovardWithEnv(t, projectDir, shim.Env(), "open", "db", "-e", "local", "--client")
	clientResult.AssertSuccess(t)
	updatedLogs := shim.ReadLog(t)
	assertContains(t, updatedLogs, browserCommand+"|mysql://magento:magento@127.0.0.1:3306/magento")
	if strings.Contains(updatedLogs, "ssh|") {
		t.Fatalf("did not expect ssh tunnel for local db client open, got logs:\n%s", updatedLogs)
	}

	unknownResult := env.RunGovardWithEnv(t, projectDir, shim.Env(), "open", "db", "-e", "missing")
	unknownResult.AssertExitCode(t, 1)
	assertContains(t, strings.ToLower(unknownResult.Stdout+unknownResult.Stderr), "unknown remote environment")

	defaultResult := env.RunGovardWithEnv(t, projectDir, shim.Env(), "open", "db")
	defaultResult.AssertSuccess(t)
	defaultLogs := shim.ReadLog(t)
	assertContains(t, defaultLogs, "project=m2-clone-basic")
	assertContains(t, defaultLogs, "db=magento")
	if strings.Contains(defaultLogs, "ssh|") {
		t.Fatalf("did not expect ssh tunnel when open db uses default local env, got logs:\n%s", defaultLogs)
	}
}

func TestOpenTargetsRemoteEnvironment(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "open-remote-targets-m2")
	shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})
	browserCommand := browserOpenCommandNameForTest(t)
	installRuntimeCommandShim(t, shim, browserCommand, 0)

	adminResult := env.RunGovardWithEnv(t, projectDir, shim.Env(), "open", "admin", "-e", "dev")
	adminResult.AssertSuccess(t)

	sftpResult := env.RunGovardWithEnv(t, projectDir, shim.Env(), "open", "sftp", "-e", "dev")
	sftpResult.AssertSuccess(t)

	shellResult := env.RunGovardWithEnv(t, projectDir, shim.Env(), "open", "shell", "-e", "dev")
	shellResult.AssertSuccess(t)

	searchResult := env.RunGovardWithEnv(t, projectDir, shim.Env(), "open", "elasticsearch", "-e", "dev")
	searchResult.AssertExitCode(t, 1)
	assertContains(t, strings.ToLower(searchResult.Stdout+searchResult.Stderr), "not supported yet")

	logs := shim.ReadLog(t)
	assertContains(t, logs, browserCommand+"|https://dev.example.com/")
	assertContains(t, logs, browserCommand+"|sftp://deploy@dev.example.com:22/var/www/html")
	assertContains(t, logs, "ssh|")
}
