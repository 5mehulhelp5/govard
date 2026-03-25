package tests

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/cmd"
)

func TestSyncCommandRuntimeFileOnlyUsesRsyncShim(t *testing.T) {
	resetSyncFlagsForRuntimeTest(t)

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: laravel
remotes:
  staging:
    host: staging.example.com
    user: deploy
    path: /srv/www/app
`)

	shimDir := t.TempDir()
	logPath := filepath.Join(shimDir, "rsync.log")
	installSyncRuntimeRsyncShim(t, shimDir)
	t.Setenv("SYNC_RUNTIME_RSYNC_LOG", logPath)
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{
		"sync",
		"--yes",
		"--source", "staging",
		"--destination", "local",
		"--file",
		"--path", "app/code",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("sync runtime failed: %v", err)
	}

	logs := readRuntimeLog(t, logPath)
	if !strings.Contains(logs, "rsync|-avz") {
		t.Fatalf("missing rsync invocation in log:\n%s", logs)
	}
	if !strings.Contains(logs, "--partial --append-verify") {
		t.Fatalf("expected resumable flags in rsync invocation:\n%s", logs)
	}
	if !strings.Contains(logs, "app/code") {
		t.Fatalf("expected path filter in rsync invocation:\n%s", logs)
	}
}

func TestSyncCommandRuntimeNoResumeNoCompressFlags(t *testing.T) {
	resetSyncFlagsForRuntimeTest(t)

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: laravel
remotes:
  staging:
    host: staging.example.com
    user: deploy
    path: /srv/www/app
`)

	shimDir := t.TempDir()
	logPath := filepath.Join(shimDir, "rsync.log")
	installSyncRuntimeRsyncShim(t, shimDir)
	t.Setenv("SYNC_RUNTIME_RSYNC_LOG", logPath)
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{
		"sync",
		"--yes",
		"--source", "staging",
		"--destination", "local",
		"--file",
		"--no-resume",
		"--no-compress",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("sync runtime with no-resume/no-compress failed: %v", err)
	}

	logs := readRuntimeLog(t, logPath)
	if !strings.Contains(logs, "rsync|-av ") {
		t.Fatalf("expected non-compressed rsync mode (-av), got:\n%s", logs)
	}
	if strings.Contains(logs, "--append-verify") || strings.Contains(logs, "--partial") {
		t.Fatalf("did not expect resume flags with --no-resume, got:\n%s", logs)
	}
}

func installSyncRuntimeRsyncShim(t *testing.T, shimDir string) {
	t.Helper()
	script := `#!/bin/sh
set -eu
log="${SYNC_RUNTIME_RSYNC_LOG:-}"
if [ -n "$log" ]; then
  printf 'rsync|%s\n' "$*" >> "$log"
fi
exit 0
`
	path := filepath.Join(shimDir, "rsync")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write rsync shim: %v", err)
	}
}

func resetSyncFlagsForRuntimeTest(t *testing.T) {
	t.Helper()
	cmd.ResetSyncFlagsForTest()
	t.Cleanup(cmd.ResetSyncFlagsForTest)
}
