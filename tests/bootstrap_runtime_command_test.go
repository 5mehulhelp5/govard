package tests

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"

	"github.com/spf13/cobra"
)

func TestBootstrapCommandRuntimePlanSkipUpDoesNotRunSubcommands(t *testing.T) {
	resetBootstrapFlagsForRuntimeTest(t)

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: laravel
`)

	calls := make([][]string, 0, 1)
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		captured := make([]string, len(args))
		copy(captured, args)
		calls = append(calls, captured)
		return nil
	})()

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"bootstrap", "--plan", "--skip-up"})
	if err := root.Execute(); err != nil {
		t.Fatalf("bootstrap --plan failed: %v", err)
	}

	if len(calls) != 0 {
		t.Fatalf("expected no subcommand calls for --plan --skip-up, got %#v", calls)
	}
}

func TestBootstrapCommandRuntimeRemoteFlowRunsExpectedSubcommands(t *testing.T) {
	resetBootstrapFlagsForRuntimeTest(t)

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
	t.Setenv("HOME", tempDir)
	if err := os.WriteFile(filepath.Join(tempDir, "composer.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write composer.json: %v", err)
	}

	calls := make([][]string, 0, 8)
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		captured := make([]string, len(args))
		copy(captured, args)
		calls = append(calls, captured)
		return nil
	})()
	defer cmd.SetBootstrapRemoteDirExistsForTest(func(remoteName string, remoteCfg engine.RemoteConfig, remotePath string) bool {
		return true
	})()

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"bootstrap", "--environment", "staging", "--skip-up"})
	if err := root.Execute(); err != nil {
		t.Fatalf("bootstrap remote flow failed: %v", err)
	}

	want := [][]string{
		{"remote", "test", "staging"},
		{"tool", "composer", "install", "-n"},
		{"tool", "composer", "dump-autoload", "-o", "-n"},
		{"db", "import", "--stream-db", "--environment", "staging"},
		{"tool", "composer", "dump-autoload", "-o", "-n"},
		{"sync", "--source", "staging", "--media"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("subcommand calls = %#v, want %#v", calls, want)
	}
}

func TestBootstrapCommandRuntimeCloneCodeOnlySkipsDBAndMedia(t *testing.T) {
	resetBootstrapFlagsForRuntimeTest(t)

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
	t.Setenv("HOME", tempDir)
	if err := os.WriteFile(filepath.Join(tempDir, "composer.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write composer.json: %v", err)
	}

	calls := make([][]string, 0, 8)
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		captured := make([]string, len(args))
		copy(captured, args)
		calls = append(calls, captured)
		return nil
	})()

	var shellConfig engine.Config
	var shellCmdLine string
	defer cmd.SetPHPContainerShellRunnerForTest(func(config engine.Config, commandLine string) error {
		shellConfig = config
		shellCmdLine = commandLine
		return nil
	})()

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"bootstrap", "--clone", "--code-only", "--environment", "dev", "--skip-up"})
	if err := root.Execute(); err != nil {
		t.Fatalf("bootstrap --clone --code-only failed: %v", err)
	}

	if shellConfig.ProjectName != "sample-project" {
		t.Fatalf("shell runner project = %q, want sample-project", shellConfig.ProjectName)
	}
	if shellCmdLine != "rm -rf vendor" {
		t.Fatalf("shell runner command = %q, want %q", shellCmdLine, "rm -rf vendor")
	}

	if len(calls) < 5 {
		t.Fatalf("expected at least 5 subcommand calls, got %#v", calls)
	}
	if !reflect.DeepEqual(calls[0], []string{"remote", "test", "dev"}) {
		t.Fatalf("first call = %#v, want remote test", calls[0])
	}
	if len(calls[1]) < 6 || calls[1][0] != "sync" || calls[1][1] != "--source" || calls[1][2] != "dev" || calls[1][3] != "--file" {
		t.Fatalf("second call should be file sync from dev, got %#v", calls[1])
	}

	for _, call := range calls {
		if len(call) == 0 {
			continue
		}
		if call[0] == "db" {
			t.Fatalf("did not expect DB sync/import in code-only mode, got calls %#v", calls)
		}
		if call[0] == "sync" && len(call) >= 4 && call[3] == "--media" {
			t.Fatalf("did not expect media sync in code-only mode, got calls %#v", calls)
		}
	}
}

func TestBootstrapCommandRuntimeDBDumpUsesLocalImportFileFlow(t *testing.T) {
	resetBootstrapFlagsForRuntimeTest(t)

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: laravel
`)

	dumpPath := filepath.Join(tempDir, "database.sql")
	calls := make([][]string, 0, 2)
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		captured := make([]string, len(args))
		copy(captured, args)
		calls = append(calls, captured)
		return nil
	})()

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{
		"bootstrap",
		"--clone=false",
		"--skip-up",
		"--no-media",
		"--no-composer",
		"--db-dump", dumpPath,
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("bootstrap with --db-dump failed: %v", err)
	}

	want := [][]string{
		{"db", "import", "--file", dumpPath},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("subcommand calls = %#v, want %#v", calls, want)
	}
}

func resetBootstrapFlagsForRuntimeTest(t *testing.T) {
	t.Helper()
	cmd.ResetBootstrapFlags()
	t.Cleanup(cmd.ResetBootstrapFlags)
}
