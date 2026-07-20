package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"govard/internal/cmd"
)

func TestNormalizeIntelephensePHPVersion(t *testing.T) {
	for _, tt := range []struct {
		in   string
		want string
	}{
		{in: "8.2", want: "8.2.0"},
		{in: "8.2.5", want: "8.2.5"},
		{in: "", want: ""},
		{in: "none", want: ""},
		{in: "  8.1  ", want: "8.1.0"},
	} {
		if got := cmd.NormalizeIntelephensePHPVersionForTest(tt.in); got != tt.want {
			t.Errorf("NormalizeIntelephensePHPVersionForTest(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestMergeJSONObjectFilePreservesUnrelatedKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"editor.tabSize": 4, "phpstan.binPath": "old"}`), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	err := cmd.MergeJSONObjectFileForTest(path,
		map[string]interface{}{"php.validate.executablePath": "/wrapper"},
		[]string{"phpstan.binPath"},
	)
	if err != nil {
		t.Fatalf("MergeJSONObjectFileForTest: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if obj["editor.tabSize"] != float64(4) {
		t.Errorf("expected unrelated key editor.tabSize to survive, got %v", obj["editor.tabSize"])
	}
	if _, exists := obj["phpstan.binPath"]; exists {
		t.Errorf("expected phpstan.binPath to be removed, still present: %v", obj["phpstan.binPath"])
	}
	if obj["php.validate.executablePath"] != "/wrapper" {
		t.Errorf("expected php.validate.executablePath to be set, got %v", obj["php.validate.executablePath"])
	}
}

func TestMergeJSONObjectFileCreatesMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "settings.json")

	if err := cmd.MergeJSONObjectFileForTest(path, map[string]interface{}{"a": "b"}, nil); err != nil {
		t.Fatalf("MergeJSONObjectFileForTest on missing file: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to be created: %v", err)
	}
}

func TestMergeLaunchConfigReplacesExistingEntryAndKeepsOthers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "launch.json")
	seed := `{
		"version": "0.2.0",
		"configurations": [
			{"name": "Some Other Config", "type": "node", "request": "launch"},
			{"name": "Listen for Xdebug (Govard)", "type": "php", "request": "launch", "port": 9000}
		]
	}`
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := cmd.MergeLaunchConfigForTest(path, "/var/www/html"); err != nil {
		t.Fatalf("MergeLaunchConfigForTest: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	var launch struct {
		Configurations []map[string]interface{} `json:"configurations"`
	}
	if err := json.Unmarshal(data, &launch); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if len(launch.Configurations) != 2 {
		t.Fatalf("expected 2 configurations, got %d", len(launch.Configurations))
	}

	var xdebug map[string]interface{}
	for _, c := range launch.Configurations {
		if c["name"] == "Some Other Config" {
			continue
		}
		xdebug = c
	}
	if xdebug == nil {
		t.Fatal("expected the Govard Xdebug configuration to still be present")
	}
	if xdebug["port"] != float64(9003) {
		t.Errorf("expected port to be updated to 9003, got %v", xdebug["port"])
	}
}
