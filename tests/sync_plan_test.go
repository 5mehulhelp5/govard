package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"
)

func TestSyncPlanDirectoryDetection(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")

	if err := os.MkdirAll(filepath.Join(sourceDir, "vendor"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}

	config := engine.Config{
		ProjectName: "test-project",
		Framework:   "magento2",
	}

	endpoints := cmd.ResolveSyncEndpointsForTest(
		cmd.SyncEndpoint{
			Name:     "staging",
			IsLocal:  false,
			RootPath: "/var/www/html",
			RemoteCfg: engine.RemoteConfig{
				Host: "staging.example.com",
				Path: "/var/www/html",
			},
		},
		cmd.SyncEndpoint{Name: "local", IsLocal: true, RootPath: destDir},
	)

	// Test Case 1: --path vendor (no slash) but it exists as a directory
	opts := cmd.SyncExecutionOptionsForTest(true, false, false)
	opts.Path = "vendor"

	plan, err := cmd.BuildSyncExecutionPlanForTest(config, endpoints, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the rsync command uses trailing slashes
	foundRsync := false
	for _, cmdStr := range plan.Commands {
		if strings.Contains(cmdStr, "rsync") {
			foundRsync = true
			// Check for trailing slashes in source and destination paths
			// Since both are local in this test, buildRsyncForEndpoints handles them.
			// We should check if the command string includes "vendor/"
			if !strings.Contains(cmdStr, "vendor/") {
				t.Errorf("expected rsync command to contain 'vendor/', got: %s", cmdStr)
			}
		}
	}
	if !foundRsync {
		t.Fatal("rsync command not found in plan")
	}
}
