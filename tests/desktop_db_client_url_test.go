package tests

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"govard/internal/desktop"
)

func TestDesktopPkgParseDockerPublishedPortForTest(t *testing.T) {
	host, port, ok := desktop.ParseDockerPublishedPortForTest("0.0.0.0:3306\n[::]:3306\n")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if host != "127.0.0.1" {
		t.Fatalf("expected localhost host, got %q", host)
	}
	if port != "3306" {
		t.Fatalf("expected port 3306, got %q", port)
	}
}

func TestDesktopPkgParseDockerPublishedPortForTest_CompositeHostOutput(t *testing.T) {
	host, port, ok := desktop.ParseDockerPublishedPortForTest("192.168.144.3:172.28.0.6:3306\n")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if host != "192.168.144.3" {
		t.Fatalf("expected host 192.168.144.3, got %q", host)
	}
	if port != "3306" {
		t.Fatalf("expected port 3306, got %q", port)
	}
}

func TestDesktopPkgBuildDesktopDBClientURLForTest_Defaults(t *testing.T) {
	got := desktop.BuildDesktopDBClientURLForTest("mysql", "user", "pass", "", "", "sample")
	want := "mysql://user:pass@127.0.0.1:3306/sample"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestDesktopPkgBuildDesktopDBClientURLForTest_EncodesCredentials(t *testing.T) {
	got := desktop.BuildDesktopDBClientURLForTest(
		"mysql",
		"user",
		"p@ss/word",
		"127.0.0.1",
		"3306",
		"sample",
	)
	if strings.Contains(got, "p@ss/word") {
		t.Fatalf("expected escaped password in URL, got %q", got)
	}
	if !strings.Contains(got, "p%40ss%2Fword") {
		t.Fatalf("expected encoded password in URL, got %q", got)
	}
}

func TestDesktopPkgParseContainerIPAddressesForTest(t *testing.T) {
	got := desktop.ParseContainerIPAddressesForTest("192.168.144.3\n172.28.0.6\n192.168.144.3\n\n")
	want := []string{"192.168.144.3", "172.28.0.6"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestDesktopPkgBuildPMAOpenURLForTest(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	desktop.ResetStateForTest()

	raw := desktop.BuildPMAOpenURLForTest("sample-project", "magento")
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse pma URL: %v", err)
	}
	if parsed.Scheme != "https" {
		t.Fatalf("expected https scheme, got %q", parsed.Scheme)
	}
	if parsed.Host != "pma.govard.test" {
		t.Fatalf("expected pma.govard.test host, got %q", parsed.Host)
	}
	if got := parsed.Query().Get("project"); got != "sample-project" {
		t.Fatalf("expected project query sample-project, got %q", got)
	}
	if got := parsed.Query().Get("db"); got != "magento" {
		t.Fatalf("expected db query magento, got %q", got)
	}
}

func TestDesktopPkgBuildPMAOpenURLForTest_ProjectOnly(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	desktop.ResetStateForTest()

	raw := desktop.BuildPMAOpenURLForTest("sample-project", "")
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse pma URL: %v", err)
	}
	if got := parsed.Query().Get("project"); got != "sample-project" {
		t.Fatalf("expected project query sample-project, got %q", got)
	}
	if got := parsed.Query().Get("db"); got != "" {
		t.Fatalf("expected empty db query, got %q", got)
	}
}
