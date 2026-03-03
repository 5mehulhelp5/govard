package tests

import (
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
