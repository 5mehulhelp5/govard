package tests

import (
	"testing"

	"govard/internal/engine/bootstrap"
)

func TestBootstrapDispatcherUnknown(t *testing.T) {
	err := bootstrap.Run("unknown", bootstrap.DefaultOptions())
	if err == nil {
		t.Fatal("expected error")
	}
}
