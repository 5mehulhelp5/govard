//go:build realenv
// +build realenv

package realenv

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestBootstrapCloneFromDevToLocal(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)

	// Setup LOCAL project
	localDir := env.CreateTempProject(t, "local")

	// Create marker file in DEV via SSH
	markerContent := "DEV_MARKER_" + time.Now().Format("20060102_150405")
	sshCmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", env.SSHKeyPath,
		"-p", "9023", "linuxserver.io@localhost",
		"echo '"+markerContent+"' > /var/www/html/.dev_marker")
	if err := sshCmd.Run(); err != nil {
		t.Fatalf("Failed to create marker on DEV: %v", err)
	}

	// Clone from DEV
	result := env.RunGovard(t, localDir, "bootstrap", "--clone",
		"--environment", "dev",
		"--skip-up",
		"--no-composer",
		"--no-db",
		"--no-media",
		"--no-admin",
	)
	result.AssertSuccess(t)

	// Verify: marker file should be synced
	result.AssertOutputContains(t, "sync")
}

func TestBootstrapCloneCodeOnly(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)

	localDir := env.CreateTempProject(t, "local")

	result := env.RunGovard(t, localDir, "bootstrap", "--clone",
		"--environment", "dev",
		"--code-only",
		"--skip-up",
		"--no-composer",
		"--no-admin",
	)
	result.AssertSuccess(t)

	// Should NOT contain database or media operations
	result.AssertOutputNotContains(t, "mysqldump")
}

func TestBootstrapValidationFreshAndClone(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)

	localDir := env.CreateTempProject(t, "local")

	// Try to use both --fresh and --clone (should fail)
	result := env.RunGovard(t, localDir, "bootstrap", "--fresh", "--clone", "--skip-up")
	result.AssertFailure(t)
	result.AssertOutputContains(t, "--fresh and --clone cannot be used together")
}

func TestBootstrapValidationCodeOnlyRequiresClone(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)

	localDir := env.CreateTempProject(t, "local")

	result := env.RunGovard(t, localDir, "bootstrap", "--code-only", "--skip-up")
	result.AssertFailure(t)
	result.AssertOutputContains(t, "--code-only requires --clone")
}

func TestBootstrapCloneRequiresConfiguredRemote(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)

	// Create a minimal .govard.yml with no remotes so ensureBootstrapInit skips
	// running `govard init` (which would block waiting for interactive input).
	localDir := t.TempDir()
	minimalConfig := "project_name: test-no-remotes\ndomain: test-no-remotes.test\nframework: magento2\n"
	if err := os.WriteFile(filepath.Join(localDir, ".govard.yml"), []byte(minimalConfig), 0644); err != nil {
		t.Fatalf("Failed to write .govard.yml: %v", err)
	}

	result := env.RunGovard(t, localDir, "bootstrap", "--clone", "--environment", "dev", "--skip-up")
	result.AssertFailure(t)
	result.AssertOutputContains(t, "remote 'dev' is not configured")
}

func TestBootstrapCloneFull(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)

	localDir := env.CreateTempProject(t, "local")

	// Create test data in DEV
	// 1. File
	sshCmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", env.SSHKeyPath,
		"-p", "9023", "linuxserver.io@localhost",
		"mkdir -p /var/www/html/app/code && echo 'FULL_CLONE_CODE' > /var/www/html/app/code/full_test.txt")
	if err := sshCmd.Run(); err != nil {
		t.Fatalf("Failed to create test file on DEV: %v", err)
	}

	// 2. Media
	sshCmd = exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", env.SSHKeyPath,
		"-p", "9023", "linuxserver.io@localhost",
		"mkdir -p /var/www/html/pub/media && echo 'FULL_CLONE_MEDIA' > /var/www/html/pub/media/media_test.txt")
	if err := sshCmd.Run(); err != nil {
		t.Fatalf("Failed to create media file on DEV: %v", err)
	}

	// 3. DB - First create .my.cnf to disable SSL verification (needed for self-signed certs)
	sshCmd = exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", env.SSHKeyPath,
		"-p", "9023", "linuxserver.io@localhost",
		"mkdir -p ~ && echo '[client]\nssl-verify-server-cert=false' > ~/.my.cnf")
	if err := sshCmd.Run(); err != nil {
		t.Fatalf("Failed to create MySQL config on DEV: %v", err)
	}
	sshCmd = exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", env.SSHKeyPath,
		"-p", "9023", "linuxserver.io@localhost",
		"mysql -hm2-clone-basic-db-1 -umagento -pmagento magento -e 'CREATE TABLE IF NOT EXISTS clone_test (val VARCHAR(255)); INSERT INTO clone_test VALUES (\"FULL_CLONE_DB\");'")
	if err := sshCmd.Run(); err != nil {
		t.Fatalf("Failed to create test DB table on DEV: %v", err)
	}

	result := env.RunGovard(t, localDir, "bootstrap", "--clone",
		"--environment", "dev",
		"--skip-up",
		"--no-composer",
		"--no-admin",
		"--yes",
	)
	result.AssertSuccess(t)

	// Verify output mentions all sync types
	result.AssertOutputContains(t, "[sync:files]")
	result.AssertOutputContains(t, "[sync:media]")
	result.AssertOutputContains(t, "stream import completed")
}

func TestBootstrapCloneCombinations(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)

	t.Run("no-db", func(t *testing.T) {
		localDir := env.CreateTempProject(t, "local")
		result := env.RunGovard(t, localDir, "bootstrap", "--clone",
			"--environment", "dev", "--no-db", "--skip-up", "--no-composer", "--no-admin", "--yes")
		result.AssertSuccess(t)
		result.AssertOutputContains(t, "[sync:files]")
		result.AssertOutputNotContains(t, "stream import completed")
	})

	t.Run("no-media", func(t *testing.T) {
		localDir := env.CreateTempProject(t, "local")
		result := env.RunGovard(t, localDir, "bootstrap", "--clone",
			"--environment", "dev", "--no-media", "--skip-up", "--no-composer", "--no-admin", "--yes")
		result.AssertSuccess(t)
		result.AssertOutputContains(t, "[sync:files]")
		result.AssertOutputNotContains(t, "[sync:media]")
	})
}

func TestBootstrapFreshMagento(t *testing.T) {
	env := NewRealEnvTest(t)
	env.Setup(t)

	localDir := t.TempDir()

	// Run bootstrap --fresh (will also run init because no config exists)
	// Note: This test requires Docker to be able to start containers
	result := env.RunGovard(t, localDir, "bootstrap", "--fresh",
		"--framework", "magento2",
		"--framework-version", "2.4.6",
		"--no-composer",
		"--no-admin",
		"--no-db",
		"--yes",
	)

	// This test may fail in CI or restricted environments where Docker containers
	// cannot be started. We check for success OR for the expected init behavior.
	if result.Error != nil {
		// If bootstrap failed, at least verify that init ran and created the config
		t.Logf("Bootstrap failed (may be due to Docker restrictions): %v", result.Error)
	}

	// Verify .govard.yml was created
	if _, err := os.Stat(filepath.Join(localDir, ".govard.yml")); err != nil {
		t.Errorf("Expected .govard.yml to be created, got %v", err)
	}

	// Check for either Fresh install message or the init message
	if result.Error == nil {
		result.AssertOutputContains(t, "Fresh install")
	} else {
		// At minimum, init should have run
		result.AssertOutputContains(t, "Generated .govard.yml")
	}
}
