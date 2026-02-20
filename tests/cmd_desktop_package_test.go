package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cmdpkg "govard/internal/cmd"
)

func TestCmdDesktopFindWailsCLIUsesGOBIN(t *testing.T) {
	tempDir := t.TempDir()
	binaryName := "wails"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	binaryPath := filepath.Join(tempDir, binaryName)
	if err := os.WriteFile(binaryPath, []byte(""), 0o755); err != nil {
		t.Fatalf("write fake wails binary: %v", err)
	}

	originalPath := os.Getenv("PATH")
	originalGOBIN := os.Getenv("GOBIN")
	originalGOPATH := os.Getenv("GOPATH")
	t.Cleanup(func() {
		_ = os.Setenv("PATH", originalPath)
		_ = os.Setenv("GOBIN", originalGOBIN)
		_ = os.Setenv("GOPATH", originalGOPATH)
	})

	_ = os.Setenv("PATH", "")
	_ = os.Setenv("GOPATH", "")
	_ = os.Setenv("GOBIN", tempDir)

	found, err := cmdpkg.FindWailsCLIForTest()
	if err != nil {
		t.Fatalf("find wails from GOBIN: %v", err)
	}
	if found != binaryPath {
		t.Fatalf("expected %s, got %s", binaryPath, found)
	}
}

func TestCmdDesktopFindWailsCLIUsesGOPATHBin(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir GOPATH bin: %v", err)
	}

	binaryName := "wails"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(binDir, binaryName)
	if err := os.WriteFile(binaryPath, []byte(""), 0o755); err != nil {
		t.Fatalf("write fake wails binary: %v", err)
	}

	originalPath := os.Getenv("PATH")
	originalGOBIN := os.Getenv("GOBIN")
	originalGOPATH := os.Getenv("GOPATH")
	t.Cleanup(func() {
		_ = os.Setenv("PATH", originalPath)
		_ = os.Setenv("GOBIN", originalGOBIN)
		_ = os.Setenv("GOPATH", originalGOPATH)
	})

	_ = os.Setenv("PATH", "")
	_ = os.Setenv("GOBIN", "")
	_ = os.Setenv("GOPATH", tempDir)

	found, err := cmdpkg.FindWailsCLIForTest()
	if err != nil {
		t.Fatalf("find wails from GOPATH/bin: %v", err)
	}
	if found != binaryPath {
		t.Fatalf("expected %s, got %s", binaryPath, found)
	}
}

func TestCmdDesktopBinaryArgsIncludesBackgroundFlag(t *testing.T) {
	args := cmdpkg.DesktopBinaryArgsForTest(true)
	if len(args) != 1 || args[0] != "--background" {
		t.Fatalf("expected [--background], got %v", args)
	}
}

func TestCmdDesktopBinaryArgsEmptyWhenBackgroundDisabled(t *testing.T) {
	args := cmdpkg.DesktopBinaryArgsForTest(false)
	if len(args) != 0 {
		t.Fatalf("expected empty args, got %v", args)
	}
}
