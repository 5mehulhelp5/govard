//go:build integration
// +build integration

package integration

import (
	"fmt"
	"testing"

	"govard/internal/engine"
)

func TestFrameworkWrapperRecipeGuardsForMagentoProject(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "wrapper-guards-m2")

	tests := []struct {
		command        string
		recipeExpected string
	}{
		{command: "artisan", recipeExpected: "laravel"},
		{command: "cake", recipeExpected: "cakephp"},
		{command: "drush", recipeExpected: "drupal"},
		{command: "magerun", recipeExpected: "magento1"},
		{command: "shopware", recipeExpected: "shopware"},
		{command: "symfony", recipeExpected: "symfony"},
		{command: "wp", recipeExpected: "wordpress"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := env.RunGovard(t, projectDir, tt.command, "--version")
			if result.Success() {
				t.Fatalf("expected %s command to fail on magento2 project", tt.command)
			}
			output := result.Stdout + result.Stderr
			assertContains(t, output, fmt.Sprintf("the '%s' command is only available for %s projects", tt.command, tt.recipeExpected))
		})
	}
}

func TestGlobalWrapperCommandsUseMagentoExecUser(t *testing.T) {
	env := NewTestEnvironment(t)

	tests := []struct {
		command string
		arg     string
	}{
		{command: "npm", arg: "--version"},
		{command: "npx", arg: "--version"},
		{command: "pnpm", arg: "--version"},
		{command: "yarn", arg: "--version"},
		{command: "grunt", arg: "--version"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "wrapper-"+tt.command+"-m2")
			shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})

			result := env.RunGovardWithEnv(t, projectDir, shim.Env(), tt.command, tt.arg)
			result.AssertSuccess(t)

			config, _, err := engine.LoadConfigFromDir(projectDir, true)
			if err != nil {
				t.Fatalf("failed to load config: %v", err)
			}

			expectedUser := "www-data"
			if config.Stack.UserID > 0 && config.Stack.GroupID > 0 {
				expectedUser = fmt.Sprintf("%d:%d", config.Stack.UserID, config.Stack.GroupID)
			}

			logs := shim.ReadLog(t)
			assertContains(t, logs, "docker|exec -i -u "+expectedUser+" -w /var/www/html m2-clone-basic-php-1 "+tt.command+" "+tt.arg)
		})
	}
}

func TestCompletionCommandRendersShellScript(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "completion-m2")

	result := env.RunGovard(t, projectDir, "completion", "bash")
	result.AssertSuccess(t)

	assertContains(t, result.Stdout, "bash completion V2 for govard")
	assertContains(t, result.Stdout, "complete -o default -F __start_govard govard")
}
