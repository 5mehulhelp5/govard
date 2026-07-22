//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProfileCommandJSONAndApply(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "profile-m2")

	jsonResult := env.RunGovard(t, projectDir, "config", "profile", "--json")
	jsonResult.AssertSuccess(t)
	assertContains(t, jsonResult.Stdout, `"framework": "magento2"`)
	assertContains(t, jsonResult.Stdout, `"selected"`)

	applyResult := env.RunGovard(t, projectDir, "config", "profile", "apply")
	applyResult.AssertSuccess(t)

	configBytes, err := os.ReadFile(filepath.Join(projectDir, ".govard.yml"))
	if err != nil {
		t.Fatalf("failed to read .govard.yml: %v", err)
	}
	config := string(configBytes)
	assertContains(t, config, "project_name: m2-clone-basic")
	assertContains(t, config, "framework_version: 2.4.8")
}
