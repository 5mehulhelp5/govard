//go:build realenv
// +build realenv

package realenv

import (
	"testing"
)

func TestOpenTargets(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)

	localDir := env.CreateTempProject(t, "local")

	// We test targets that are expected to print a URL and then try to open it.
	// Even if opening fails (no browser), we can check the output for the URL.
	targets := []string{"admin", "mail", "sftp", "elasticsearch", "opensearch"}

	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			result := env.RunGovard(t, localDir, "open", target)

			switch target {
			case "admin":
				result.AssertOutputContains(t, "Opening")
				result.AssertOutputContains(t, "admin")
			case "mail":
				result.AssertOutputContains(t, "Opening")
				result.AssertOutputContains(t, "mail.govard.test")
			case "sftp":
				// sftp requires a remote environment by default if no local sftp configured
				// Actually open_targets.go says "SFTP is not supported for local target"
				// For local SFTP, it prints an info message instead of "Opening"
				result.AssertOutputContains(t, "SFTP is not supported for local target")
			case "elasticsearch":
				result.AssertOutputContains(t, "Opening")
				result.AssertOutputContains(t, "elasticsearch.govard.test")
			case "opensearch":
				result.AssertOutputContains(t, "Opening")
				result.AssertOutputContains(t, "opensearch.govard.test")
			}
		})
	}
}

func TestOpenRemoteTargets(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)

	localDir := env.CreateTempProject(t, "local")

	t.Run("admin-dev", func(t *testing.T) {
		result := env.RunGovard(t, localDir, "open", "admin", "--environment", "dev")
		result.AssertOutputContains(t, "Opening")
		// Remote admin URL uses the remote host (localhost) without SSH port
		result.AssertOutputContains(t, "https://localhost/admin")
	})

	t.Run("sftp-staging", func(t *testing.T) {
		result := env.RunGovard(t, localDir, "open", "sftp", "--environment", "staging")
		result.AssertOutputContains(t, "Opening")
		result.AssertOutputContains(t, "sftp://linuxserver.io@localhost:9024")
	})
}
