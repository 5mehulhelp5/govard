//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestBootstrapSymfonyFreshIntegration(t *testing.T) {
	env := NewTestEnvironment(t)
	// Symfony bootstrap expects an empty dir or it will clean it up.
	// But it keeps .govard.yml
	projectDir := env.CreateTestProject(t, "symfony-fresh", map[string]string{
		".govard.yml": MustMarshalYAML(t, map[string]interface{}{
			"project_name": "symfony-test",
			"framework":    "symfony",
			"domain":       "symfony-test.test",
		}),
	})

	shim := env.SetupRuntimeShims(t, map[string]int{
		"docker": 0,
	})

	result := env.RunGovardWithEnv(
		t,
		projectDir,
		shim.Env(),
		"bootstrap",
		"--fresh", "--yes",
		"--skip-up",
	)
	result.AssertSuccess(t)

	logs := shim.ReadLog(t)

	// Check for composer create-project in docker exec
	if !strings.Contains(logs, "docker|exec") || !strings.Contains(logs, "composer create-project") {
		t.Errorf("Expected docker exec composer create-project in logs, got:\n%s", logs)
	}
}

func TestBootstrapLaravelFreshIntegration(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateTestProject(t, "laravel-fresh", map[string]string{
		".govard.yml": MustMarshalYAML(t, map[string]interface{}{
			"project_name": "laravel-test",
			"framework":    "laravel",
			"domain":       "laravel-test.test",
		}),
	})

	shim := env.SetupRuntimeShims(t, map[string]int{
		"docker": 0,
	})

	result := env.RunGovardWithEnv(
		t,
		projectDir,
		shim.Env(),
		"bootstrap",
		"--fresh", "--yes",
		"--skip-up",
	)
	result.AssertSuccess(t)

	logs := shim.ReadLog(t)

	// Check for composer create-project in docker exec
	if !strings.Contains(logs, "docker|exec") || !strings.Contains(logs, "composer create-project") {
		t.Errorf("Expected docker exec composer create-project in logs, got:\n%s", logs)
	}
}
