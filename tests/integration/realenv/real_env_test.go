//go:build realenv
// +build realenv

// Package realenv provides integration tests that run against real Docker environments.
// These tests require the three-environment setup to be running.
//
// To run these tests:
//  1. cd tests/integration/environments && ./setup-three-env.sh
//  2. go test -tags realenv ./tests/integration/realenv/... -v
//
// To clean up:
//
//	cd tests/integration/environments && ./setup-three-env.sh cleanup
package realenv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// Auth represents SSH authentication configuration
type Auth struct {
	Method         string `yaml:"method,omitempty"`
	KeyPath        string `yaml:"key_path,omitempty"`
	StrictHostKey  bool   `yaml:"strict_host_key,omitempty"`
	KnownHostsFile string `yaml:"known_hosts_file,omitempty"`
}

// Remote represents a remote environment configuration
type Remote struct {
	Host         string       `yaml:"host"`
	Port         int          `yaml:"port,omitempty"`
	User         string       `yaml:"user"`
	Path         string       `yaml:"path"`
	Environment  string       `yaml:"environment,omitempty"`
	Protected    bool         `yaml:"protected,omitempty"`
	Auth         Auth         `yaml:"auth,omitempty"`
	Capabilities Capabilities `yaml:"capabilities"`
}

// Capabilities represents remote capabilities
type Capabilities struct {
	Files  bool `yaml:"files"`
	Media  bool `yaml:"media"`
	DB     bool `yaml:"db"`
	Deploy bool `yaml:"deploy"`
}

// RealEnvTest provides infrastructure for real environment tests
type RealEnvTest struct {
	SSHKeyPath  string
	ProjectRoot string
	EnvDir      string
	FixturesDir string
	BinaryPath  string
}

var (
	buildOnce sync.Once
	buildErr  error
)

// NewRealEnvTest creates a new real environment test harness
func NewRealEnvTest(t *testing.T) *RealEnvTest {
	t.Helper()

	projectRoot := findProjectRoot(t)

	return &RealEnvTest{
		SSHKeyPath:  filepath.Join(projectRoot, "tests/integration/realenv/.ssh/id_rsa"),
		ProjectRoot: projectRoot,
		EnvDir:      filepath.Join(projectRoot, "tests/integration/realenv"),
		FixturesDir: filepath.Join(projectRoot, "tests/integration/projects/magento2"),
		BinaryPath:  filepath.Join(projectRoot, "bin/govard-test"),
	}
}

// Setup ensures the three-environment setup is running
func (r *RealEnvTest) Setup(t *testing.T) {
	t.Helper()

	// Check if environment is running
	if !r.isEnvironmentRunning() {
		t.Skip("Three-environment setup not running. Run: cd tests/integration/realenv && ./setup-three-env.sh")
	}

	// Ensure binary exists
	buildOnce.Do(func() {
		buildErr = r.buildBinary()
	})

	if buildErr != nil {
		t.Fatalf("Failed to build govard binary: %v", buildErr)
	}
}

func (r *RealEnvTest) isEnvironmentRunning() bool {
	containers := []string{
		"m2-clone-basic-db-1",
		"govard-test-dev-db",
		"govard-test-staging-db",
	}

	for _, container := range containers {
		cmd := exec.Command("docker", "ps", "-q", "-f", fmt.Sprintf("name=%s", container), "-f", "status=running")
		output, err := cmd.Output()
		if err != nil || len(output) == 0 {
			return false
		}
	}
	return true
}

func (r *RealEnvTest) buildBinary() error {
	if err := os.MkdirAll(filepath.Dir(r.BinaryPath), 0755); err != nil {
		return fmt.Errorf("create bin dir: %w", err)
	}

	cmd := exec.Command("go", "build", "-mod=mod", "-o", r.BinaryPath, "-tags", "integration",
		filepath.Join(r.ProjectRoot, "cmd/govard/main.go"))
	cmd.Dir = r.ProjectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build govard binary: %w: %s", err, string(output))
	}
	return nil
}

// RunGovard executes govard with given arguments in the specified project directory
func (r *RealEnvTest) RunGovard(t *testing.T, projectDir string, args ...string) *RealEnvResult {
	t.Helper()

	cmd := exec.Command(r.BinaryPath, args...)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(),
		"GOVARD_TEST_MODE=true",
		fmt.Sprintf("SSH_AUTH_SOCK=%s", os.Getenv("SSH_AUTH_SOCK")),
	)

	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	return &RealEnvResult{
		Output:   string(output),
		ExitCode: cmd.ProcessState.ExitCode(),
		Duration: duration,
		Error:    err,
	}
}

