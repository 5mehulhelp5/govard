package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestDoctorCommandExists(t *testing.T) {
	command := cmd.DoctorCommand()
	if command.Use != "doctor" {
		t.Fatalf("unexpected doctor use: %s", command.Use)
	}
	if command.Flags().Lookup("json") == nil {
		t.Fatal("expected --json flag on doctor command")
	}
	if command.Flags().Lookup("fix") == nil {
		t.Fatal("expected --fix flag on doctor command")
	}
	if command.Flags().Lookup("pack") == nil {
		t.Fatal("expected --pack flag on doctor command")
	}
	if command.Flags().Lookup("pack-dir") == nil {
		t.Fatal("expected --pack-dir flag on doctor command")
	}
}

func TestFixDepsCommandExists(t *testing.T) {
	command := cmd.FixDepsCommand()
	if command.Use != "deps" {
		t.Fatalf("unexpected deps use: %s", command.Use)
	}

	aliases := command.Aliases
	hasAlias := false
	for _, alias := range aliases {
		if alias == "fix-deps" {
			hasAlias = true
			break
		}
	}
	if !hasAlias {
		t.Fatal("expected fix-deps compatibility alias")
	}
}
