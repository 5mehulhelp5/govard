//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBootstrapInitAutoTune(t *testing.T) {
	env := NewTestEnvironment(t)

	t.Run("BootstrapAutoInitMagentoProject", func(t *testing.T) {
		// Create a directory with a Magento 2 composer.json
		projectDir := t.TempDir()
		composerJson := `{
			"name": "test/magento2-project",
			"require": {
				"magento/product-community-edition": "2.4.7"
			}
		}`
		if err := os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(composerJson), 0644); err != nil {
			t.Fatal(err)
		}

		shim := env.SetupRuntimeShims(t, map[string]int{
			"docker": 0,
		})

		// Run bootstrap with --yes and --fresh.
		// --yes triggers govard init --yes (Auto-Tune)
		result := env.RunGovardWithEnv(
			t,
			projectDir,
			append(shim.Env(), isolatedHomeEnv(t)...),
			"bootstrap",
			"--yes",
			"--fresh",
			"--skip-up",
			"--no-pii", // Dummy flag to bypass some logic, or just don't use non-existent flags
		)

		if !result.Success() {
			t.Logf("STDOUT: %s", result.Stdout)
			t.Logf("STDERR: %s", result.Stderr)
		}

		// Note: even if bootstrap fails later due to missing auth.json or composer,
		// we want to check if .govard.yml was created correctly by the Auto-Tune.

		// Verify .govard.yml was created
		if _, err := os.Stat(filepath.Join(projectDir, ".govard.yml")); err != nil {
			t.Fatalf(".govard.yml was not created by auto-tune bootstrap: %v. Output:\n%s", err, result.Stdout)
		}

		// Verify it detected as magento2
		configData, _ := os.ReadFile(filepath.Join(projectDir, ".govard.yml"))
		if !strings.Contains(string(configData), "framework: magento2") {
			t.Fatalf("expected framework: magento2 in .govard.yml, got:\n%s", string(configData))
		}
	})
}
