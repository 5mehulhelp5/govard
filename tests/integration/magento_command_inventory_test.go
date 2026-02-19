//go:build integration
// +build integration

package integration

import "testing"

func TestMagentoRelevantCommandsPresent(t *testing.T) {
	env := NewTestEnvironment(t)

	result := env.RunGovard(t, t.TempDir(), "--help")
	result.AssertSuccess(t)

	required := []string{
		"sync",
		"bootstrap",
		"db",
		"configure",
		"deploy",
		"remote",
		"profile",
		"snapshot",
		"upgrade",
	}

	for _, name := range required {
		assertContains(t, result.Stdout, name)
	}
}
