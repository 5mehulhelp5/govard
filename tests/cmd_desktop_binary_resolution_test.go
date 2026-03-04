package tests

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	cmdpkg "govard/internal/cmd"
)

func TestFindDesktopBinaryForTestPrefersSiblingOfCurrentGovardBinary(t *testing.T) {
	tempDir := t.TempDir()
	cliDir := filepath.Join(tempDir, "cli")
	pathDir := filepath.Join(tempDir, "path")
	if err := os.MkdirAll(cliDir, 0o755); err != nil {
		t.Fatalf("mkdir cli dir: %v", err)
	}
	if err := os.MkdirAll(pathDir, 0o755); err != nil {
		t.Fatalf("mkdir path dir: %v", err)
	}

	cliBinary := filepath.Join(cliDir, "govard")
	siblingDesktop := filepath.Join(cliDir, "govard-desktop")
	pathDesktop := filepath.Join(pathDir, "govard-desktop")
	for _, candidate := range []string{cliBinary, siblingDesktop, pathDesktop} {
		if err := os.WriteFile(candidate, []byte(""), 0o755); err != nil {
			t.Fatalf("write %s: %v", candidate, err)
		}
	}

	restoreExecutable := cmdpkg.SetDesktopExecutablePathForTest(func() (string, error) {
		return cliBinary, nil
	})
	t.Cleanup(restoreExecutable)

	restoreLookPath := cmdpkg.SetDesktopLookPathForTest(func(_ string) (string, error) {
		return pathDesktop, nil
	})
	t.Cleanup(restoreLookPath)

	got, err := cmdpkg.FindDesktopBinaryForTest()
	if err != nil {
		t.Fatalf("FindDesktopBinaryForTest() error = %v", err)
	}
	if got != siblingDesktop {
		t.Fatalf("FindDesktopBinaryForTest() = %q, want sibling %q", got, siblingDesktop)
	}
}

func TestFindDesktopBinaryForTestFallsBackToLookPathWhenSiblingMissing(t *testing.T) {
	tempDir := t.TempDir()
	cliDir := filepath.Join(tempDir, "cli")
	pathDir := filepath.Join(tempDir, "path")
	if err := os.MkdirAll(cliDir, 0o755); err != nil {
		t.Fatalf("mkdir cli dir: %v", err)
	}
	if err := os.MkdirAll(pathDir, 0o755); err != nil {
		t.Fatalf("mkdir path dir: %v", err)
	}

	cliBinary := filepath.Join(cliDir, "govard")
	pathDesktop := filepath.Join(pathDir, "govard-desktop")
	if err := os.WriteFile(cliBinary, []byte(""), 0o755); err != nil {
		t.Fatalf("write cli binary: %v", err)
	}
	if err := os.WriteFile(pathDesktop, []byte(""), 0o755); err != nil {
		t.Fatalf("write path desktop binary: %v", err)
	}

	restoreExecutable := cmdpkg.SetDesktopExecutablePathForTest(func() (string, error) {
		return cliBinary, nil
	})
	t.Cleanup(restoreExecutable)

	restoreLookPath := cmdpkg.SetDesktopLookPathForTest(func(_ string) (string, error) {
		return pathDesktop, nil
	})
	t.Cleanup(restoreLookPath)

	got, err := cmdpkg.FindDesktopBinaryForTest()
	if err != nil {
		t.Fatalf("FindDesktopBinaryForTest() error = %v", err)
	}
	if got != pathDesktop {
		t.Fatalf("FindDesktopBinaryForTest() = %q, want PATH binary %q", got, pathDesktop)
	}
}

func TestFindDesktopBinaryForTestReturnsErrorWhenNoSiblingAndNoLookPathMatch(t *testing.T) {
	tempDir := t.TempDir()
	cliBinary := filepath.Join(tempDir, "govard")
	if err := os.WriteFile(cliBinary, []byte(""), 0o755); err != nil {
		t.Fatalf("write cli binary: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("read working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	outsideRepo := filepath.Join(tempDir, "outside-repo")
	if err := os.MkdirAll(outsideRepo, 0o755); err != nil {
		t.Fatalf("mkdir outside repo: %v", err)
	}
	if err := os.Chdir(outsideRepo); err != nil {
		t.Fatalf("chdir outside repo: %v", err)
	}

	restoreExecutable := cmdpkg.SetDesktopExecutablePathForTest(func() (string, error) {
		return cliBinary, nil
	})
	t.Cleanup(restoreExecutable)

	restoreLookPath := cmdpkg.SetDesktopLookPathForTest(func(_ string) (string, error) {
		return "", errors.New("not found")
	})
	t.Cleanup(restoreLookPath)

	if _, err := cmdpkg.FindDesktopBinaryForTest(); err == nil {
		t.Fatal("FindDesktopBinaryForTest() expected error, got nil")
	}
}
