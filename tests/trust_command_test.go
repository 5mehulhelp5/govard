package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestTrustCommandUsesRunE(t *testing.T) {
	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"doctor", "trust"})
	if err != nil {
		t.Fatalf("find doctor trust: %v", err)
	}
	if command.RunE == nil {
		t.Fatal("expected doctor trust command to use RunE so failures return a non-zero exit code")
	}
}
