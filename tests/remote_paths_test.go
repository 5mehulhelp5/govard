package tests

import (
	"testing"

	"govard/internal/engine"
)

func TestResolveRemoteMediaPath(t *testing.T) {
	cfg := engine.Config{Framework: "magento2", Remotes: map[string]engine.RemoteConfig{
		"staging": {Path: "/var/www/html"},
	}}
	root, media := engine.ResolveRemotePaths(cfg, "staging")
	if root != "/var/www/html" {
		t.Fatalf("root mismatch")
	}
	if media != "/var/www/html/pub/media" {
		t.Fatalf("media mismatch: %s", media)
	}
}
