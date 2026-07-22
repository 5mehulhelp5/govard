//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestDesktopCommandRuntimePaths(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "desktop-m2")

	t.Run("DesktopLaunchesBinaryFromPATH", func(t *testing.T) {
		shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})
		installRuntimeCommandShim(t, shim, "govard-desktop", 0)

		result := env.RunGovardWithEnv(t, projectDir, shim.Env(), "desktop")
		result.AssertSuccess(t)

		logs := shim.ReadLog(t)
		assertContains(t, logs, "govard-desktop|")
	})

	t.Run("DesktopLaunchesBinaryWithBackgroundFlag", func(t *testing.T) {
		shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})
		installRuntimeCommandShim(t, shim, "govard-desktop", 0)

		result := env.RunGovardWithEnv(t, projectDir, shim.Env(), "desktop", "--background")
		result.AssertSuccess(t)

		logs := shim.ReadLog(t)
		assertContains(t, logs, "govard-desktop|--background")
	})

	t.Run("DesktopDevUsesWailsWhenAvailable", func(t *testing.T) {
		shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})
		installRuntimeCommandShim(t, shim, "wails", 0)

		result := env.RunGovardWithEnv(t, projectDir, shim.Env(), "desktop", "--dev")
		result.AssertSuccess(t)

		logs := shim.ReadLog(t)
		if !strings.Contains(logs, "wails|dev -tags desktop") {
			t.Fatalf("expected 'wails|dev -tags desktop' in logs, got: %s\n\nstdout: %s\nstderr: %s", logs, result.Stdout, result.Stderr)
		}
	})
}
