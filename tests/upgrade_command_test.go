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

	if f := command.Flags().Lookup("no-db-upgrade"); f == nil {
		t.Errorf("expected flag --no-db-upgrade to exist")
	}
	if f := command.Flags().Lookup("no-env-update"); f == nil {
		t.Errorf("expected flag --no-env-update to exist")
	}
}
