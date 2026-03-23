package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"govard/internal/cmd"
)

func TestOpenCommandRuntimeLocalTargetsUseOpenerShim(t *testing.T) {
	resetOpenFlagsForRuntimeTest(t)

	openerBinary, ok := openRuntimeOpenerBinary()
	if !ok {
		t.Skipf("open command runtime shim is not supported on %s", runtime.GOOS)
	}

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: laravel
`)

	shimDir := t.TempDir()
	logPath := filepath.Join(shimDir, "open.log")
	installOpenRuntimeShim(t, shimDir, openerBinary)
	t.Setenv("OPEN_RUNTIME_LOG", logPath)
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	root.SetArgs([]string{"open", "admin"})
	if err := root.Execute(); err != nil {
		t.Fatalf("open admin failed: %v", err)
	}

	root.SetArgs([]string{"open", "db", "--pma"})
	if err := root.Execute(); err != nil {
		t.Fatalf("open db --pma failed: %v", err)
	}

	logs := readRuntimeLog(t, logPath)
	if !strings.Contains(logs, "https://sample.test/admin") {
		t.Fatalf("missing admin URL in opener log:\n%s", logs)
	}
	if !strings.Contains(logs, "project=sample-project") || !strings.Contains(logs, "db=laravel") {
		t.Fatalf("missing db/project params in PMA URL in opener log:\n%s", logs)
	}
}

func TestOpenCommandRuntimeDBClientUsesDockerAndOpenerShims(t *testing.T) {
	resetOpenFlagsForRuntimeTest(t)

	openerBinary, ok := openRuntimeOpenerBinary()
	if !ok {
		t.Skipf("open command runtime shim is not supported on %s", runtime.GOOS)
	}

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: laravel
`)

	shimDir := t.TempDir()
	openLogPath := filepath.Join(shimDir, "open.log")
	dockerLogPath := filepath.Join(shimDir, "docker.log")
	installOpenRuntimeShim(t, shimDir, openerBinary)
	installDBRuntimeDockerShim(t, shimDir)
	t.Setenv("OPEN_RUNTIME_LOG", openLogPath)
	t.Setenv("DB_RUNTIME_DOCKER_LOG", dockerLogPath)
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"open", "db", "--client"})
	if err := root.Execute(); err != nil {
		t.Fatalf("open db --client failed: %v", err)
	}

	openLogs := readRuntimeLog(t, openLogPath)
	if !strings.Contains(openLogs, "mysql://dbuser:dbpass@127.0.0.1:3306/dbname") {
		t.Fatalf("missing local DB URL in opener log:\n%s", openLogs)
	}

	dockerLogs := readRuntimeLog(t, dockerLogPath)
	if !strings.Contains(dockerLogs, "docker|inspect -f {{.State.Running}} sample-project-db-1") {
		t.Fatalf("missing docker running check in log:\n%s", dockerLogs)
	}
	if !strings.Contains(dockerLogs, "docker|inspect -f {{range .Config.Env}}{{println .}}{{end}} sample-project-db-1") {
		t.Fatalf("missing docker env inspect in log:\n%s", dockerLogs)
	}
}

func TestOpenCommandRuntimeRejectsRemoteMail(t *testing.T) {
	resetOpenFlagsForRuntimeTest(t)

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: laravel
remotes:
  dev:
    host: dev.example.com
    user: deploy
    path: /srv/www/app
`)

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"open", "mail", "--environment", "dev"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected remote open mail to fail")
	}
	if !strings.Contains(err.Error(), "not supported yet") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenCommandRuntimeRejectsUnknownTarget(t *testing.T) {
	resetOpenFlagsForRuntimeTest(t)

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: laravel
`)

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"open", "unknown-target"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected unknown open target to fail")
	}
	if !strings.Contains(err.Error(), "unknown target") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func openRuntimeOpenerBinary() (string, bool) {
	switch runtime.GOOS {
	case "linux":
		return "xdg-open", true
	case "darwin":
		return "open", true
	default:
		return "", false
	}
}

func installOpenRuntimeShim(t *testing.T, shimDir string, binaryName string) {
	t.Helper()
	script := `#!/bin/sh
set -eu
log="${OPEN_RUNTIME_LOG:-}"
if [ -n "$log" ]; then
  printf '%s|%s\n' "$0" "$*" >> "$log"
fi
exit 0
`
	path := filepath.Join(shimDir, binaryName)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write open shim: %v", err)
	}
}

func installDBRuntimeDockerShim(t *testing.T, shimDir string) {
	t.Helper()
	script := `#!/bin/sh
set -eu
log="${DB_RUNTIME_DOCKER_LOG:-}"
if [ -n "$log" ]; then
  printf 'docker|%s\n' "$*" >> "$log"
fi
if [ "${1:-}" = "inspect" ] && [ "${2:-}" = "-f" ]; then
  case "${3:-}" in
    "{{.State.Running}}")
      echo "true"
      exit 0
      ;;
    "{{range .Config.Env}}{{println .}}{{end}}")
      cat <<'EOF'
MYSQL_USER=dbuser
MYSQL_PASSWORD=dbpass
MYSQL_DATABASE=dbname
EOF
      exit 0
      ;;
  esac
fi
exit 0
`
	path := filepath.Join(shimDir, "docker")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write docker shim: %v", err)
	}
}

func writeRuntimeConfig(t *testing.T, dir string, content string) {
	t.Helper()
	path := filepath.Join(dir, ".govard.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config %s: %v", path, err)
	}
}

func readRuntimeLog(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(bytes.TrimSpace(data))
}

func TestOpenCommandRuntimeSupportsExplicitEnvironmentAliasLookup(t *testing.T) {
	resetOpenFlagsForRuntimeTest(t)

	openerBinary, ok := openRuntimeOpenerBinary()
	if !ok {
		t.Skipf("open command runtime shim is not supported on %s", runtime.GOOS)
	}

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: laravel
remotes:
  stg:
    host: staging.example.com
    user: deploy
    path: /srv/www/app
`)

	shimDir := t.TempDir()
	logPath := filepath.Join(shimDir, "open.log")
	installOpenRuntimeShim(t, shimDir, openerBinary)
	t.Setenv("OPEN_RUNTIME_LOG", logPath)
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"open", "admin", "-e", "staging"})
	if err := root.Execute(); err != nil {
		t.Fatalf("open admin -e staging failed: %v", err)
	}

	logs := readRuntimeLog(t, logPath)
	if !strings.Contains(logs, "https://staging.example.com/admin") {
		t.Fatalf("missing remote admin URL in opener log:\n%s", logs)
	}
}

func TestOpenCommandRuntimeDBClientRejectsUnknownRemote(t *testing.T) {
	resetOpenFlagsForRuntimeTest(t)

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: laravel
`)

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"open", "db", "--client", "--environment", "missing"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected unknown remote environment error")
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("unknown remote environment %q", "missing")) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func resetOpenFlagsForRuntimeTest(t *testing.T) {
	t.Helper()
	cmd.ResetOpenFlagsForTest()
	t.Cleanup(cmd.ResetOpenFlagsForTest)
}
