package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestProxyCommandStructure(t *testing.T) {
	subcommands := map[string]bool{
		"start":   false,
		"stop":    false,
		"restart": false,
		"status":  false,
		"routes":  false,
	}

	for _, command := range cmd.ProxyCommand().Commands() {
		if _, ok := subcommands[command.Name()]; ok {
			subcommands[command.Name()] = true
		}
	}

	for name, seen := range subcommands {
		if !seen {
			t.Fatalf("Expected proxy subcommand %s", name)
		}
	}
}
