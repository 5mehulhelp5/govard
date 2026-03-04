package tests

import (
	"path/filepath"
	"reflect"
	"testing"

	"govard/internal/desktop"
)

func TestDesktopRunComposeAddsRemoveOrphansForTest(t *testing.T) {
	desktop.ResetStateForTest()

	projectDir := filepath.Clean("/tmp/sample-project")
	composeFile := filepath.Join(projectDir, ".govard", "compose", "sample.yml")
	capturedDir := ""
	capturedArgs := []string{}

	restore := desktop.SetRunEnvironmentComposeForDesktopForTest(
		func(dir string, args []string) error {
			capturedDir = dir
			capturedArgs = append([]string{}, args...)
			return nil
		},
	)
	defer restore()

	if err := desktop.RunComposeForTest(projectDir, "sample-project", composeFile, true); err != nil {
		t.Fatalf("run compose: %v", err)
	}

	if capturedDir != projectDir {
		t.Fatalf("expected compose dir %q, got %q", projectDir, capturedDir)
	}

	expectedArgs := []string{
		"compose",
		"--project-directory",
		projectDir,
		"-p",
		"sample-project",
		"-f",
		composeFile,
		"up",
		"-d",
		"--remove-orphans",
	}
	if !reflect.DeepEqual(capturedArgs, expectedArgs) {
		t.Fatalf("unexpected compose args:\n got: %#v\nwant: %#v", capturedArgs, expectedArgs)
	}
}

func TestDesktopRunComposeSkipsRemoveOrphansWhenDisabledForTest(t *testing.T) {
	desktop.ResetStateForTest()

	projectDir := filepath.Clean("/tmp/sample-project")
	composeFile := filepath.Join(projectDir, ".govard", "compose", "sample.yml")
	capturedArgs := []string{}

	restore := desktop.SetRunEnvironmentComposeForDesktopForTest(
		func(_ string, args []string) error {
			capturedArgs = append([]string{}, args...)
			return nil
		},
	)
	defer restore()

	if err := desktop.RunComposeForTest(projectDir, "sample-project", composeFile, false); err != nil {
		t.Fatalf("run compose: %v", err)
	}

	for _, arg := range capturedArgs {
		if arg == "--remove-orphans" {
			t.Fatalf("did not expect --remove-orphans in compose args: %#v", capturedArgs)
		}
	}
}

func TestDesktopRunComposePullUsesPullSubcommandForTest(t *testing.T) {
	desktop.ResetStateForTest()

	projectDir := filepath.Clean("/tmp/sample-project")
	composeFile := filepath.Join(projectDir, ".govard", "compose", "sample.yml")
	capturedArgs := []string{}

	restore := desktop.SetRunEnvironmentComposeForDesktopForTest(
		func(_ string, args []string) error {
			capturedArgs = append([]string{}, args...)
			return nil
		},
	)
	defer restore()

	if err := desktop.RunComposePullForTest(projectDir, "sample-project", composeFile); err != nil {
		t.Fatalf("run compose pull: %v", err)
	}

	expectedArgs := []string{
		"compose",
		"--project-directory",
		projectDir,
		"-p",
		"sample-project",
		"-f",
		composeFile,
		"pull",
	}
	if !reflect.DeepEqual(capturedArgs, expectedArgs) {
		t.Fatalf("unexpected compose pull args:\n got: %#v\nwant: %#v", capturedArgs, expectedArgs)
	}
}
