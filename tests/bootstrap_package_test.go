package tests

import (
	"testing"

	"govard/internal/engine/bootstrap"
)

func TestBootstrapPkgDefaultOptions(t *testing.T) {
	opts := bootstrap.DefaultOptions()
	if opts.Source != "staging" {
		t.Fatalf("expected source staging, got %s", opts.Source)
	}
}

func TestBootstrapPkgMagento2FreshCommands(t *testing.T) {
	cmds := bootstrap.Magento2FreshCommands(bootstrap.Options{})
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
}

func TestBootstrapPkgRunUnsupportedRecipe(t *testing.T) {
	if err := bootstrap.Run("unknown", bootstrap.Options{}); err == nil {
		t.Fatal("expected unsupported recipe error")
	}
}
