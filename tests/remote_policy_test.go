package tests

import (
	"strings"
	"testing"

	"govard/internal/engine"
)

func TestNormalizeRemoteEnvironment(t *testing.T) {
	cases := map[string]string{
		"production":  "production",
		"Prod":        "production",
		"development": "development",
		"live":        "production",
		"staging":     "staging",
		"stage":       "staging",
		"stg":         "staging",
		"dev":         "development",
		"develop":     "development",
		// Custom names pass through unchanged (no longer aliased).
		"local":      "local",
		"qa":         "qa",
		"uat":        "uat",
		"test":       "test",
		"preprod":    "preprod",
		"demo":       "demo",
		"client-uat": "client-uat",
	}

	for in, expected := range cases {
		if got := engine.NormalizeRemoteEnvironment(in); got != expected {
			t.Fatalf("normalize env %q: expected %q, got %q", in, expected, got)
		}
	}
}

func TestParseRemoteCapabilitiesCSV(t *testing.T) {
	// "all" or "none" or "" should return empty struct (all nil -> allowed)
	all, err := engine.ParseRemoteCapabilitiesCSV("all")
	if err != nil {
		t.Fatalf("parse all capabilities: %v", err)
	}
	if all != nil {
		t.Fatalf("expected all capabilities nil (allowed), got %+v", all)
	}

	// "files,db" should set those to false (blocked)
	custom, err := engine.ParseRemoteCapabilitiesCSV("files,db")
	if err != nil {
		t.Fatalf("parse custom capabilities: %v", err)
	}
	if custom.Files == nil || *custom.Files != false {
		t.Fatal("expected files to be false")
	}
	if custom.Media != nil {
		t.Fatal("expected media to be nil (allowed)")
	}
	if custom.DB == nil || *custom.DB != false {
		t.Fatal("expected db to be false")
	}
}

func TestRemoteCapabilityEnabled(t *testing.T) {
	// Default (all nil) should be enabled
	cfg := engine.RemoteConfig{}
	if !engine.RemoteCapabilityEnabled(cfg, engine.RemoteCapabilityFiles) {
		t.Fatal("expected files enabled by default")
	}

	// Explicit false should be disabled
	falseVal := false
	cfg.Capabilities = &engine.RemoteCapabilities{Files: &falseVal}
	if engine.RemoteCapabilityEnabled(cfg, engine.RemoteCapabilityFiles) {
		t.Fatal("expected files disabled when explicitly false")
	}

	// Explicit true should be enabled
	trueVal := true
	cfg.Capabilities.Files = &trueVal
	if !engine.RemoteCapabilityEnabled(cfg, engine.RemoteCapabilityFiles) {
		t.Fatal("expected files enabled when explicitly true")
	}
}

func TestParseRemoteCapabilitiesRejectsUnknown(t *testing.T) {
	_, err := engine.ParseRemoteCapabilitiesCSV("files,unknown")
	if err == nil {
		t.Fatal("expected unknown capability error")
	}
	if !strings.Contains(err.Error(), "unsupported remote capability") {
		t.Fatalf("unexpected parse error: %v", err)
	}
}

func TestParseRemoteCapabilitiesEmpty(t *testing.T) {
	parsed, err := engine.ParseRemoteCapabilitiesCSV("")
	if err != nil {
		t.Fatalf("parse empty: %v", err)
	}
	if parsed != nil {
		t.Fatal("expected empty set to result in nil")
	}
}

func TestRemoteWriteBlocked(t *testing.T) {
	explicit := engine.RemoteConfig{
		Protected: engine.BoolPtr(true),
	}
	if blocked, _ := engine.RemoteWriteBlocked("staging", explicit); !blocked {
		t.Fatal("expected explicit protected remote to block writes")
	}

	override := engine.RemoteConfig{
		Protected: engine.BoolPtr(false),
	}
	if blocked, _ := engine.RemoteWriteBlocked("prod", override); blocked {
		t.Fatal("expected explicit unprotected prod remote to allow writes")
	}

	prod := engine.RemoteConfig{}
	if blocked, _ := engine.RemoteWriteBlocked("production", prod); !blocked {
		t.Fatal("expected production remote to block writes (auto-default)")
	}

	dev := engine.RemoteConfig{}
	if blocked, _ := engine.RemoteWriteBlocked("development", dev); blocked {
		t.Fatal("expected development remote writes to be allowed")
	}
}

func TestValidateConfigRejectsInvalidRemoteName(t *testing.T) {
	cfg := engine.Config{
		ProjectName: "demo",
		Domain:      "demo.test",
		Remotes: map[string]engine.RemoteConfig{
			"!!!bad!!!": {
				Host: "example.com",
				User: "deploy",
				Path: "/srv/www/app",
				Port: 22,
			},
		},
	}
	engine.NormalizeConfig(&cfg, "")
	err := engine.ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected invalid remote name validation error")
	}
	if !strings.Contains(err.Error(), "not a valid identifier") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateConfigAcceptsCustomRemoteEnvironments(t *testing.T) {
	customNames := []string{"qa", "preprod", "demo", "client-uat", "integration", "load-test", "dr"}
	for _, name := range customNames {
		t.Run(name, func(t *testing.T) {
			cfg := engine.Config{
				ProjectName: "demo",
				Domain:      "demo.test",
				Stack: engine.Stack{
					Services: engine.Services{
						WebServer: "nginx",
						Search:    "none",
						Cache:     "none",
						Queue:     "none",
					},
				},
				Remotes: engine.RemoteConfigMap{
					name: {
						Host: "example.com",
						User: "deploy",
						Path: "/srv/www/app",
						Port: 22,
					},
				},
			}
			engine.NormalizeConfig(&cfg, "")
			if err := engine.ValidateConfig(cfg); err != nil {
				t.Fatalf("custom remote name %q should be valid, got error: %v", name, err)
			}
		})
	}
}

func TestIsValidRemoteName(t *testing.T) {
	valid := []string{
		"dev", "staging", "prod", "production",
		"qa", "uat", "preprod", "demo",
		"client-uat", "load-test", "dr",
		"env1", "my-env", "my_env",
	}
	for _, name := range valid {
		if !engine.IsValidRemoteName(name) {
			t.Errorf("IsValidRemoteName(%q) = false, want true", name)
		}
	}

	invalid := []string{
		"", "!!!bad!!!", "has space", "special@char", "dot.name",
	}
	for _, name := range invalid {
		if engine.IsValidRemoteName(name) {
			t.Errorf("IsValidRemoteName(%q) = true, want false", name)
		}
	}
}
