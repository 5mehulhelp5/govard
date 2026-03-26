package tests

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"
	"strings"

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
	defer cmd.SetPHPContainerShellRunnerForTest(func(config engine.Config, commandLine string) error {
		return nil
	})()

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"bootstrap", "--environment", "staging", "--yes", "--skip-up"})
	if err := root.Execute(); err != nil {
		t.Fatalf("bootstrap remote flow failed: %v", err)
	}

	want := [][]string{
		{"remote", "test", "staging"},
		{"tool", "composer", "install", "-n"},
		{"tool", "composer", "dump-autoload", "-n"},
		{"db", "import", "--yes", "--stream-db", "--environment", "staging"},
		{"tool", "composer", "dump-autoload", "-n"},
		{"sync", "--source", "staging", "--media", "--yes"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("subcommand calls = %#v, want %#v", calls, want)
	}
}

func TestBootstrapCommandRuntimeMagento1RemoteFlowUsesConfigAuto(t *testing.T) {
	resetBootstrapFlagsForRuntimeTest(t)

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: magento1
remotes:
  staging:
    host: staging.example.com
    user: deploy
    path: /srv/www/app
`)

	calls := make([][]string, 0, 4)
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		captured := make([]string, len(args))
		copy(captured, args)
		calls = append(calls, captured)
		return nil
	})()

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"bootstrap", "--environment", "staging", "--yes", "--clone=false", "--no-media", "--no-composer"})
	if err := root.Execute(); err != nil {
		t.Fatalf("bootstrap Magento 1 remote flow failed: %v", err)
	}

	want := [][]string{
		{"env", "up", "--remove-orphans"},
		{"remote", "test", "staging"},
		{"db", "import", "--yes", "--stream-db", "--environment", "staging"},
		{"config", "auto"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("subcommand calls = %#v, want %#v", calls, want)
	}
}

func TestBootstrapCommandRuntimeOpenMagePostCloneCreatesLocalXML(t *testing.T) {
	resetBootstrapFlagsForRuntimeTest(t)

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: openmage
`)

	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		return nil
	})()

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"bootstrap", "--yes", "--clone=false", "--skip-up", "--no-db", "--no-media"})
	if err := root.Execute(); err != nil {
		t.Fatalf("bootstrap OpenMage post-clone flow failed: %v", err)
	}

	localXMLPath := filepath.Join(tempDir, "app", "etc", "local.xml")
	if _, err := os.Stat(localXMLPath); err != nil {
		t.Fatalf("expected OpenMage bootstrap to create %s: %v", localXMLPath, err)
	}

	for _, dir := range []string{
		filepath.Join(tempDir, "var", "cache"),
		filepath.Join(tempDir, "var", "session"),
	} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected OpenMage bootstrap to create %s: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}
}

func TestBootstrapCommandRuntimeOpenMageRemoteFlowUsesConfigAuto(t *testing.T) {
	resetBootstrapFlagsForRuntimeTest(t)

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: openmage
remotes:
  staging:
    host: staging.example.com
    user: deploy
    path: /srv/www/app
`)

	calls := make([][]string, 0, 4)
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		captured := make([]string, len(args))
		copy(captured, args)
		calls = append(calls, captured)
		return nil
	})()

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"bootstrap", "--environment", "staging", "--yes", "--clone=false", "--no-media", "--no-composer"})
	if err := root.Execute(); err != nil {
		t.Fatalf("bootstrap OpenMage remote flow failed: %v", err)
	}

	want := [][]string{
		{"env", "up", "--remove-orphans"},
		{"remote", "test", "staging"},
		{"db", "import", "--yes", "--stream-db", "--environment", "staging"},
		{"config", "auto"},
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
	root.SetArgs([]string{"bootstrap", "--clone", "--code-only", "--environment", "dev", "--yes", "--skip-up"})
	if err := root.Execute(); err != nil {
		t.Fatalf("bootstrap --clone --code-only failed: %v", err)
	}

	if shellConfig.ProjectName != "sample-project" {
		t.Fatalf("shell runner project = %q, want sample-project", shellConfig.ProjectName)
	}
	// Post-clone setup runs for Laravel, which includes composer install in container
	if shellCmdLine != "rm -rf vendor" && !strings.Contains(shellCmdLine, "composer install") {
		t.Fatalf("shell runner command = %q, does not contain expected commands", shellCmdLine)
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
		"--yes",
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
		{"db", "import", "--yes", "--file", dumpPath},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("subcommand calls = %#v, want %#v", calls, want)
	}
}

func TestBootstrapCommandRuntimeNoNoiseAndNoPIIPassedToDBImport(t *testing.T) {
	resetBootstrapFlagsForRuntimeTest(t)

	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: sample-project
domain: sample.test
framework: magento2
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

	calls := make([][]string, 0)
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
	// Test with both --no-noise and --no-pii
	root.SetArgs([]string{"bootstrap", "--environment", "staging", "--yes", "--no-noise", "--no-pii", "--skip-up", "--no-media", "--no-composer"})
	if err := root.Execute(); err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	found := false
	for _, call := range calls {
		if len(call) >= 2 && call[0] == "db" && call[1] == "import" {
			found = true
			hasNoNoise := false
			hasNoPII := false
			for _, arg := range call {
				if arg == "--no-noise" {
					hasNoNoise = true
				}
				if arg == "--no-pii" {
					hasNoPII = true
				}
			}
			if !hasNoNoise || !hasNoPII {
				t.Errorf("db import call missing flags: %#v", call)
			}
		}
	}
	if !found {
		t.Error("db import call not found in calls")
	}
}

func resetBootstrapFlagsForRuntimeTest(t *testing.T) {
	t.Helper()
	cmd.ResetBootstrapFlags()
	t.Cleanup(cmd.ResetBootstrapFlags)
}
