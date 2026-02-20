package tests

import (
	"testing"

	"govard/internal/engine"
)

func TestNormalizeProviderName(t *testing.T) {
	if got := engine.NormalizeProviderName("  Cloudflare-Tunnel "); got != "cloudflare-tunnel" {
		t.Fatalf("expected normalized provider name, got %q", got)
	}
}

func TestValidateProviderRef(t *testing.T) {
	if err := engine.ValidateProviderRef(engine.ProviderRef{Kind: engine.ProviderKindTunnel, Name: "cloudflare"}); err != nil {
		t.Fatalf("expected valid provider ref: %v", err)
	}

	if err := engine.ValidateProviderRef(engine.ProviderRef{Kind: "", Name: "cloudflare"}); err == nil {
		t.Fatal("expected missing kind validation error")
	}

	if err := engine.ValidateProviderRef(engine.ProviderRef{Kind: engine.ProviderKindSecrets, Name: ""}); err == nil {
		t.Fatal("expected missing provider name validation error")
	}

	if err := engine.ValidateProviderRef(engine.ProviderRef{Kind: engine.ProviderKindBlueprintRegistry, Name: "bad provider"}); err == nil {
		t.Fatal("expected provider name format validation error")
	}
}
