//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSvcCommandsWithShims(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "proxy-m2")
	shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})
	homeDir := filepath.Join(projectDir, ".home")
	if err := os.MkdirAll(filepath.Join(homeDir, ".govard", "proxy"), 0o755); err != nil {
		t.Fatalf("failed to create global proxy directory: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(homeDir, ".govard", "proxy", "docker-compose.yml"),
		[]byte("services: {}\n"),
		0o644,
	); err != nil {
		t.Fatalf("failed to write global proxy compose fixture: %v", err)
	}
	envVars := append(shim.Env(), "HOME="+homeDir)

	downResult := env.RunGovardWithEnv(t, projectDir, envVars, "svc", "down")
	downResult.AssertSuccess(t)

	psResult := env.RunGovardWithEnv(t, projectDir, envVars, "svc", "ps")
	psResult.AssertSuccess(t)

	logsResult := env.RunGovardWithEnv(t, projectDir, envVars, "svc", "logs")
	logsResult.AssertSuccess(t)

	pullResult := env.RunGovardWithEnv(t, projectDir, envVars, "svc", "pull")
	pullResult.AssertSuccess(t)

	logs := shim.ReadLog(t)
	assertContains(t, logs, "docker|compose --project-directory")
	assertContains(t, logs, " -p proxy ")
	assertContains(t, logs, " down")
	assertContains(t, logs, " ps")
	assertContains(t, logs, " logs")
	assertContains(t, logs, " pull")
}
