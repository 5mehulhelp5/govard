package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/desktop"
)

func TestDesktopRemoteOpenDBUsesBridgeStarterForTest(t *testing.T) {
	desktop.ResetStateForTest()
	projectRoot := t.TempDir()

	config := strings.TrimSpace(`
project_name: bridge-remote
framework: laravel
domain: bridge-remote.test
stack:
  php_version: "8.3"
  db_type: mariadb
  db_version: "10.6"
  services:
    web_server: nginx
remotes:
  staging:
    host: stage.example.com
    user: deploy
    path: /var/www/stage
    capabilities:
      files: true
      media: true
      db: true
      deploy: false
`) + "\n"
	if err := os.WriteFile(filepath.Join(projectRoot, ".govard.yml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	capturedRoot := ""
	capturedArgs := []string{}
	restore := desktop.SetStartGovardCommandForDesktopForTest(func(root string, args []string) error {
		capturedRoot = root
		capturedArgs = append([]string{}, args...)
		return nil
	})
	defer restore()

	app := desktop.NewApp()
	message, err := app.Remote.OpenRemoteDB(projectRoot, "staging")
	if err != nil {
		t.Fatalf("open remote db: %v", err)
	}
	if !strings.Contains(strings.ToLower(message), "opening remote database") {
		t.Fatalf("unexpected message: %q", message)
	}

	if capturedRoot != projectRoot {
		t.Fatalf("expected root %q, got %q", projectRoot, capturedRoot)
	}
	joined := strings.Join(capturedArgs, " ")
	for _, token := range []string{"open", "db", "-e", "staging", "--client"} {
		if !containsToken(capturedArgs, token) {
			t.Fatalf("expected token %q in args: %q", token, joined)
		}
	}
}

func TestDesktopRemoteOpenDBRejectsUnsupportedCapabilityForTest(t *testing.T) {
	desktop.ResetStateForTest()
	projectRoot := t.TempDir()

	config := strings.TrimSpace(`
project_name: bridge-remote
framework: laravel
domain: bridge-remote.test
stack:
  php_version: "8.3"
  db_type: mariadb
  db_version: "10.6"
  services:
    web_server: nginx
remotes:
  staging:
    host: stage.example.com
    user: deploy
    path: /var/www/stage
    capabilities:
      files: true
      media: true
      db: false
      deploy: false
`) + "\n"
	if err := os.WriteFile(filepath.Join(projectRoot, ".govard.yml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	app := desktop.NewApp()
	_, err := app.Remote.OpenRemoteDB(projectRoot, "staging")
	if err == nil {
		t.Fatalf("expected capability error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "does not allow db operations") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDesktopRemoteOpenShellUsesSSHURLForTest(t *testing.T) {
	desktop.ResetStateForTest()
	projectRoot := t.TempDir()

	config := strings.TrimSpace(`
project_name: bridge-remote
framework: laravel
domain: bridge-remote.test
stack:
  php_version: "8.3"
  db_type: mariadb
  db_version: "10.6"
  services:
    web_server: nginx
remotes:
  staging:
    host: stage.example.com
    user: deploy
    path: /var/www/stage
    capabilities:
      files: true
      media: true
      db: true
      deploy: false
`) + "\n"
	if err := os.WriteFile(filepath.Join(projectRoot, ".govard.yml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	app := desktop.NewApp()
	message, err := app.Remote.OpenRemoteShell(projectRoot, "staging")
	if err != nil {
		t.Fatalf("open remote shell: %v", err)
	}
	if !strings.Contains(message, "ssh://deploy@stage.example.com:22/var/www/stage") {
		t.Fatalf("unexpected shell open message: %q", message)
	}
}
