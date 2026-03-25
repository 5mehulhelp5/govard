package tests

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/cmd"
)

func TestSyncCommandFromAndToAliasesAffectPlanEndpoints(t *testing.T) {
	resetSyncFlagsForAliasTest(t)

	tempDir := t.TempDir()
	writeSyncAliasConfig(t, tempDir)
	chdirForTest(t, tempDir)

	root := cmd.RootCommandForTest()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"sync", "--plan", "--file", "--from", "local", "--to", "dev"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Source:      local") {
		t.Fatalf("expected source from --from alias, got: %s", out)
	}
	if !strings.Contains(out, "Destination: dev") {
		t.Fatalf("expected destination from --to alias, got: %s", out)
	}
}

func TestSyncCommandLegacyEnvironmentAliasStillResolvesSource(t *testing.T) {
	resetSyncFlagsForAliasTest(t)

	tempDir := t.TempDir()
	writeSyncAliasConfig(t, tempDir)
	chdirForTest(t, tempDir)

	root := cmd.RootCommandForTest()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"sync", "--plan", "--file", "--environment", "dev"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Source:      dev") {
		t.Fatalf("expected source from legacy --environment alias, got: %s", out)
	}
}

func TestSyncCommandSourceWinsOverLegacyEnvironmentAlias(t *testing.T) {
	resetSyncFlagsForAliasTest(t)

	tempDir := t.TempDir()
	writeSyncAliasConfig(t, tempDir)
	chdirForTest(t, tempDir)

	root := cmd.RootCommandForTest()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"sync", "--plan", "--file", "--environment", "dev", "--source", "staging"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Source:      staging") {
		t.Fatalf("expected --source to take precedence over --environment, got: %s", out)
	}
	if strings.Contains(out, "Source:      dev") {
		t.Fatalf("did not expect legacy --environment to override --source, got: %s", out)
	}
}

func TestResetSyncFlagsForTestClearsStringArrayFlags(t *testing.T) {
	resetSyncFlagsForAliasTest(t)

	syncCmd := cmd.SyncCommand()
	if err := syncCmd.Flags().Set("include", "app/*"); err != nil {
		t.Fatalf("set include: %v", err)
	}
	if err := syncCmd.Flags().Set("exclude", "vendor/"); err != nil {
		t.Fatalf("set exclude: %v", err)
	}

	cmd.ResetSyncFlagsForTest()

	include, err := syncCmd.Flags().GetStringArray("include")
	if err != nil {
		t.Fatalf("get include: %v", err)
	}
	exclude, err := syncCmd.Flags().GetStringArray("exclude")
	if err != nil {
		t.Fatalf("get exclude: %v", err)
	}
	if len(include) != 0 {
		t.Fatalf("expected include flags to reset cleanly, got: %#v", include)
	}
	if len(exclude) != 0 {
		t.Fatalf("expected exclude flags to reset cleanly, got: %#v", exclude)
	}
}

func resetSyncFlagsForAliasTest(t *testing.T) {
	t.Helper()
	cmd.ResetSyncFlagsForTest()
	t.Cleanup(cmd.ResetSyncFlagsForTest)
}

func writeSyncAliasConfig(t *testing.T, tempDir string) {
	t.Helper()

	configPath := filepath.Join(tempDir, ".govard.yml")
	config := `project_name: sample-project
domain: sample.test
framework: laravel
remotes:
  staging:
    host: staging.example.com
    user: deploy
    path: /srv/www/staging
  dev:
    host: dev.example.com
    user: deploy
    path: /srv/www/dev
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
}
