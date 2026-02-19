package tests

import (
	"strings"
	"testing"

	"govard/internal/engine"
)

func TestNormalizeRemoteEnvironment(t *testing.T) {
	cases := map[string]string{
		"":            "staging",
		"production":  "prod",
		"Prod":        "prod",
		"development": "dev",
		"qa":          "staging",
	}

	for in, expected := range cases {
		if got := engine.NormalizeRemoteEnvironment(in); got != expected {
			t.Fatalf("normalize env %q: expected %q, got %q", in, expected, got)
		}
	}
}

func TestParseRemoteCapabilitiesCSV(t *testing.T) {
	all, err := engine.ParseRemoteCapabilitiesCSV("all")
	if err != nil {
		t.Fatalf("parse all capabilities: %v", err)
	}
	if !all.Files || !all.Media || !all.DB || !all.Deploy {
		t.Fatalf("expected all capabilities enabled, got %+v", all)
	}

	custom, err := engine.ParseRemoteCapabilitiesCSV("files,db")
	if err != nil {
		t.Fatalf("parse custom capabilities: %v", err)
	}
	if !custom.Files || custom.Media || !custom.DB || custom.Deploy {
		t.Fatalf("unexpected custom capabilities: %+v", custom)
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

func TestParseRemoteCapabilitiesRejectsEmptySet(t *testing.T) {
	_, err := engine.ParseRemoteCapabilitiesCSV(",")
	if err == nil {
		t.Fatal("expected empty capability set error")
	}
	if !strings.Contains(err.Error(), "at least one remote capability is required") {
		t.Fatalf("unexpected parse error: %v", err)
	}
}

func TestRemoteWriteBlocked(t *testing.T) {
	explicit := engine.RemoteConfig{
		Environment: "staging",
		Protected:   true,
	}
	if blocked, _ := engine.RemoteWriteBlocked(explicit); !blocked {
		t.Fatal("expected explicit protected remote to block writes")
	}

	prod := engine.RemoteConfig{
		Environment: "prod",
	}
	if blocked, _ := engine.RemoteWriteBlocked(prod); !blocked {
		t.Fatal("expected production remote to block writes")
	}

	dev := engine.RemoteConfig{
		Environment: "dev",
	}
	if blocked, _ := engine.RemoteWriteBlocked(dev); blocked {
		t.Fatal("expected dev remote writes to be allowed")
	}
}

func TestValidateConfigRejectsInvalidRemoteEnvironment(t *testing.T) {
	cfg := engine.Config{
		ProjectName: "demo",
		Domain:      "demo.test",
		Remotes: map[string]engine.RemoteConfig{
			"bad": {
				Host:        "example.com",
				User:        "deploy",
				Path:        "/srv/www/app",
				Port:        22,
				Environment: "moon",
			},
		},
	}
	engine.NormalizeConfig(&cfg)
	if err := engine.ValidateConfig(cfg); err == nil {
		t.Fatal("expected invalid environment validation error")
	}
}
