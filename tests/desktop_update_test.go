package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"govard/internal/desktop"
)

func TestDesktopCheckForUpdatesOutdated(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	desktop.ResetStateForTest()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v9.9.9"}`))
	}))
	defer server.Close()

	t.Setenv("GOVARD_UPDATE_CHECK_URL", server.URL)
	previousVersion := desktop.Version
	desktop.Version = "1.0.0"
	t.Cleanup(func() {
		desktop.Version = previousVersion
	})

	app := desktop.NewApp()
	result, err := app.CheckForUpdates()
	if err != nil {
		t.Fatalf("CheckForUpdates failed: %v", err)
	}
	if !result.Outdated {
		t.Fatalf("expected outdated=true, got false: %+v", result)
	}
	if result.CurrentVersion != "v1.0.0" {
		t.Fatalf("expected current version v1.0.0, got %q", result.CurrentVersion)
	}
	if result.LatestVersion != "v9.9.9" {
		t.Fatalf("expected latest version v9.9.9, got %q", result.LatestVersion)
	}
}

func TestDesktopCheckForUpdatesUpToDate(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	desktop.ResetStateForTest()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v1.16.1"}`))
	}))
	defer server.Close()

	t.Setenv("GOVARD_UPDATE_CHECK_URL", server.URL)
	previousVersion := desktop.Version
	desktop.Version = "1.16.1"
	t.Cleanup(func() {
		desktop.Version = previousVersion
	})

	app := desktop.NewApp()
	result, err := app.CheckForUpdates()
	if err != nil {
		t.Fatalf("CheckForUpdates failed: %v", err)
	}
	if result.Outdated {
		t.Fatalf("expected outdated=false, got true: %+v", result)
	}
	if !strings.Contains(result.Message, "up to date") {
		t.Fatalf("expected up-to-date message, got %q", result.Message)
	}
}

func TestDesktopInstallLatestUpdateRunsSelfUpdate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("desktop self-update bridge is not supported on windows")
	}

	t.Setenv("HOME", t.TempDir())
	desktop.ResetStateForTest()

	called := false
	restore := desktop.SetRunDesktopSelfUpdateForTest(func() (string, error) {
		called = true
		return "Successfully updated Govard to v1.16.1", nil
	})
	defer restore()

	app := desktop.NewApp()
	message, err := app.InstallLatestUpdate()
	if err != nil {
		t.Fatalf("InstallLatestUpdate failed: %v", err)
	}
	if !called {
		t.Fatal("expected self-update command to be called")
	}
	if !strings.Contains(message, "Successfully updated Govard") {
		t.Fatalf("unexpected install message: %q", message)
	}
}

func TestDesktopInstallLatestUpdateReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("desktop self-update bridge is not supported on windows")
	}

	t.Setenv("HOME", t.TempDir())
	desktop.ResetStateForTest()

	restore := desktop.SetRunDesktopSelfUpdateForTest(func() (string, error) {
		return "", fmt.Errorf("boom")
	})
	defer restore()

	app := desktop.NewApp()
	_, err := app.InstallLatestUpdate()
	if err == nil {
		t.Fatal("expected InstallLatestUpdate to return an error")
	}
}

func TestDesktopRestartDesktopAppStartsBinary(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	desktop.ResetStateForTest()

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "govard-desktop")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write fake desktop binary: %v", err)
	}

	restoreExec := desktop.SetDesktopExecutablePathForRestartForTest(func() (string, error) {
		return binaryPath, nil
	})
	defer restoreExec()

	restoreLookPath := desktop.SetDesktopBinaryLookPathForRestartForTest(func(file string) (string, error) {
		return "", fmt.Errorf("%s not found", file)
	})
	defer restoreLookPath()

	called := ""
	restoreRestart := desktop.SetRestartDesktopBinaryForTest(func(path string) error {
		called = path
		return nil
	})
	defer restoreRestart()

	app := desktop.NewApp()
	message, err := app.RestartDesktopApp()
	if err != nil {
		t.Fatalf("RestartDesktopApp failed: %v", err)
	}
	if !strings.Contains(message, "Restarting Govard Desktop") {
		t.Fatalf("unexpected restart message: %q", message)
	}
	if called != binaryPath {
		t.Fatalf("expected restart command to target %q, got %q", binaryPath, called)
	}
}

func TestDesktopRestartDesktopAppReturnsError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	desktop.ResetStateForTest()

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "govard-desktop")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write fake desktop binary: %v", err)
	}

	restoreExec := desktop.SetDesktopExecutablePathForRestartForTest(func() (string, error) {
		return binaryPath, nil
	})
	defer restoreExec()

	restoreRestart := desktop.SetRestartDesktopBinaryForTest(func(path string) error {
		return fmt.Errorf("boom")
	})
	defer restoreRestart()

	app := desktop.NewApp()
	_, err := app.RestartDesktopApp()
	if err == nil {
		t.Fatal("expected RestartDesktopApp to return an error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected error to include boom, got %v", err)
	}
}
