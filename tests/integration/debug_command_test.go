//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestDebugStatusAndShellDisabled(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "debug-m2")

	statusResult := env.RunGovard(t, projectDir, "debug", "status")
	statusResult.AssertSuccess(t)
	assertContains(t, strings.ToLower(statusResult.Stdout+statusResult.Stderr), "xdebug is currently")

	shellResult := env.RunGovard(t, projectDir, "debug", "shell")
	shellResult.AssertExitCode(t, 1)
	assertContains(t, strings.ToLower(shellResult.Stdout+shellResult.Stderr), "xdebug is disabled")
}
