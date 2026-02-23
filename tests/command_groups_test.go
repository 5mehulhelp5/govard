package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestSvcCommandStructure(t *testing.T) {
	root := cmd.RootCommandForTest()
	required := [][]string{
		{"svc"},
		{"svc", "up"},
		{"svc", "down"},
		{"svc", "restart"},
		{"svc", "ps"},
		{"svc", "logs"},
		{"svc", "sleep"},
		{"svc", "wake"},
	}

	for _, path := range required {
		path := path
		t.Run(path[len(path)-1], func(t *testing.T) {
			command, _, err := root.Find(path)
			if err != nil {
				t.Fatalf("find command %v: %v", path, err)
			}
			if command == nil {
				t.Fatalf("command %v not found", path)
			}
		})
	}
}

func TestEnvCommandStructure(t *testing.T) {
	root := cmd.RootCommandForTest()
	required := [][]string{
		{"env"},
		{"env", "up"},
		{"env", "start"},
		{"env", "stop"},
		{"env", "down"},
		{"env", "restart"},
		{"env", "ps"},
		{"env", "logs"},
		{"env", "redis"},
		{"env", "valkey"},
		{"env", "elasticsearch"},
		{"env", "opensearch"},
		{"env", "varnish"},
	}

	for _, path := range required {
		path := path
		t.Run(path[len(path)-1], func(t *testing.T) {
			command, _, err := root.Find(path)
			if err != nil {
				t.Fatalf("find command %v: %v", path, err)
			}
			if command == nil {
				t.Fatalf("command %v not found", path)
			}
		})
	}
}

func TestDoctorSubcommandsIncludeTrustAndFixDeps(t *testing.T) {
	root := cmd.RootCommandForTest()

	for _, path := range [][]string{
		{"doctor", "trust"},
		{"doctor", "fix-deps"},
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

func TestConfigSubcommandsIncludeAutoAndProfile(t *testing.T) {
	root := cmd.RootCommandForTest()

	for _, path := range [][]string{
		{"config", "get"},
		{"config", "set"},
		{"config", "auto"},
		{"config", "profile"},
		{"config", "profile", "apply"},
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

func TestToolCommandStructure(t *testing.T) {
	root := cmd.RootCommandForTest()

	for _, path := range [][]string{
		{"tool"},
		{"tool", "magento"},
		{"tool", "artisan"},
		{"tool", "composer"},
		{"tool", "npm"},
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
