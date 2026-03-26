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

func TestResolveRemoteMediaPathMagento1UsesLegacyMediaDir(t *testing.T) {
	cfg := engine.Config{Framework: "magento1", Remotes: map[string]engine.RemoteConfig{
		"development": {Path: "/home/m1.example.com/public_html"},
	}}
	root, media := engine.ResolveRemotePaths(cfg, "development")
	if root != "/home/m1.example.com/public_html" {
		t.Fatalf("root mismatch: %s", root)
	}
	if media != "/home/m1.example.com/public_html/media" {
		t.Fatalf("media mismatch: %s", media)
	}
}

func TestResolveRemoteMediaPathOpenMageUsesLegacyMediaDir(t *testing.T) {
	cfg := engine.Config{Framework: "openmage", Remotes: map[string]engine.RemoteConfig{
		"development": {Path: "/home/openmage.example.com/public_html"},
	}}
	root, media := engine.ResolveRemotePaths(cfg, "development")
	if root != "/home/openmage.example.com/public_html" {
		t.Fatalf("root mismatch: %s", root)
	}
	if media != "/home/openmage.example.com/public_html/media" {
		t.Fatalf("media mismatch: %s", media)
	}
}
