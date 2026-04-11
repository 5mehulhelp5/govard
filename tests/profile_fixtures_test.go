package tests

import (
	"strings"
	"testing"

	"govard/internal/engine"
)

func TestRuntimeProfileFixtures(t *testing.T) {
	fixtures := engine.GetFrameworkTestFixtures()

	for _, fixture := range fixtures {
		fixture := fixture
		name := strings.TrimSpace(fixture.Name)
		if name == "" {
			name = fixture.Framework + "@" + fixture.Version
		}
		t.Run(name, func(t *testing.T) {
			result, err := engine.ResolveRuntimeProfile(fixture.Framework, fixture.Version)
			if fixture.ExpectError {
				if err == nil {
					t.Fatalf("expected error for %s@%s", fixture.Framework, fixture.Version)
				}
				return
			}

			if err != nil {
				t.Fatalf("resolve profile: %v", err)
			}

			if fixture.Source != "" && result.Source != fixture.Source {
				t.Fatalf("expected source %q, got %q", fixture.Source, result.Source)
			}
			if fixture.SourcePrefix != "" && !strings.HasPrefix(result.Source, fixture.SourcePrefix) {
				t.Fatalf("expected source prefix %q, got %q", fixture.SourcePrefix, result.Source)
			}
			if fixture.WarningContains != "" {
				found := false
				for _, warning := range result.Warnings {
					if strings.Contains(warning, fixture.WarningContains) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected warning containing %q, got %v", fixture.WarningContains, result.Warnings)
				}
			}

			for key, want := range fixture.Expected {
				got, ok := profileFieldValue(result.Profile, key)
				if !ok {
					t.Fatalf("unknown expected field key %q", key)
				}
				if got != want {
					t.Fatalf("field %q: expected %q, got %q", key, want, got)
				}
			}
		})
	}
}

func profileFieldValue(profile engine.RuntimeProfile, key string) (string, bool) {
	switch key {
	case "framework":
		return profile.Framework, true
	case "framework_version":
		return profile.FrameworkVersion, true
	case "php_version":
		return profile.PHPVersion, true
	case "node_version":
		return profile.NodeVersion, true
	case "db_type":
		return profile.DBType, true
	case "db_version":
		return profile.DBVersion, true
	case "web_root":
		return profile.WebRoot, true
	case "web_server":
		return profile.WebServer, true
	case "cache":
		return profile.Cache, true
	case "cache_version":
		return profile.CacheVersion, true
	case "search":
		return profile.Search, true
	case "search_version":
		return profile.SearchVersion, true
	case "queue":
		return profile.Queue, true
	case "queue_version":
		return profile.QueueVersion, true
	case "xdebug_session":
		return profile.XdebugSession, true
	default:
		return "", false
	}
}
