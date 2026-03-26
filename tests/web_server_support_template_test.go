package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestApacheSupportTemplateMatchesDockerTemplate(t *testing.T) {
	projectRoot := testProjectRoot(t)
	assertFilesEqual(
		t,
		filepath.Join(projectRoot, "docker", "apache", "etc", "httpd.conf"),
		filepath.Join(projectRoot, "internal", "blueprints", "files", "support", "apache", "httpd.conf"),
	)
}

func TestNginxSupportTemplatesMatchDockerTemplates(t *testing.T) {
	projectRoot := testProjectRoot(t)

	dockerDir := filepath.Join(projectRoot, "docker", "nginx", "etc", "templates")
	supportDir := filepath.Join(projectRoot, "internal", "blueprints", "files", "support", "nginx", "templates")

	entries, err := os.ReadDir(dockerDir)
	if err != nil {
		t.Fatalf("read docker nginx templates: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".conf" {
			continue
		}
		assertFilesEqual(
			t,
			filepath.Join(dockerDir, entry.Name()),
			filepath.Join(supportDir, entry.Name()),
		)
	}
}

func testProjectRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..")
}

func assertFilesEqual(t *testing.T, leftPath, rightPath string) {
	t.Helper()

	left, err := os.ReadFile(leftPath)
	if err != nil {
		t.Fatalf("read %s: %v", leftPath, err)
	}
	right, err := os.ReadFile(rightPath)
	if err != nil {
		t.Fatalf("read %s: %v", rightPath, err)
	}

	if string(left) != string(right) {
		t.Fatalf("expected files to match:\nleft:  %s\nright: %s", leftPath, rightPath)
	}
}
