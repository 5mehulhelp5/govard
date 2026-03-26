package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPHPDockerfileInstallsCronieForMagentoCronInstall(t *testing.T) {
	content := readProjectFileForTest(t, filepath.Join("docker", "php", "Dockerfile"))
	if !strings.Contains(content, "cronie") {
		t.Fatalf("expected docker/php/Dockerfile to install cronie so non-root crontab works, got:\n%s", content)
	}
}

func TestPHPEntrypointStartsCrondForInstalledCrontabs(t *testing.T) {
	content := readProjectFileForTest(t, filepath.Join("docker", "php", "etc", "entrypoint.sh"))
	if !strings.Contains(content, "crond") {
		t.Fatalf("expected docker/php/etc/entrypoint.sh to start crond for installed crontabs, got:\n%s", content)
	}
}

func readProjectFileForTest(t *testing.T, relPath string) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location")
	}

	projectRoot := filepath.Join(filepath.Dir(filename), "..")
	fullPath := filepath.Join(projectRoot, relPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read %s: %v", fullPath, err)
	}
	return string(data)
}
