package tests

import (
	"govard/internal/engine"
	"testing"
)

func TestResolveMagento2Profiles(t *testing.T) {
	tests := []struct {
		version          string
		expectedPHP      string
		expectedSearchV  string
		expectedCache    string
		expectedCacheV   string
		expectedDBV      string
		expectedNginxV   string
		expectedQueueV   string
		expectedVarnishV string
	}{
		{"2.4.9", "8.5", "3.0", "valkey", "9.0", "11.8", "1.28", "4.2", "8.0"},
		{"2.4.8-p2", "8.4", "3.0", "redis", "7.2", "11.4", "1.28", "3.13", "7.7"},
		{"2.4.8-p1", "8.4", "2.19", "redis", "7.2", "11.4", "1.26", "3.13", "7.6"},
		{"2.4.7-p7", "8.3", "2.19", "redis", "7.2", "10.11", "1.28", "3.13", "7.7"},
		{"2.4.7-p3", "8.3", "2.12", "redis", "7.2", "10.6", "1.24", "3.12", "7.5"},
		{"2.4.6-p11", "8.2", "2.19", "redis", "7.2", "10.11", "1.28", "3.13", "7.7"},
		{"2.4.6-p5", "8.2", "2.12", "redis", "7.0", "10.6", "1.24", "3.12", "7.1"},
		{"2.4.5-p16", "8.1", "2.19", "redis", "7.2", "10.4", "1.28", "3.13", "7.7"},
		{"2.4.5-p8", "8.1", "1.3", "redis", "7.2", "10.4", "1.24", "3.13", "7.5"},
		{"2.4.4-p17", "8.1", "2.19", "redis", "7.2", "10.4", "1.28", "3.13", "7.7"},
		{"2.4.4-p8", "8.1", "1.3", "redis", "7.0", "10.4", "1.24", "3.13", "7.5"},
		{"2.4.4", "8.1", "1.2", "redis", "6.2", "10.4", "1.20", "3.9", "7.0"},
		{"2.4.3-p2", "7.4", "1.2", "redis", "6.2", "10.4", "1.18", "3.8", "6.0"},
	}

	for _, tt := range tests {
		result, err := engine.ResolveRuntimeProfile("magento2", tt.version)
		if err != nil {
			t.Errorf("version %s: unexpected error: %v", tt.version, err)
			continue
		}

		profile := result.Profile
		if profile.PHPVersion != tt.expectedPHP {
			t.Errorf("version %s: expected PHP %s, got %s", tt.version, tt.expectedPHP, profile.PHPVersion)
		}
		if profile.SearchVersion != tt.expectedSearchV {
			t.Errorf("version %s: expected SearchVersion %s, got %s", tt.version, tt.expectedSearchV, profile.SearchVersion)
		}
		if profile.Cache != tt.expectedCache {
			t.Errorf("version %s: expected Cache %s, got %s", tt.version, tt.expectedCache, profile.Cache)
		}
		if profile.CacheVersion != tt.expectedCacheV {
			t.Errorf("version %s: expected CacheVersion %s, got %s", tt.version, tt.expectedCacheV, profile.CacheVersion)
		}
		if profile.DBVersion != tt.expectedDBV {
			t.Errorf("version %s: expected DBVersion %s, got %s", tt.version, tt.expectedDBV, profile.DBVersion)
		}
		if profile.NginxVersion != tt.expectedNginxV {
			t.Errorf("version %s: expected NginxVersion %s, got %s", tt.version, tt.expectedNginxV, profile.NginxVersion)
		}
		if profile.QueueVersion != tt.expectedQueueV {
			t.Errorf("version %s: expected QueueVersion %s, got %s", tt.version, tt.expectedQueueV, profile.QueueVersion)
		}
		if profile.VarnishVersion != tt.expectedVarnishV {
			t.Errorf("version %s: expected VarnishVersion %s, got %s", tt.version, tt.expectedVarnishV, profile.VarnishVersion)
		}
	}
}
