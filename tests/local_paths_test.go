package tests

import (
	"testing"

	"govard/internal/engine"
)

func TestResolveLocalMediaPathMagento1UsesLegacyMediaDir(t *testing.T) {
	path := engine.ResolveLocalMediaPath(engine.Config{Framework: "magento1"}, "/srv/www/app")
	if path != "/srv/www/app/media" {
		t.Fatalf("ResolveLocalMediaPath() = %s, want /srv/www/app/media", path)
	}
}

func TestResolveLocalMediaPathOpenMageUsesLegacyMediaDir(t *testing.T) {
	path := engine.ResolveLocalMediaPath(engine.Config{Framework: "openmage"}, "/srv/www/app")
	if path != "/srv/www/app/media" {
		t.Fatalf("ResolveLocalMediaPath() = %s, want /srv/www/app/media", path)
	}
}
