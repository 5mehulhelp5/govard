package tests

import (
	"testing"

	"govard/internal/engine/bootstrap"
)

func TestBootstrapOptionsDefaults(t *testing.T) {
	opts := bootstrap.DefaultOptions()
	if opts.Source == "" {
		t.Fatal("expected default source set")
	}
}
