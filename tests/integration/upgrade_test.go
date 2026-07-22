//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpgradeCommandMessageFallback(t *testing.T) {
	env := NewTestEnvironment(t)
	// Create a project with a framework not yet supported for upgrade
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "upgrade-fallback")

	// Override framework to something unsupported
	configPath := filepath.Join(projectDir, ".govard.yml")
	content, _ := os.ReadFile(configPath)
	newContent := strings.Replace(string(content), "framework: magento2", "framework: unknown-framework", 1)
	_ = os.WriteFile(configPath, []byte(newContent), 0644)

	result := env.RunGovard(t, projectDir, "upgrade")
	result.AssertSuccess(t)
	assertContains(t, strings.ToLower(result.Stdout+result.Stderr), "not implemented yet")
}

func TestUpgradeMagentoDryRun(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "upgrade-dryrun")

	result := env.RunGovard(t, projectDir, "upgrade", "--version=2.4.8-p4", "--dry-run")
	result.AssertSuccess(t)
	assertContains(t, result.Stdout, "Target version: 2.4.8-p4")
	assertContains(t, result.Stdout, "[DRY RUN] Would perform the following steps:")
	assertContains(t, result.Stdout, "Update .govard.yml configuration")
}
