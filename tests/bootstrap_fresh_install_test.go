package tests

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"
	bootstrapengine "govard/internal/engine/bootstrap"

	"github.com/spf13/cobra"
)

func TestRunBootstrapFreshCreateProjectForTestBuildsExpectedCommand(t *testing.T) {
	var gotCommandLine string
	defer cmd.SetPHPContainerShellRunnerForTest(func(config engine.Config, commandLine string) error {
		gotCommandLine = commandLine
		return nil
	})()

	err := cmd.RunBootstrapFreshCreateProjectForTest(
		&cobra.Command{},
		engine.Config{ProjectName: "sample-project"},
		"magento/project-community-edition",
		"2.4.8",
	)
	if err != nil {
		t.Fatalf("RunBootstrapFreshCreateProjectForTest() error = %v", err)
	}

	if !strings.Contains(gotCommandLine, "composer create-project -n --ignore-platform-reqs --repository-url=https://repo.magento.com 'magento/project-community-edition' /tmp/govard-create-project '2.4.8'") {
		t.Fatalf("unexpected create-project command: %s", gotCommandLine)
	}
	if !strings.Contains(gotCommandLine, "rm -rf /tmp/govard-create-project") {
		t.Fatalf("expected cleanup commands in shell line: %s", gotCommandLine)
	}
}

func TestRunBootstrapFrameworkFreshInstallForTestRejectsUnsupportedFramework(t *testing.T) {
	err := cmd.RunBootstrapFrameworkFreshInstallForTest(
		&cobra.Command{},
		engine.Config{Framework: "custom"},
		"dev",
		"",
	)
	if err == nil {
		t.Fatal("expected unsupported framework error")
	}
	if !strings.Contains(err.Error(), "fresh install not supported for framework: custom") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunBootstrapFrameworkFreshInstallForTestNextJSUsesThrowawayContainer(t *testing.T) {
	tempDir := t.TempDir()
	chdirForTest(t, tempDir)

	var capturedProjectDir string
	var capturedCommand string
	defer cmd.SetNodeCreateProjectRunnerForTest(func(config engine.Config, projectDir string, commandLine string) error {
		capturedProjectDir = projectDir
		capturedCommand = commandLine
		stageDir := extractStageHostDir(t, commandLine)
		return os.WriteFile(filepath.Join(stageDir, "package.json"), []byte("{\"name\":\"nextjs-app\"}\n"), 0o644)
	})()

	err := cmd.RunBootstrapFrameworkFreshInstallForTest(
		&cobra.Command{},
		engine.Config{
			ProjectName: "sample-project",
			Framework:   "nextjs",
			Domain:      "sample.test",
		},
		"dev",
		"",
	)
	if err != nil {
		t.Fatalf("RunBootstrapFrameworkFreshInstallForTest() error = %v", err)
	}

	if capturedProjectDir != tempDir {
		t.Fatalf("expected runner to receive project dir %q, got %q", tempDir, capturedProjectDir)
	}
	if !strings.Contains(capturedCommand, "npx create-next-app@latest") {
		t.Fatalf("expected create-next-app invocation, got: %s", capturedCommand)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "package.json")); err != nil {
		t.Fatalf("expected staged package.json to land in the project dir: %v", err)
	}
}

func TestRunBootstrapFrameworkFreshInstallForTestWordPressDoesNotRestartEnvironment(t *testing.T) {
	tempDir := t.TempDir()
	chdirForTest(t, tempDir)

	restoreDownloader := bootstrapengine.SetWordPressCoreDownloaderForTest(func(projectDir string) error {
		samplePath := filepath.Join(projectDir, "wp-config-sample.php")
		return os.WriteFile(samplePath, []byte("<?php\n"), 0o644)
	})
	defer restoreDownloader()

	defer cmd.SetPHPContainerShellRunnerForTest(func(config engine.Config, commandLine string) error {
		return nil
	})()

	calls := make([][]string, 0, 2)
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		captured := append([]string{}, args...)
		calls = append(calls, captured)
		return nil
	})()

	err := cmd.RunBootstrapFrameworkFreshInstallForTest(
		&cobra.Command{},
		engine.Config{
			ProjectName: "sample-project",
			Framework:   "wordpress",
			Domain:      "sample.test",
		},
		"dev",
		"",
	)
	if err != nil {
		t.Fatalf("RunBootstrapFrameworkFreshInstallForTest() error = %v", err)
	}

	want := [][]string{
		{"config", "auto"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("subcommand calls = %#v, want %#v", calls, want)
	}
}
