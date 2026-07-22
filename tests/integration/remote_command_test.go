//go:build integration
// +build integration

package integration

import (
	"testing"
)

func TestRemoteCommandRuntimeWithShims(t *testing.T) {
	env := NewTestEnvironment(t)

	t.Run("RemoteTestUsesSSHChecks", func(t *testing.T) {
		projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "remote-test-shims")
		shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})

		result := env.RunGovardWithEnv(t, projectDir, shim.Env(), "remote", "test", "dev")
		result.AssertSuccess(t)

		logs := shim.ReadLog(t)
		assertContains(t, logs, "ssh|")
		assertContains(t, logs, "govard-remote-ok")
		assertContains(t, logs, "govard-rsync-ok")
	})

	t.Run("RemoteExecUsesSSH", func(t *testing.T) {
		projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "remote-exec-shims")
		shim := env.SetupRuntimeShims(t, map[string]int{"docker": 0, "ssh": 0, "rsync": 0})

		result := env.RunGovardWithEnv(t, projectDir, shim.Env(), "remote", "exec", "dev", "--", "pwd")
		result.AssertSuccess(t)

		logs := shim.ReadLog(t)
		assertContains(t, logs, "ssh|")
		assertContains(t, logs, "pwd")
	})
}
