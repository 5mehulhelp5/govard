package tests

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"

	"github.com/spf13/cobra"
)

func TestNeedsRemoteEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		opts     cmd.BootstrapRuntimeOptions
		expected bool
	}{
		{
			name:     "Fresh install should NOT need remote",
			opts:     cmd.BootstrapRuntimeOptions{Fresh: true},
			expected: false,
		},
		{
			name:     "Plan should need remote (for resolution)",
			opts:     cmd.BootstrapRuntimeOptions{Plan: true},
			expected: true,
		},
		{
			name:     "Clone should need remote",
			opts:     cmd.BootstrapRuntimeOptions{Clone: true},
			expected: true,
		},
		{
			name:     "DB Import (from remote) should need remote",
			opts:     cmd.BootstrapRuntimeOptions{DBImport: true, DBDump: ""},
			expected: true,
		},
		{
			name:     "DB Import (from local dump) should NOT need remote",
			opts:     cmd.BootstrapRuntimeOptions{DBImport: true, DBDump: "dump.sql"},
			expected: false,
		},
		{
			name:     "Media sync should need remote",
			opts:     cmd.BootstrapRuntimeOptions{MediaSync: cmd.MediaSyncOptimized},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cmd.NeedsRemoteEnvironmentForTest(tt.opts); got != tt.expected {
				t.Errorf("NeedsRemoteEnvironment() = %v; want %v", got, tt.expected)
			}
		})
	}
}

func TestRunBootstrapRemoteSyncsVendorWithTrailingSlash(t *testing.T) {
	tempDir := t.TempDir()
	chdirForTest(t, tempDir)

	config := engine.Config{
		ProjectName: "sample-project",
		Framework:   "magento2",
		Remotes: map[string]engine.RemoteConfig{
			"staging": {
				Host: "staging.example.com",
				Path: "/var/www/html",
			},
		},
	}

	opts := cmd.DefaultBootstrapRuntimeOptionsForTest()
	opts.Source = "staging"
	opts.ComposerInstall = true
	opts.AssumeYes = true

	// Mock composer.json existence
	if err := os.WriteFile(filepath.Join(tempDir, "composer.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	subcommandCalls := 0
	vendorSyncFound := false
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		subcommandCalls++
		// The call we are looking for is: govard sync --source staging --file --path vendor/ --yes
		if len(args) >= 6 && args[0] == "sync" && args[5] == "vendor/" {
			vendorSyncFound = true
		}

		// Simulate composer install failure to trigger vendor sync
		if len(args) >= 3 && args[0] == "tool" && args[1] == "composer" && args[2] == "install" {
			return errors.New("composer install failed")
		}
		return nil
	})()

	// Mock remote directory check
	defer cmd.SetBootstrapRemoteDirExistsForTest(func(remoteName string, remoteCfg engine.RemoteConfig, remotePath string) bool {
		return true
	})()

	err := cmd.RunBootstrapRemoteForTest(&cobra.Command{}, config, opts)
	if err != nil && !strings.Contains(err.Error(), "composer install failed") {
		t.Fatalf("unexpected error: %v", err)
	}

	if !vendorSyncFound {
		t.Fatal("expected vendor sync with trailing slash 'vendor/', but it was not found in subcommand calls")
	}
}

func TestRunBootstrapRemoteSkipsComposerInstallWhenVendorSatisfiesLock(t *testing.T) {
	tempDir := t.TempDir()
	chdirForTest(t, tempDir)

	config := engine.Config{
		ProjectName: "sample-project",
		Framework:   "magento2",
	}

	opts := cmd.DefaultBootstrapRuntimeOptionsForTest()
	opts.ComposerInstall = true
	opts.AssumeYes = true
	// Avoid triggering the unrelated "remote is required" gate at the top of
	// runBootstrapRemote (requiresRemote), which is orthogonal to the
	// composer-install-skip behavior under test here.
	opts.DBImport = false
	opts.MediaSync = ""

	if err := os.WriteFile(filepath.Join(tempDir, "composer.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "composer.lock"), []byte(`{
  "packages": [{"name": "psr/log", "version": "1.1.4"}],
  "packages-dev": []
}`), 0644); err != nil {
		t.Fatal(err)
	}
	installedDir := filepath.Join(tempDir, "vendor", "composer")
	if err := os.MkdirAll(installedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installedDir, "installed.json"), []byte(`{
  "packages": [{"name": "psr/log", "version": "1.1.4"}]
}`), 0644); err != nil {
		t.Fatal(err)
	}

	composerInstallCalls := 0
	defer cmd.SetGovardSubcommandRunnerForTest(func(subCmd *cobra.Command, args ...string) error {
		if len(args) >= 3 && args[0] == "tool" && args[1] == "composer" && args[2] == "install" {
			composerInstallCalls++
		}
		return nil
	})()

	if err := cmd.RunBootstrapRemoteForTest(&cobra.Command{}, config, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if composerInstallCalls != 0 {
		t.Fatalf("expected composer install to be skipped, but it was called %d time(s)", composerInstallCalls)
	}
}
