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

func TestShouldRunSymfonyPostClone(t *testing.T) {
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
			name:            "skips for symfony when composer install disabled",
			framework:       "symfony",
			composerInstall: false,
			want:            false,
		},
		{
			name:            "skips for non symfony",
			framework:       "laravel",
			composerInstall: true,
			want:            false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			got := cmd.ShouldRunSymfonyPostCloneForTest(testCase.framework, testCase.composerInstall)
			if got != testCase.want {
				t.Fatalf("expected %v, got %v", testCase.want, got)
			}
		})
	}
}

func TestShouldIgnoreSymfonyPostCloneError(t *testing.T) {
	cwd := t.TempDir()
	if err := os.MkdirAll(filepath.Join(cwd, "vendor"), 0o755); err != nil {
		t.Fatalf("mkdir vendor: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cwd, "vendor", "autoload.php"), []byte("<?php"), 0o644); err != nil {
		t.Fatalf("write autoload: %v", err)
	}

	if !cmd.ShouldIgnoreSymfonyPostCloneErrorForTest(errors.New(`composer install failed: exec: "composer": executable file not found in $PATH`), cwd) {
		t.Fatal("expected composer post-clone error to be ignored when vendor/autoload.php exists")
	}

	if cmd.ShouldIgnoreSymfonyPostCloneErrorForTest(errors.New("some other failure"), cwd) {
		t.Fatal("expected non-composer error to remain fatal")
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
