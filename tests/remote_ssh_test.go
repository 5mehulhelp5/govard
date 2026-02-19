package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/engine"
	"govard/internal/engine/remote"
)

func TestBuildSSHArgs(t *testing.T) {
	cfg := engine.RemoteConfig{Host: "example.com", User: "deploy", Port: 2222}
	args := remote.BuildSSHArgs("staging", cfg, true)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-p 2222") {
		t.Fatal("missing port")
	}
	if !strings.Contains(joined, "-A") {
		t.Fatal("missing agent forwarding")
	}
	if !strings.Contains(joined, "StrictHostKeyChecking=no") {
		t.Fatal("expected insecure host key checking by default")
	}
	if !strings.Contains(joined, "UserKnownHostsFile=/dev/null") {
		t.Fatal("expected dev-null known hosts by default")
	}
}

func TestBuildSSHArgsStrictHostKey(t *testing.T) {
	cfg := engine.RemoteConfig{
		Host: "example.com",
		User: "deploy",
		Port: 22,
		Auth: engine.RemoteAuth{
			StrictHostKey:  true,
			KnownHostsFile: "/tmp/govard-known-hosts",
		},
	}
	args := remote.BuildSSHArgs("staging", cfg, false)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "StrictHostKeyChecking=yes") {
		t.Fatal("expected strict host key checking")
	}
	if !strings.Contains(joined, "UserKnownHostsFile=/tmp/govard-known-hosts") {
		t.Fatal("expected custom known hosts file")
	}
	if strings.Contains(joined, "UserKnownHostsFile=/dev/null") {
		t.Fatal("did not expect insecure known hosts override in strict mode")
	}
}

func TestBuildRsyncCommandUsesSSHPolicy(t *testing.T) {
	cfg := engine.RemoteConfig{
		Host: "example.com",
		User: "deploy",
		Port: 22,
		Auth: engine.RemoteAuth{
			StrictHostKey: true,
		},
	}
	cmd := remote.BuildRsyncCommand("staging", "src/", "deploy@example.com:/srv/www/app/", cfg, false, true, nil, nil)
	joined := strings.Join(cmd.Args, " ")
	if !strings.Contains(joined, "StrictHostKeyChecking=yes") {
		t.Fatalf("expected strict host key in rsync ssh args, got: %s", joined)
	}
}

func TestBuildRsyncCommandIncludeExcludePatterns(t *testing.T) {
	cfg := engine.RemoteConfig{
		Host: "example.com",
		User: "deploy",
		Port: 22,
	}
	cmd := remote.BuildRsyncCommand(
		"staging",
		"src/",
		"deploy@example.com:/srv/www/app/",
		cfg,
		false,
		true,
		[]string{"app/*"},
		[]string{"vendor/"},
	)
	joined := strings.Join(cmd.Args, " ")
	if !strings.Contains(joined, "--include app/*") {
		t.Fatalf("expected include pattern in rsync args, got: %s", joined)
	}
	if !strings.Contains(joined, "--exclude vendor/") {
		t.Fatalf("expected exclude pattern in rsync args, got: %s", joined)
	}
	if !strings.Contains(joined, "--partial --append-verify") {
		t.Fatalf("expected resume rsync args, got: %s", joined)
	}
}

func TestBuildSSHArgsKeychainStoreFallback(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "auth.json")
	t.Setenv("GOVARD_AUTH_STORE_PATH", storePath)
	if err := remote.PersistSSHKeyPath("staging", "/opt/keys/staging"); err != nil {
		t.Fatalf("persist key path: %v", err)
	}

	cfg := engine.RemoteConfig{
		Host: "example.com",
		User: "deploy",
		Auth: engine.RemoteAuth{
			Method: remote.AuthMethodKeychain,
		},
	}
	args := remote.BuildSSHArgs("staging", cfg, false)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-i /opt/keys/staging") {
		t.Fatalf("expected key path resolved from auth store, got: %s", joined)
	}
}

func TestBuildSSHArgsKeyPathPriorityConfigOverEnvAndStore(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "auth.json")
	t.Setenv("GOVARD_AUTH_STORE_PATH", storePath)
	if err := remote.PersistSSHKeyPath("staging", "/store/key"); err != nil {
		t.Fatalf("persist key path: %v", err)
	}
	t.Setenv(remote.RemoteKeyPathEnvVar("staging"), "/env/key")

	cfg := engine.RemoteConfig{
		Host: "example.com",
		User: "deploy",
		Auth: engine.RemoteAuth{
			Method:  remote.AuthMethodKeychain,
			KeyPath: "/config/key",
		},
	}
	args := remote.BuildSSHArgs("staging", cfg, false)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-i /config/key") {
		t.Fatalf("expected config key path to win priority, got: %s", joined)
	}
}

func TestBuildSSHArgsKeyfileDefaultFallback(t *testing.T) {
	home := t.TempDir()
	keyPath := filepath.Join(home, ".ssh", "id_ed25519")
	if err := os.MkdirAll(filepath.Dir(keyPath), 0o755); err != nil {
		t.Fatalf("mkdir ssh dir: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("dummy-private-key"), 0o600); err != nil {
		t.Fatalf("write key file: %v", err)
	}
	t.Setenv("HOME", home)

	cfg := engine.RemoteConfig{
		Host: "example.com",
		User: "deploy",
		Auth: engine.RemoteAuth{
			Method: remote.AuthMethodKeyfile,
		},
	}
	args := remote.BuildSSHArgs("staging", cfg, false)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-i "+keyPath) {
		t.Fatalf("expected default keyfile fallback path, got: %s", joined)
	}
}
