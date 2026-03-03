package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/desktop"
)

func TestDesktopGlobalServiceDefinitionsIncludeDNSMasqForTest(t *testing.T) {
	desktop.ResetStateForTest()

	service, ok := desktop.ResolveGlobalServiceForTest("dnsmasq")
	if !ok {
		t.Fatalf("expected dnsmasq global service definition")
	}
	if service.ComposeService != "dnsmasq" {
		t.Fatalf("expected compose service dnsmasq, got %q", service.ComposeService)
	}
	if service.ContainerName != "govard-proxy-dnsmasq" {
		t.Fatalf("expected dnsmasq container name, got %q", service.ContainerName)
	}
	if service.Openable {
		t.Fatalf("dnsmasq should not be openable in desktop UI")
	}
}

func TestDesktopGlobalServiceStatusDerivationForTest(t *testing.T) {
	status, health, running := desktop.DeriveGlobalContainerStatusForTest(
		"running",
		"Up 3 minutes (healthy)",
	)
	if status != "running" {
		t.Fatalf("expected running status, got %q", status)
	}
	if health != "healthy" {
		t.Fatalf("expected healthy health status, got %q", health)
	}
	if !running {
		t.Fatalf("expected running=true")
	}
}

func TestDesktopStartGlobalServiceRunsComposeForTargetServiceForTest(t *testing.T) {
	desktop.ResetStateForTest()
	restoreEnsure := desktop.SetEnsureGlobalServicesForDesktopForTest(func() error {
		return nil
	})
	defer restoreEnsure()

	var capturedArgs []string
	restoreRun := desktop.SetRunGlobalServicesComposeForDesktopForTest(
		func(args ...string) (string, error) {
			capturedArgs = append([]string{}, args...)
			return "", nil
		},
	)
	defer restoreRun()

	app := desktop.NewApp()
	message, err := app.StartGlobalService("dnsmasq")
	if err != nil {
		t.Fatalf("StartGlobalService failed: %v", err)
	}
	if !strings.Contains(strings.ToLower(message), "started") {
		t.Fatalf("unexpected message: %q", message)
	}

	expected := []string{"up", "-d", "dnsmasq"}
	if strings.Join(capturedArgs, " ") != strings.Join(expected, " ") {
		t.Fatalf("unexpected compose args: got %v want %v", capturedArgs, expected)
	}
}

func TestDesktopStopGlobalServiceRunsComposeForTargetServiceForTest(t *testing.T) {
	desktop.ResetStateForTest()
	home := t.TempDir()
	t.Setenv("HOME", home)

	composeDir := filepath.Join(home, ".govard", "proxy")
	if err := os.MkdirAll(composeDir, 0o755); err != nil {
		t.Fatalf("mkdir compose dir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(composeDir, "docker-compose.yml"),
		[]byte("services:\n  caddy: {}\n"),
		0o644,
	); err != nil {
		t.Fatalf("write compose file: %v", err)
	}

	var capturedArgs []string
	restoreRun := desktop.SetRunGlobalServicesComposeForDesktopForTest(
		func(args ...string) (string, error) {
			capturedArgs = append([]string{}, args...)
			return "", nil
		},
	)
	defer restoreRun()

	app := desktop.NewApp()
	message, err := app.StopGlobalService("mail")
	if err != nil {
		t.Fatalf("StopGlobalService failed: %v", err)
	}
	if !strings.Contains(strings.ToLower(message), "stopped") {
		t.Fatalf("unexpected message: %q", message)
	}

	expected := []string{"stop", "mail"}
	if strings.Join(capturedArgs, " ") != strings.Join(expected, " ") {
		t.Fatalf("unexpected compose args: got %v want %v", capturedArgs, expected)
	}
}

func TestDesktopGlobalServiceLogsRejectUnknownServiceForTest(t *testing.T) {
	desktop.ResetStateForTest()

	app := desktop.NewApp()
	_, err := app.GetGlobalServiceLogs("unknown", 100)
	if err == nil {
		t.Fatalf("expected unknown global service error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "unknown global service") {
		t.Fatalf("unexpected error: %v", err)
	}
}
