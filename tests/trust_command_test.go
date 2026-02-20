package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestTrustCommandUsesRunE(t *testing.T) {
	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"trust"})
	if err != nil {
		t.Fatalf("find trust: %v", err)
	}
	if command.RunE == nil {
		t.Fatal("expected trust command to use RunE so failures return a non-zero exit code")
	}
}
