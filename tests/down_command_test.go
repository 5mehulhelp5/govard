package tests

import (
	"reflect"
	"testing"

	"govard/internal/cmd"
)

func TestDownCommandFlagsExist(t *testing.T) {
	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"down"})
	if err != nil {
		t.Fatalf("find down: %v", err)
	}

	for _, name := range []string{"remove-orphans", "volumes", "rmi", "timeout"} {
		if command.Flags().Lookup(name) == nil {
			t.Fatalf("expected --%s flag on down command", name)
		}
	}
}

func TestBuildDownComposeArgsDefaults(t *testing.T) {
	args, err := cmd.BuildDownComposeArgsForTest(
		"/work/project",
		"/tmp/compose.yml",
		"demo",
		true,
		false,
		"",
		0,
	)
	if err != nil {
		t.Fatalf("build args: %v", err)
	}

	expected := []string{
		"compose",
		"--project-directory",
		"/work/project",
		"-p",
		"demo",
		"-f",
		"/tmp/compose.yml",
		"down",
		"--remove-orphans",
	}
	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected args\nwant: %#v\ngot:  %#v", expected, args)
	}
}

func TestBuildDownComposeArgsWithAllOptions(t *testing.T) {
	args, err := cmd.BuildDownComposeArgsForTest(
		"/work/project",
		"/tmp/compose.yml",
		"demo",
		true,
		true,
		"local",
		15,
	)
	if err != nil {
		t.Fatalf("build args: %v", err)
	}

	expected := []string{
		"compose",
		"--project-directory",
		"/work/project",
		"-p",
		"demo",
		"-f",
		"/tmp/compose.yml",
		"down",
		"--remove-orphans",
		"--volumes",
		"--rmi",
		"local",
		"--timeout",
		"15",
	}
	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected args\nwant: %#v\ngot:  %#v", expected, args)
	}
}

func TestBuildDownComposeArgsRejectsInvalidRMI(t *testing.T) {
	_, err := cmd.BuildDownComposeArgsForTest(
		"/work/project",
		"/tmp/compose.yml",
		"demo",
		true,
		false,
		"invalid",
		0,
	)
	if err == nil {
		t.Fatal("expected error for invalid --rmi value")
	}
}

func TestBuildDownComposeArgsRejectsNegativeTimeout(t *testing.T) {
	_, err := cmd.BuildDownComposeArgsForTest(
		"/work/project",
		"/tmp/compose.yml",
		"demo",
		true,
		false,
		"",
		-1,
	)
	if err == nil {
		t.Fatal("expected error for negative timeout")
	}
}
