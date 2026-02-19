package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestDbEnvironmentFlag(t *testing.T) {
	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"db"})
	if err != nil {
		t.Fatalf("find db: %v", err)
	}
	if command.Flags().Lookup("environment") == nil {
		t.Fatal("expected --environment flag on db command")
	}
}
