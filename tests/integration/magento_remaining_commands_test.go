//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTrustCommandWithShims(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "trust-m2")
	shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})

	installTrustDockerShim(t, shim)
	installRuntimeCommandShim(t, shim, "sudo", 0)

	homeDir := filepath.Join(projectDir, ".home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("failed to create home dir: %v", err)
	}

	result := env.RunGovardWithEnv(t, projectDir, append(shim.Env(), "HOME="+homeDir), "trust")
	result.AssertSuccess(t)

	certPath := filepath.Join(homeDir, ".govard", "ssl", "root.crt")
	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("expected extracted cert at %s: %v", certPath, err)
	}
	assertContains(t, string(certBytes), "mock-cert")

	logs := shim.ReadLog(t)
	assertContains(t, logs, "docker|cp proxy-caddy-1:/data/caddy/pki/authorities/local/root.crt "+certPath)
	assertContains(t, logs, "sudo|cp "+certPath+" /usr/local/share/ca-certificates/govard.crt")
	assertContains(t, logs, "sudo|update-ca-certificates")
}

func TestDesktopCommandRuntimePaths(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "desktop-m2")

	t.Run("DesktopLaunchesBinaryFromPATH", func(t *testing.T) {
		shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})
		installRuntimeCommandShim(t, shim, "govard-desktop", 0)

		result := env.RunGovardWithEnv(t, projectDir, shim.Env(), "desktop")
		result.AssertSuccess(t)

		logs := shim.ReadLog(t)
		assertContains(t, logs, "govard-desktop|")
	})

	t.Run("DesktopDevUsesWailsWhenAvailable", func(t *testing.T) {
		shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})
		installRuntimeCommandShim(t, shim, "wails", 0)

		result := env.RunGovardWithEnv(t, projectDir, shim.Env(), "desktop", "--dev")
		result.AssertSuccess(t)

		logs := shim.ReadLog(t)
		assertContains(t, logs, "wails|dev -tags desktop")
	})
}

func TestSelfUpdateCommandDoesNotBlockInNonInteractiveMode(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "self-update-m2")

	result := runGovardWithTimeout(t, env, projectDir, 2*time.Second, nil, "self-update")
	if errorsContain(result.Error, "context deadline exceeded") {
		t.Fatalf("self-update blocked in non-interactive mode; output:\nstdout=%s\nstderr=%s", result.Stdout, result.Stderr)
	}
	result.AssertSuccess(t)
	assertContains(t, result.Stdout+result.Stderr, "Update cancelled.")
}

func TestSelfUpdateAutoConfirmViaEnv(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "self-update-confirm-m2")

	result := runGovardWithTimeout(
		t,
		env,
		projectDir,
		2*time.Second,
		[]string{"GOVARD_SELF_UPDATE_CONFIRM=yes"},
		"self-update",
	)
	if errorsContain(result.Error, "context deadline exceeded") {
		t.Fatalf("self-update blocked even with GOVARD_SELF_UPDATE_CONFIRM=yes; output:\nstdout=%s\nstderr=%s", result.Stdout, result.Stderr)
	}
	result.AssertSuccess(t)
	assertContains(t, result.Stdout+result.Stderr, "Successfully updated to the latest version")
}

func runGovardWithTimeout(t *testing.T, env *TestEnvironment, projectDir string, timeout time.Duration, extraEnv []string, args ...string) *CommandResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, env.BinaryPath, args...)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), extraEnv...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	started := time.Now()
	err := cmd.Run()
	duration := time.Since(started)

	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	return &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
		Error:    err,
	}
}

func errorsContain(err error, needle string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), needle)
}

func installTrustDockerShim(t *testing.T, shims *RuntimeShims) {
	t.Helper()
	script := `#!/bin/sh
set -eu
log="${GOVARD_TEST_RUNTIME_LOG:-}"
if [ -n "$log" ]; then
  printf '%s|%s\n' "docker" "$*" >> "$log"
fi
if [ "$#" -ge 3 ] && [ "$1" = "cp" ]; then
  dest="$3"
  mkdir -p "$(dirname "$dest")"
  printf '%s\n' "mock-cert" > "$dest"
fi
exit "${GOVARD_TEST_EXIT_DOCKER:-0}"
`
	path := filepath.Join(shims.Dir, "docker")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to install docker trust shim: %v", err)
	}
}
