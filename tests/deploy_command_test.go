package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestDeployFlags(t *testing.T) {
	command := cmd.DeployCommand()
	if command.Flags().Lookup("strategy") == nil {
		t.Fatal("missing --strategy flag")
	}
	if command.Flags().Lookup("deployer") == nil {
		t.Fatal("missing --deployer flag")
	}
	if command.Flags().Lookup("deployer-config") == nil {
		t.Fatal("missing --deployer-config flag")
	}
}
