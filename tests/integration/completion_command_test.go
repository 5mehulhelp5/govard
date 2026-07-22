//go:build integration
// +build integration

package integration

import (
	"testing"
)

func TestCompletionCommandIsAvailable(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "completion-m2")

	result := env.RunGovard(t, projectDir, "completion", "bash")
	result.AssertSuccess(t)
	assertContains(t, result.Stdout, "bash completion V2 for govard")
}
