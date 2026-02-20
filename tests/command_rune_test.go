package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestCommandsUseRunEForErrorPropagation(t *testing.T) {
	tests := []struct {
		name string
		path []string
	}{
		{name: "init", path: []string{"init"}},
		{name: "logs", path: []string{"logs"}},
		{name: "stop", path: []string{"stop"}},
		{name: "redis", path: []string{"redis"}},
		{name: "valkey", path: []string{"valkey"}},
		{name: "elasticsearch", path: []string{"elasticsearch"}},
		{name: "opensearch", path: []string{"opensearch"}},
		{name: "snapshot create", path: []string{"snapshot", "create"}},
		{name: "snapshot list", path: []string{"snapshot", "list"}},
		{name: "snapshot restore", path: []string{"snapshot", "restore"}},
		{name: "remote audit tail", path: []string{"remote", "audit", "tail"}},
		{name: "remote audit stats", path: []string{"remote", "audit", "stats"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			root := cmd.RootCommandForTest()
			command, _, err := root.Find(tt.path)
			if err != nil {
				t.Fatalf("find command %v: %v", tt.path, err)
			}
			if command == nil {
				t.Fatalf("command %v not found", tt.path)
			}
			if command.RunE == nil {
				t.Fatalf("expected command %v to use RunE for proper non-zero exit codes", tt.path)
			}
		})
	}
}
