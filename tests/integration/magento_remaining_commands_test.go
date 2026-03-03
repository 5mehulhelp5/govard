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

	result := env.RunGovardWithEnv(t, projectDir, append(shim.Env(), "HOME="+homeDir), "doctor", "trust")
	result.AssertSuccess(t)

	certPath := filepath.Join(homeDir, ".govard", "ssl", "root.crt")
	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("expected extracted cert at %s: %v", certPath, err)
	}
	assertContains(t, string(certBytes), "BEGIN CERTIFICATE")

	logs := shim.ReadLog(t)
	assertContains(t, logs, "docker|cp govard-proxy-caddy:/data/caddy/pki/authorities/local/root.crt "+certPath)
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

	t.Run("DesktopLaunchesBinaryWithBackgroundFlag", func(t *testing.T) {
		shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})
		installRuntimeCommandShim(t, shim, "govard-desktop", 0)

		result := env.RunGovardWithEnv(t, projectDir, shim.Env(), "desktop", "--background")
		result.AssertSuccess(t)

		logs := shim.ReadLog(t)
		assertContains(t, logs, "govard-desktop|--background")
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
		10*time.Second,
		[]string{"GOVARD_SELF_UPDATE_CONFIRM=yes"},
		"self-update",
		"--version",
		"v1.0.2",
	)
	if errorsContain(result.Error, "context deadline exceeded") {
		t.Fatalf("self-update blocked even with GOVARD_SELF_UPDATE_CONFIRM=yes; output:\nstdout=%s\nstderr=%s", result.Stdout, result.Stderr)
	}
	result.AssertSuccess(t)
	assertContains(t, result.Stdout+result.Stderr, "Successfully updated Govard to")
}

func runGovardWithTimeout(t *testing.T, env *TestEnvironment, projectDir string, timeout time.Duration, extraEnv []string, args ...string) *CommandResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	isolatedBinary := filepath.Join(t.TempDir(), "govard-self-update-test")
	copyBinaryForTest(t, env.BinaryPath, isolatedBinary)

	cmd := exec.CommandContext(ctx, isolatedBinary, args...)
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

func copyBinaryForTest(t *testing.T, src string, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("failed to read binary %s: %v", src, err)
	}
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		t.Fatalf("failed to write isolated binary %s: %v", dst, err)
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
  cat > "$dest" <<'EOF_CERT'
-----BEGIN CERTIFICATE-----
MIIDEzCCAfugAwIBAgIUciZkq4eXPE7ktpQ5jc8mUERKBe4wDQYJKoZIhvcNAQEL
BQAwGTEXMBUGA1UEAwwOR292YXJkIFRlc3QgQ0EwHhcNMjYwMzAzMDMxNTE5WhcN
MzYwMjI5MDMxNTE5WjAZMRcwFQYDVQQDDA5Hb3ZhcmQgVGVzdCBDQTCCASIwDQYJ
KoZIhvcNAQEBBQADggEPADCCAQoCggEBAKBpdvlGnEsieYi5mj/9dDPvT5Fkwbir
UvPmS/9ekFsAXNaqD6/XmM1vXHsFDf1P9OVPwnTkicq+iVShuekOMSzOI+ZOBG+C
GdZWnXUUny3wQBxAJLCcqqlp9aA1Y+XSn47TWPWmIAWNddxr0mvn2BloW4gDssss
g4egYlcbHHe7JxQZUEcHLm49uuE/o87y5KPtwdVi/B7pgmOh75+2N4XcxHp+rc0l
LHFZ+5QPZiW9N8Nl60N+1Wskx7wh7D/mvs7HUUEdFZ1f9WJQLAbEZr8kCROaNlUb
/58q7txwzvrb0pwlApkB6bi+gzOYGHPE3nRH6shTnSKexoas8aDlI2ECAwEAAaNT
MFEwHQYDVR0OBBYEFMSQLC8cLtJkqYX/sNhyHviQcyLsMB8GA1UdIwQYMBaAFMSQ
LC8cLtJkqYX/sNhyHviQcyLsMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQEL
BQADggEBADP8g7Znris9/eFl8+Oclk/Zj9b0vkMTabZCj4wFxzbAmKA67OlwlnqU
IKA+pwBo1lNXpUWUQ06O/9eaNeuypWSqpFkNMqu91AD6Y6XWghxFBZ61bBIX3S1/
Kib0XGGkTTRkHjFyRcyhE9NmkSrM0MN9pvB5ACUvge8AEmWqG93paBVjuMTWgw1Z
6Gm/ewY5+pnMQvJEqyPAPVQS1kQ9UiL4SLi1EiM57/8Vot84u5lmaYn0jsZe0KTS
y8GOGWKnl7a+ZYmEoX841u9GcWl2EWAWmoIuE75YxmBDoUj5v8qD/LsJ2qOBckJu
tPqeILVmoltqkVAloHKAMzbHtwE7J2o=
-----END CERTIFICATE-----
EOF_CERT
fi
exit "${GOVARD_TEST_EXIT_DOCKER:-0}"
`
	path := filepath.Join(shims.Dir, "docker")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to install docker trust shim: %v", err)
	}
}
