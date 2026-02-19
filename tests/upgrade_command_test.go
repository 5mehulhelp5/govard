package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestUpgradeCommandExists(t *testing.T) {
	command := cmd.UpgradeCommand()
	if command.Use != "upgrade" {
		t.Fatalf("unexpected use: %s", command.Use)
	}
}
