package tests

import (
	"govard/internal/cmd"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitPopulatesFeaturesExplicitly(t *testing.T) {
	tempDir := t.TempDir()
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"init", "--framework", "magento2", "-y"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	configPath := filepath.Join(tempDir, ".govard.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	requiredKeys := []string{
		"cache: redis",
		"search: opensearch",
		"db: mariadb",
		"varnish: true",
		"xdebug: true",
	}

	for _, key := range requiredKeys {
		if !strings.Contains(content, key) {
			t.Errorf("expected .govard.yml to contain %q, but it didn't. Content:\n%s", key, content)
		}
	}
}
