package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestCommandAliasesResolveToExpectedCommands(t *testing.T) {
	root := cmd.RootCommandForTest()

	for _, tt := range []struct {
		path []string
		name string
	}{
		{path: []string{"boot"}, name: "bootstrap"},
		{path: []string{"cfg"}, name: "config"},
		{path: []string{"dbg"}, name: "debug"},
		{path: []string{"diag"}, name: "doctor"},
		{path: []string{"ext"}, name: "extensions"},
		{path: []string{"gui"}, name: "desktop"},
		{path: []string{"prj"}, name: "project"},
		{path: []string{"projects"}, name: "project"},
		{path: []string{"registry"}, name: "project"},
		{path: []string{"rmt"}, name: "remote"},
		{path: []string{"sh"}, name: "shell"},
		{path: []string{"snap"}, name: "snapshot"},
	} {
		command, _, err := root.Find(tt.path)
		if err != nil {
			t.Fatalf("find command %v: %v", tt.path, err)
		}
		if command == nil {
			t.Fatalf("command %v not found", tt.path)
		}
		if command.Name() != tt.name {
			t.Fatalf("command %v resolved to %q, want %q", tt.path, command.Name(), tt.name)
		}
	}
}

func TestRootEnvironmentShortcutsExist(t *testing.T) {
	root := cmd.RootCommandForTest()

	for _, path := range [][]string{
		{"up"},
		{"down"},
		{"restart"},
		{"ps"},
		{"logs"},
	} {
		command, _, err := root.Find(path)
		if err != nil {
			t.Fatalf("find command %v: %v", path, err)
		}
		if command == nil {
			t.Fatalf("command %v not found", path)
		}
	}
}
