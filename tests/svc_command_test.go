package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestSvcUpCommandTrustFlagsExist(t *testing.T) {
	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"svc", "up"})
	if err != nil {
		t.Fatalf("find svc up: %v", err)
	}
	for _, name := range []string{"auto-trust", "trust-browsers"} {
		if command.Flags().Lookup(name) == nil {
			t.Fatalf("expected --%s flag on svc up command", name)
		}
	}
}

func TestSvcRestartCommandTrustFlagsExist(t *testing.T) {
	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"svc", "restart"})
	if err != nil {
		t.Fatalf("find svc restart: %v", err)
	}
	for _, name := range []string{"pull", "remove-orphans", "auto-trust", "trust-browsers"} {
		if command.Flags().Lookup(name) == nil {
			t.Fatalf("expected --%s flag on svc restart command", name)
		}
	}
}
