package tests

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFrameworkSupportDocsExist(t *testing.T) {
	required := []string{
		"docs/frameworks/support-matrix.md",
	}

	for _, path := range required {
		if _, err := os.Stat(filepath.Join("..", path)); err != nil {
			t.Fatalf("missing doc: %s", path)
		}
	}
}
