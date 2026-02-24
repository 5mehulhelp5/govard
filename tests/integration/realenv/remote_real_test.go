//go:build realenv
// +build realenv

package realenv

import (
	"testing"
)

func TestRemoteTestConnectionToDev(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)
	
	localDir := env.CreateTempProject(t, "local")
	
	result := env.RunGovard(t, localDir, "remote", "test", "dev")
	result.AssertSuccess(t)
}

func TestRemoteTestConnectionToStaging(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)
	
	localDir := env.CreateTempProject(t, "local")
	
	result := env.RunGovard(t, localDir, "remote", "test", "staging")
	result.AssertSuccess(t)
}

func TestRemoteTestAllEnvironments(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)
	
	localDir := env.CreateTempProject(t, "local")
	
	// Test both DEV and STAGING
	for _, remote := range []string{"dev", "staging"} {
		result := env.RunGovard(t, localDir, "remote", "test", remote)
		result.AssertSuccess(t)
	}
}

func TestRemoteTestInvalidRemote(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)
	
	localDir := env.CreateTempProject(t, "local")
	
	// Try to test a non-configured remote
	result := env.RunGovard(t, localDir, "remote", "test", "production")
	result.AssertFailure(t)
}

func TestRemoteExecOnDev(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)
	
	localDir := env.CreateTempProject(t, "local")
	
	result := env.RunGovard(t, localDir, "remote", "exec", "dev", "--", "pwd")
	result.AssertSuccess(t)
	result.AssertOutputContains(t, "/var/www/html")
}

func TestRemoteExecOnStaging(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)
	
	localDir := env.CreateTempProject(t, "local")
	
	result := env.RunGovard(t, localDir, "remote", "exec", "staging", "--", "echo", "test")
	result.AssertSuccess(t)
	result.AssertOutputContains(t, "test")
}

func TestRemoteAuditDev(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)
	
	localDir := env.CreateTempProject(t, "local")
	
	result := env.RunGovard(t, localDir, "remote", "audit", "dev")
	// May succeed or fail depending on implementation, but should not panic
	if result.ExitCode != 0 {
		t.Logf("Remote audit returned non-zero exit code (may be expected): %d", result.ExitCode)
	}
}
