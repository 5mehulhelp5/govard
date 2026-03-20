package tests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"govard/internal/engine"
	"govard/internal/engine/tunnel"
)

func TestMagento2ManagerUpdate(t *testing.T) {
	var executedSQL string
	mgr := &tunnel.Magento2Manager{
		Executor: func(name string, args ...string) ([]byte, error) {
			if name == "docker" && args[0] == "exec" {
				executedSQL = args[len(args)-1]
			}
			return []byte(""), nil
		},
	}

	config := engine.Config{ProjectName: "demo", Domain: "demo.test"}
	err := mgr.Update(".", config, "https://tunnel.live")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "UPDATE core_config_data SET value='https://tunnel.live/' WHERE path IN ('web/unsecure/base_url', 'web/secure/base_url')"
	if executedSQL != expected {
		t.Fatalf("expected SQL %q, got %q", expected, executedSQL)
	}
}

func TestLaravelManagerUpdate(t *testing.T) {
	envContent := "APP_NAME=Laravel\nAPP_URL=http://localhost\nDB_CONNECTION=mysql"
	var writtenContent string

	mgr := &tunnel.LaravelManager{
		ReadFile: func(path string) ([]byte, error) {
			return []byte(envContent), nil
		},
		WriteFile: func(path string, data []byte, mode os.FileMode) error {
			writtenContent = string(data)
			return nil
		},
	}

	config := engine.Config{Domain: "laravel.test"}
	err := mgr.Update(".", config, "https://tunnel.live")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(writtenContent, "APP_URL=https://tunnel.live") {
		t.Fatalf("expected APP_URL update, got:\n%s", writtenContent)
	}
}

func TestBaseURLManagerFactory(t *testing.T) {
	tests := []struct {
		framework string
		expected  string
	}{
		{"magento2", "*tunnel.Magento2Manager"},
		{"Laravel", "*tunnel.LaravelManager"},
		{"wordpress", "*tunnel.WordPressManager"},
		{"Symfony", "*tunnel.SymfonyManager"},
		{"Unknown", "*tunnel.NoopManager"},
	}

	for _, tt := range tests {
		mgr := tunnel.NewBaseURLManager(tt.framework)
		typeName := fmt.Sprintf("%T", mgr)
		if typeName != tt.expected {
			t.Errorf("framework %s: expected type %s, got %s", tt.framework, tt.expected, typeName)
		}
	}
}
