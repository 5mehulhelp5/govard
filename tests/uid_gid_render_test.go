package tests

import (
	"strings"
	"testing"

	"govard/internal/engine"
)

func TestRenderIncludesUIDGID(t *testing.T) {
	content := renderComposeWithConfig(t, engine.Config{
		ProjectName: "uidgid-test",
		Framework:   "laravel",
		Domain:      "uidgid.test",
	})

	if !strings.Contains(content, "PUID=") {
		t.Fatal("missing PUID in compose output")
	}
	if !strings.Contains(content, "PGID=") {
		t.Fatal("missing PGID in compose output")
	}
}
