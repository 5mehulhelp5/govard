package tests

import (
	"io"
	"strings"
	"testing"

	"govard/internal/cmd"
)

func TestDbEnvironmentFlag(t *testing.T) {
	cmd.ResetDBFlagsForTest()
	t.Cleanup(cmd.ResetDBFlagsForTest)

	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"db"})
	if err != nil {
		t.Fatalf("find db: %v", err)
	}
	if command.Flags().Lookup("environment") == nil {
		t.Fatal("expected --environment flag on db command")
	}
}

func TestDbNoPIIPrimaryShorthandIsP(t *testing.T) {
	cmd.ResetDBFlagsForTest()
	t.Cleanup(cmd.ResetDBFlagsForTest)

	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"db"})
	if err != nil {
		t.Fatalf("find db: %v", err)
	}

	flag := command.Flags().Lookup("no-pii")
	if flag == nil {
		t.Fatal("expected --no-pii flag on db command")
	}
	if flag.Shorthand != "P" {
		t.Fatalf("expected --no-pii shorthand to be -P, got -%s", flag.Shorthand)
	}
	if command.Flags().Lookup("sanitize") == nil {
		t.Fatal("expected --sanitize alias on db command")
	}
}

func TestDbSanitizeAliasAndLegacyShorthandBehaveLikeNoPII(t *testing.T) {
	cmd.ResetDBFlagsForTest()
	t.Cleanup(cmd.ResetDBFlagsForTest)

	for _, args := range [][]string{
		{"db", "connect", "--sanitize"},
		{"db", "connect", "-S"},
		{"db", "connect", "-P"},
	} {
		root := cmd.RootCommandForTest()
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		root.SetArgs(args)

		err := root.Execute()
		if err == nil {
			t.Fatalf("expected validation error for args %v", args)
		}
		if !strings.Contains(err.Error(), "--no-pii") {
			t.Fatalf("expected --no-pii validation semantics for args %v, got: %v", args, err)
		}
	}
}
