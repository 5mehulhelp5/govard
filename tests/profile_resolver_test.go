package tests

import (
	"testing"

	"govard/internal/engine"
)

func TestResolveRuntimeProfileWithVersionOverride(t *testing.T) {
	result, err := engine.ResolveRuntimeProfile("laravel", "^11.0")
	if err != nil {
		t.Fatalf("resolve profile: %v", err)
	}

	if result.Profile.Framework != "laravel" {
		t.Fatalf("expected framework laravel, got %s", result.Profile.Framework)
	}
	if result.Profile.PHPVersion != "8.3" {
		t.Fatalf("expected laravel 11 PHP 8.3 profile, got %s", result.Profile.PHPVersion)
	}
	if result.Source == "framework-defaults" {
		t.Fatalf("expected version-specific source, got %s", result.Source)
	}
}

func TestResolveRuntimeProfileUnknownVersionFallsBack(t *testing.T) {
	result, err := engine.ResolveRuntimeProfile("laravel", "99.0")
	if err != nil {
		t.Fatalf("resolve profile: %v", err)
	}
	if result.Source != "framework-defaults" {
		t.Fatalf("expected framework-defaults source, got %s", result.Source)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected fallback warning for unknown major version")
	}
}

func TestResolveRuntimeProfileNextjsNoDatabase(t *testing.T) {
	result, err := engine.ResolveRuntimeProfile("nextjs", "14")
	if err != nil {
		t.Fatalf("resolve profile: %v", err)
	}
	if result.Profile.DBType != "none" {
		t.Fatalf("expected nextjs db_type none, got %s", result.Profile.DBType)
	}
	if result.Profile.DBVersion != "" {
		t.Fatalf("expected nextjs empty db version, got %s", result.Profile.DBVersion)
	}
}

func TestResolveRuntimeProfileEmdashUsesNodeAndNoDatabase(t *testing.T) {
	result, err := engine.ResolveRuntimeProfile("emdash", "^0.1.0")
	if err != nil {
		t.Fatalf("resolve profile: %v", err)
	}
	if result.Profile.NodeVersion != "22" {
		t.Fatalf("expected emdash node version 22, got %s", result.Profile.NodeVersion)
	}
	if result.Profile.DBType != "none" {
		t.Fatalf("expected emdash db_type none, got %s", result.Profile.DBType)
	}
	if result.Profile.WebServer != "none" {
		t.Fatalf("expected emdash web_server none, got %s", result.Profile.WebServer)
	}
}

func TestExtractMajorVersion(t *testing.T) {
	major, ok := engine.ExtractMajorVersion("^11.3")
	if !ok {
		t.Fatal("expected major version to be extracted")
	}
	if major != 11 {
		t.Fatalf("expected major 11, got %d", major)
	}
}
