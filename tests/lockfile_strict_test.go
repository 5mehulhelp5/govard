package tests

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"govard/internal/cmd"
	"govard/internal/engine"
)

func testLockStrictConfig() engine.Config {
	return engine.Config{
		ProjectName: "demo",
		Recipe:      "magento2",
		Domain:      "demo.test",
		Stack: engine.Stack{
			PHPVersion:    "8.4",
			NodeVersion:   "24",
			DBType:        "mariadb",
			DBVersion:     "11.4",
			CacheVersion:  "8.0.0",
			SearchVersion: "2.19.0",
			QueueVersion:  "",
		},
	}
}

func withDeterministicLockDeps(t *testing.T) func() {
	t.Helper()
	return cmd.SetLockDependenciesForTest(engine.LockDependencies{
		ReadDockerVersion:        func() (string, error) { return "27.2.1", nil },
		ReadDockerComposeVersion: func() (string, error) { return "2.29.7", nil },
		ReadServiceImages: func(composePath string) (map[string]string, error) {
			_ = composePath
			return map[string]string{"web": "nginx:1.27"}, nil
		},
		Now: func() time.Time { return time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC) },
	})
}

func TestEvaluateUpLockPolicyStrictDisabledDoesNotBlockOnMismatch(t *testing.T) {
	tempDir := t.TempDir()
	config := testLockStrictConfig()

	restore := withDeterministicLockDeps(t)
	defer restore()

	lock, err := engine.BuildLockFileFromConfig(tempDir, config, "1.0.0", engine.LockDependencies{
		ReadDockerVersion:        func() (string, error) { return "27.2.1", nil },
		ReadDockerComposeVersion: func() (string, error) { return "2.29.7", nil },
		ReadServiceImages: func(composePath string) (map[string]string, error) {
			return map[string]string{"web": "nginx:1.27"}, nil
		},
		Now: func() time.Time { return time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("build lock fixture: %v", err)
	}
	lock.Host.DockerComposeVersion = "2.30.0"
	if err := engine.WriteLockFile(filepath.Join(tempDir, "govard.lock"), lock); err != nil {
		t.Fatalf("write lock fixture: %v", err)
	}

	warnings, err := cmd.EvaluateUpLockPolicyForTest(tempDir, config)
	if err != nil {
		t.Fatalf("expected no strict error, got: %v", err)
	}
	if len(warnings) == 0 {
		t.Fatal("expected lock mismatch warnings")
	}
}

func TestEvaluateUpLockPolicyStrictEnabledBlocksOnMismatch(t *testing.T) {
	tempDir := t.TempDir()
	config := testLockStrictConfig()
	config.Lock.Strict = true

	restore := withDeterministicLockDeps(t)
	defer restore()

	lock, err := engine.BuildLockFileFromConfig(tempDir, config, "1.0.0", engine.LockDependencies{
		ReadDockerVersion:        func() (string, error) { return "27.2.1", nil },
		ReadDockerComposeVersion: func() (string, error) { return "2.29.7", nil },
		ReadServiceImages: func(composePath string) (map[string]string, error) {
			return map[string]string{"web": "nginx:1.27"}, nil
		},
		Now: func() time.Time { return time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("build lock fixture: %v", err)
	}
	lock.Project.Domain = "changed.test"
	if err := engine.WriteLockFile(filepath.Join(tempDir, "govard.lock"), lock); err != nil {
		t.Fatalf("write lock fixture: %v", err)
	}

	warnings, err := cmd.EvaluateUpLockPolicyForTest(tempDir, config)
	if err == nil {
		t.Fatal("expected strict mode mismatch error")
	}
	if len(warnings) == 0 {
		t.Fatal("expected mismatch warnings in strict mode")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "strict") {
		t.Fatalf("expected strict-mode error hint, got: %v", err)
	}
}

func TestEvaluateUpLockPolicyStrictEnabledRequiresLockFile(t *testing.T) {
	tempDir := t.TempDir()
	config := testLockStrictConfig()
	config.Lock.Strict = true

	restore := withDeterministicLockDeps(t)
	defer restore()

	warnings, err := cmd.EvaluateUpLockPolicyForTest(tempDir, config)
	if err == nil {
		t.Fatal("expected strict mode error when lock file is missing")
	}
	if len(warnings) == 0 {
		t.Fatal("expected warning guidance when lock file is missing")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "lock") {
		t.Fatalf("expected lock-file guidance, got: %v", err)
	}
}

func TestEvaluateUpLockPolicyStrictEnabledPassesWhenCompliant(t *testing.T) {
	tempDir := t.TempDir()
	config := testLockStrictConfig()
	config.Lock.Strict = true

	restore := withDeterministicLockDeps(t)
	defer restore()

	lock, err := engine.BuildLockFileFromConfig(tempDir, config, cmd.Version, engine.LockDependencies{
		ReadDockerVersion:        func() (string, error) { return "27.2.1", nil },
		ReadDockerComposeVersion: func() (string, error) { return "2.29.7", nil },
		ReadServiceImages: func(composePath string) (map[string]string, error) {
			return map[string]string{"web": "nginx:1.27"}, nil
		},
		Now: func() time.Time { return time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("build lock fixture: %v", err)
	}
	if err := engine.WriteLockFile(filepath.Join(tempDir, "govard.lock"), lock); err != nil {
		t.Fatalf("write lock fixture: %v", err)
	}

	warnings, err := cmd.EvaluateUpLockPolicyForTest(tempDir, config)
	if err != nil {
		t.Fatalf("expected compliant strict mode to pass, got: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for compliant lock, got: %v", warnings)
	}
}