// CopyConfig copies the appropriate .govard.yml to project directory
// Uses existing fixtures from tests/integration/projects/magento2/
// For real env tests, only 'local' fixture should be used as the project.
// DEV and STAGING are remote targets, not local projects.
func (r *RealEnvTest) CopyConfig(t *testing.T, env string, projectDir string) {
	t.Helper()

	// For real env tests, we always use the LOCAL workstation as the project
	// DEV and STAGING are remote environments, not local projects
	if env != "local" {
		t.Fatalf("For real environment tests, only 'local' should be used as project. " +
			"DEV and STAGING are remote targets, not local projects. " +
			"Use 'local' as project and specify remote via --environment flag")
	}

	dst := filepath.Join(projectDir, ".govard.yml")
	// Build a completely new config structure instead of parsing
	// This ensures correct YAML formatting
	config := struct {
		ProjectName string            `yaml:"project_name"`
		Domain      string            `yaml:"domain"`
		Framework   string            `yaml:"framework"`
		Remotes     map[string]Remote `yaml:"remotes"`
	}{
		ProjectName: "m2-clone-basic",
		Domain:      "m2-clone-basic.test",
		Framework:   "magento2",
		Remotes: map[string]Remote{
			"dev": {
				Host: "localhost",
				Port: 9023,
				User: "linuxserver.io",
				Path: "/var/www/html",
				Auth: Auth{
					KeyPath: r.SSHKeyPath,
				},
				Capabilities: Capabilities{
					Files:  true,
					Media:  true,
					DB:     true,
					Deploy: false,
				},
			},
			"staging": {
				Host: "localhost",
				Port: 9024,
				User: "linuxserver.io",
				Path: "/var/www/html",
				Auth: Auth{
					KeyPath: r.SSHKeyPath,
				},
				Capabilities: Capabilities{
					Files:  true,
					Media:  true,
					DB:     true,
					Deploy: false,
				},
			},
			"production": {
				Host:        "localhost",
				Port:        9025,
				User:        "linuxserver.io",
				Path:        "/var/www/html",
				Environment: "production",
				Protected:   true,
				Auth: Auth{
					KeyPath: r.SSHKeyPath,
				},
				Capabilities: Capabilities{
					Files:  true,
					Media:  true,
					DB:     true,
					Deploy: true,
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(dst, yamlData, 0644); err != nil {
		t.Fatalf("Failed to write .govard.yml: %v", err)
	}
	t.Logf("Generated .govard.yml in %s:\n%s", projectDir, string(yamlData))
}

// CreateTempProject creates a temporary project directory with config
func (r *RealEnvTest) CreateTempProject(t *testing.T, env string) string {
	t.Helper()

	dir := t.TempDir()
	r.CopyConfig(t, env, dir)
	return dir
}

// RealEnvResult holds test execution results
type RealEnvResult struct {
	Output   string
	ExitCode int
	Duration time.Duration
	Error    error
}

// AssertSuccess fails the test if command didn't succeed
func (r *RealEnvResult) AssertSuccess(t *testing.T) {
	t.Helper()
	if r.Error != nil || r.ExitCode != 0 {
		t.Fatalf("Command failed with exit code %d\nError: %v\nOutput:\n%s",
			r.ExitCode, r.Error, r.Output)
	}
}

// AssertFailure fails the test if command succeeded
func (r *RealEnvResult) AssertFailure(t *testing.T) {
	t.Helper()
	if r.Error == nil && r.ExitCode == 0 {
		t.Fatalf("Expected command to fail, but it succeeded\nOutput:\n%s", r.Output)
	}
}

// AssertOutputContains fails if output doesn't contain expected string
func (r *RealEnvResult) AssertOutputContains(t *testing.T, expected string) {
	t.Helper()
	if !contains(r.Output, expected) {
		t.Fatalf("Expected output to contain %q\n\nActual output:\n%s", expected, r.Output)
	}
}

// AssertOutputNotContains fails if output contains unexpected string
func (r *RealEnvResult) AssertOutputNotContains(t *testing.T, unexpected string) {
	t.Helper()
	if contains(r.Output, unexpected) {
		t.Fatalf("Expected output NOT to contain %q\n\nActual output:\n%s", unexpected, r.Output)
	}
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(findSubstring(s, substr) >= 0)
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func findProjectRoot(t *testing.T) string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Fatal("Could not find project root (no go.mod found)")
	return ""
}
