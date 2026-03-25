package tests

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocsCommandsExist(t *testing.T) {
	required := []string{
		"docs/README.md",
		"docs/getting-started.md",
		"docs/commands.md",
		"docs/configuration.md",
		"docs/remotes-and-sync.md",
		"docs/frameworks.md",
		"docs/ssl-and-domains.md",
		"docs/desktop.md",
		"docs/architecture.md",
		"docs/contributing.md",
	}

	for _, path := range required {
		if _, err := os.Stat(filepath.Join("..", path)); err != nil {
			t.Fatalf("missing doc: %s", path)
		}
	}
}
