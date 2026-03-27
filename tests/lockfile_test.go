package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"govard/internal/engine"
)

func TestBuildLockFileFromConfigWithDependencies(t *testing.T) {
	projectDir := t.TempDir()
	composePath := engine.ComposeFilePath(projectDir, "demo")
	if err := osWriteFile(composePath, []byte(`services:
  web:
    image: nginx:1.27
  db:
    image: mariadb:11.4
`)); err != nil {
		t.Fatalf("write compose fixture: %v", err)
	}

	cfg := engine.Config{
		ProjectName:      "demo",
		Domain:           "demo.test",
		Framework:        "magento2",
		FrameworkVersion: "2.4.8-p3",
		Stack: engine.Stack{
			PHPVersion:    "8.4",
			NodeVersion:   "22",
			DBType:        "mariadb",
			DBVersion:     "11.4",
			CacheVersion:  "8.0.0",
			SearchVersion: "3.4.0",
			QueueVersion:  "3.13.7",
		},
	}

	deps := engine.LockDependencies{
		ReadDockerVersion:        func() (string, error) { return "27.2.1", nil },
		ReadDockerComposeVersion: func() (string, error) { return "2.29.7", nil },
		ReadServiceImages:        engine.ReadServiceImagesFromCompose,
		Now:                      func() time.Time { return time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC) },
	}

	lock, err := engine.BuildLockFileFromConfig(projectDir, cfg, "1.2.3", deps)
	if err != nil {
		t.Fatalf("build lock file: %v", err)
	}

	if lock.Project.Name != "demo" {
		t.Fatalf("expected project name demo, got %s", lock.Project.Name)
	}
	if lock.Govard.Version != "1.2.3" {
		t.Fatalf("expected govard version 1.2.3, got %s", lock.Govard.Version)
	}
	if lock.Host.DockerVersion != "27.2.1" {
		t.Fatalf("expected docker version 27.2.1, got %s", lock.Host.DockerVersion)
	}
	if lock.Services["web"] != "nginx:1.27" {
		t.Fatalf("expected web image nginx:1.27, got %s", lock.Services["web"])
	}
}

func TestWriteReadLockFileRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "govard.lock")
	original := engine.LockFile{
		Version:     1,
		GeneratedAt: "2026-02-20T12:00:00Z",
		Govard: engine.LockGovardInfo{
			Version: "1.2.3",
		},
		Host: engine.LockHostInfo{
			OS:                   "linux",
			Arch:                 "amd64",
			DockerVersion:        "27.2.1",
			DockerComposeVersion: "2.29.7",
		},
		Project: engine.LockProjectInfo{
			Name: "demo",
		},
		Services: map[string]string{"web": "nginx:1.27"},
	}

	if err := engine.WriteLockFile(path, original); err != nil {
		t.Fatalf("write lock file: %v", err)
	}
	parsed, err := engine.ReadLockFile(path)
	if err != nil {
		t.Fatalf("read lock file: %v", err)
	}
	if parsed.Govard.Version != "1.2.3" {
		t.Fatalf("expected govard version 1.2.3, got %s", parsed.Govard.Version)
	}
	if parsed.Services["web"] != "nginx:1.27" {
		t.Fatalf("expected web image nginx:1.27, got %s", parsed.Services["web"])
	}
}

func TestCompareLockFileDetectsMismatches(t *testing.T) {
	expected := engine.LockFile{
		Govard:   engine.LockGovardInfo{Version: "1.2.3"},
		Host:     engine.LockHostInfo{DockerVersion: "27.2.1", DockerComposeVersion: "2.29.7"},
		Project:  engine.LockProjectInfo{Framework: "magento2"},
		Stack:    engine.LockStackInfo{PHPVersion: "8.4", DBVersion: "11.4"},
		Services: map[string]string{"db": "mariadb:11.4"},
	}
	current := expected
	current.Govard.Version = "1.2.4"
	current.Host.DockerComposeVersion = "2.30.0"
	current.Stack.DBVersion = "11.5"
	current.Services = map[string]string{"db": "mariadb:11.5"}

	result := engine.CompareLockFile(expected, current, nil)
	if result.Compliant {
		t.Fatal("expected non-compliant result")
	}
	joined := strings.Join(result.Mismatches, "\n")
	for _, token := range []string{"govard.version", "host.docker_compose_version", "stack.db_version", "services.db"} {
		if !strings.Contains(joined, token) {
			t.Fatalf("expected mismatch token %q in %q", token, joined)
		}
	}
}

func TestCompareLockFileRespectsIgnoreFields(t *testing.T) {
	expected := engine.LockFile{
		Govard:  engine.LockGovardInfo{Version: "1.2.3"},
		Host:    engine.LockHostInfo{DockerVersion: "27.2.1", DockerComposeVersion: "2.29.7"},
		Project: engine.LockProjectInfo{Framework: "magento2"},
		Stack:   engine.LockStackInfo{PHPVersion: "8.4", DBVersion: "11.4"},
	}
	current := expected
	current.Host.DockerVersion = "28.0.0"
	current.Host.DockerComposeVersion = "2.30.0"

	// Without ignore — should have mismatches
	result := engine.CompareLockFile(expected, current, nil)
	if result.Compliant {
		t.Fatal("expected non-compliant without ignore")
	}

	// With ignore — should be compliant
	ignore := []string{"host.docker_version", "host.docker_compose_version"}
	result = engine.CompareLockFile(expected, current, ignore)
	if !result.Compliant {
		t.Fatalf("expected compliant with ignore fields, got mismatches: %v", result.Mismatches)
	}
}

func TestCompareLockFileDetectsDomainMismatches(t *testing.T) {
	expected := engine.LockFile{
		Project: engine.LockProjectInfo{
			Domain:       "main.test",
			ExtraDomains: []string{"extra1.test", "extra2.test"},
			StoreDomains: map[string]string{"fr": "fr.test", "de": "de.test"},
		},
	}
	current := expected

	// 1. Mismatch in extra_domains (order change should be OK but value change not)
	current.Project.ExtraDomains = []string{"extra1.test", "extra3.test"}
	result := engine.CompareLockFile(expected, current, nil)
	if result.Compliant {
		t.Fatal("expected non-compliant for extra_domains mismatch")
	}

	// 2. Mismatch in store_domains
	current = expected
	current.Project.StoreDomains = map[string]string{"fr": "fr.test", "de": "changed.test"}
	result = engine.CompareLockFile(expected, current, nil)
	if result.Compliant {
		t.Fatal("expected non-compliant for store_domains mismatch")
	}
}

func osWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := osMkdirAll(dir); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func osMkdirAll(path string) error {
	return os.MkdirAll(path, 0o755)
}
