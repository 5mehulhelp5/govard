package tests

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"
)

func TestShouldRunFrameworkPostClone(t *testing.T) {
	testCases := []struct {
		name            string
		framework       string
		composerInstall bool
		want            bool
	}{
		{
			name:            "runs for symfony when composer install enabled",
			framework:       "symfony",
			composerInstall: true,
			want:            true,
		},
		{
			name:            "runs for laravel when composer install enabled",
			framework:       "laravel",
			composerInstall: true,
			want:            true,
		},
		{
			name:            "runs for wordpress when composer install enabled",
			framework:       "wordpress",
			composerInstall: true,
			want:            true,
		},
		{
			name:            "runs for openmage when composer install enabled",
			framework:       "openmage",
			composerInstall: true,
			want:            true,
		},
		{
			name:            "skips for symfony when composer install disabled",
			framework:       "symfony",
			composerInstall: false,
			want:            false,
		},
		{
			name:            "skips for magento2 (handled separately)",
			framework:       "magento2",
			composerInstall: true,
			want:            false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			got := cmd.ShouldRunFrameworkPostCloneForTest(testCase.framework, testCase.composerInstall)
			if got != testCase.want {
				t.Fatalf("expected %v, got %v", testCase.want, got)
			}
		})
	}
}

func TestShouldIgnoreFrameworkPostCloneError(t *testing.T) {
	cwd := t.TempDir()
	if err := os.MkdirAll(filepath.Join(cwd, "vendor"), 0o755); err != nil {
		t.Fatalf("mkdir vendor: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cwd, "vendor", "autoload.php"), []byte("<?php"), 0o644); err != nil {
		t.Fatalf("write autoload: %v", err)
	}

	if !cmd.ShouldIgnoreFrameworkPostCloneErrorForTest("symfony", errors.New(`composer install failed: exec: "composer": executable file not found in $PATH`), cwd) {
		t.Fatal("expected composer post-clone error to be ignored when vendor/autoload.php exists")
	}

	if cmd.ShouldIgnoreFrameworkPostCloneErrorForTest("symfony", errors.New("some other failure"), cwd) {
		t.Fatal("expected non-composer error to remain fatal")
	}

	// WordPress specific
	if cmd.ShouldIgnoreFrameworkPostCloneErrorForTest("wordpress", errors.New("any error"), cwd) {
		t.Fatal("expected WP error to be fatal if wp-config.php is missing")
	}
	_ = os.WriteFile(filepath.Join(cwd, "wp-config.php"), []byte("<?php"), 0o644)
	if !cmd.ShouldIgnoreFrameworkPostCloneErrorForTest("wordpress", errors.New("any error"), cwd) {
		t.Fatal("expected WP error to be ignored if wp-config.php exists")
	}
}

func TestShouldSkipBootstrapMediaSync(t *testing.T) {
	restore := cmd.SetBootstrapRemoteDirExistsForTest(func(remoteName string, remoteCfg engine.RemoteConfig, remotePath string) bool {
		return false
	})
	defer restore()

	config := engine.Config{
		Framework: "symfony",
		Remotes: map[string]engine.RemoteConfig{
			"dev": {
				Path: "/srv/www/app",
			},
		},
	}

	skip, reason := cmd.ShouldSkipBootstrapMediaSyncForTest(config, "dev", true, false, false)
	if !skip {
		t.Fatal("expected media sync to be skipped when remote media path is missing")
	}
	if !strings.Contains(reason, "does not exist") {
		t.Fatalf("unexpected skip reason: %s", reason)
	}

	restore = cmd.SetBootstrapRemoteDirExistsForTest(func(remoteName string, remoteCfg engine.RemoteConfig, remotePath string) bool {
		return true
	})
	defer restore()

	skip, reason = cmd.ShouldSkipBootstrapMediaSyncForTest(config, "dev", true, false, false)
	if skip {
		t.Fatalf("expected media sync to run when remote media path exists, reason=%s", reason)
	}
}
