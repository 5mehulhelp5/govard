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
		{name: "env logs", path: []string{"env", "logs"}},
		{name: "env stop", path: []string{"env", "stop"}},
		{name: "env redis flush", path: []string{"env", "redis", "flush"}},
		{name: "env redis info", path: []string{"env", "redis", "info"}},
		{name: "env redis cli", path: []string{"env", "redis", "cli"}},
		{name: "env valkey flush", path: []string{"env", "valkey", "flush"}},
		{name: "env elasticsearch", path: []string{"env", "elasticsearch"}},
		{name: "env opensearch", path: []string{"env", "opensearch"}},
		{name: "svc sleep", path: []string{"svc", "sleep"}},
		{name: "svc wake", path: []string{"svc", "wake"}},
		{name: "config get", path: []string{"config", "get"}},
		{name: "config set", path: []string{"config", "set"}},
		{name: "db top", path: []string{"db", "top"}},
		{name: "db query", path: []string{"db", "query"}},
		{name: "db info", path: []string{"db", "info"}},
		{name: "debug on", path: []string{"debug", "on"}},
		{name: "debug off", path: []string{"debug", "off"}},
		{name: "debug status", path: []string{"debug", "status"}},
		{name: "debug shell", path: []string{"debug", "shell"}},
		{name: "test phpunit", path: []string{"test", "phpunit"}},
		{name: "test phpstan", path: []string{"test", "phpstan"}},
		{name: "lock generate", path: []string{"lock", "generate"}},
		{name: "lock check", path: []string{"lock", "check"}},
		{name: "tunnel start", path: []string{"tunnel", "start"}},
		{name: "tunnel stop", path: []string{"tunnel", "stop"}},
		{name: "tunnel status", path: []string{"tunnel", "status"}},
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
			} else if command.RunE == nil {
				t.Fatalf("expected command %v to use RunE for proper non-zero exit codes", tt.path)
			}
		})
	}
}
