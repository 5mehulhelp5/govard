package tests

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocsCommandsExist(t *testing.T) {
	required := []string{
		"docs/commands/bootstrap.md",
		"docs/commands/env.md",
		"docs/commands/remote.md",
		"docs/commands/sync.md",
		"docs/commands/db.md",
		"docs/commands/deploy.md",
		"docs/commands/open.md",
		"docs/commands/lock.md",
		"docs/commands/tunnel.md",
		"docs/commands/upgrade.md",
		"docs/commands/snapshot.md",
		"docs/commands/profile.md",
	}

	for _, path := range required {
		if _, err := os.Stat(filepath.Join("..", path)); err != nil {
			t.Fatalf("missing doc: %s", path)
		}
	}
}
