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

func TestPHPDockerfileInstallsCACertificatesForGovardRootCA(t *testing.T) {
	content := readProjectFileForTest(t, filepath.Join("docker", "php", "Dockerfile"))
	if !strings.Contains(content, "ca-certificates") {
		t.Fatalf("expected docker/php/Dockerfile to install ca-certificates so local Govard TLS trust can be refreshed, got:\n%s", content)
	}
}

func TestPHPEntrypointStartsCrondForInstalledCrontabs(t *testing.T) {
	content := readProjectFileForTest(t, filepath.Join("docker", "php", "etc", "entrypoint.sh"))
	if !strings.Contains(content, "crond") {
		t.Fatalf("expected docker/php/etc/entrypoint.sh to start crond for installed crontabs, got:\n%s", content)
	}
}

func TestPHPEntrypointRefreshesTrustStoreWhenGovardRootCAMounted(t *testing.T) {
	content := readProjectFileForTest(t, filepath.Join("docker", "php", "etc", "entrypoint.sh"))
	if !strings.Contains(content, "/usr/local/share/ca-certificates/govard.crt") {
		t.Fatalf("expected php entrypoint to look for mounted Govard Root CA, got:\n%s", content)
	}
	if !strings.Contains(content, "update-ca-certificates") {
		t.Fatalf("expected php entrypoint to refresh the trust store when Govard Root CA is mounted, got:\n%s", content)
	}
}

func TestPHPEntrypointDoesNotAbortOnUIDGIDRemapFailure(t *testing.T) {
	content := readProjectFileForTest(t, filepath.Join("docker", "php", "etc", "entrypoint.sh"))
	if !strings.Contains(content, "Warning: could not update www-data UID") {
		t.Fatalf("expected php entrypoint to warn instead of exiting on UID remap failure, got:\n%s", content)
	}
	if !strings.Contains(content, "Warning: could not update www-data GID") {
		t.Fatalf("expected php entrypoint to warn instead of exiting on GID remap failure, got:\n%s", content)
	}
}

func TestPHPEntrypointUpdatesGIDBeforeUID(t *testing.T) {
	content := readProjectFileForTest(t, filepath.Join("docker", "php", "etc", "entrypoint.sh"))

	gidIndex := strings.Index(content, "CURRENT_GID=$(id -g www-data)")
	uidIndex := strings.Index(content, "CURRENT_UID=$(id -u www-data)")
	if gidIndex == -1 || uidIndex == -1 {
		t.Fatalf("expected php entrypoint to contain both UID and GID remap blocks, got:\n%s", content)
	}
	if gidIndex > uidIndex {
		t.Fatalf("expected php entrypoint to update GID before UID, got:\n%s", content)
	}
}

func TestPHPEntrypointRelaunchesProcessAfterUIDRemap(t *testing.T) {
	content := readProjectFileForTest(t, filepath.Join("docker", "php", "etc", "entrypoint.sh"))
	if !strings.Contains(content, "UID_REMAP_CHANGED=1") {
		t.Fatalf("expected php entrypoint to track successful UID remaps, got:\n%s", content)
	}
	reenterIndex := strings.Index(content, "exec sudo -E -H -u www-data /usr/local/bin/entrypoint.sh \"$@\"")
	chownIndex := strings.Index(content, "if [ -n \"${CHOWN_DIR_LIST:-}\" ]; then")
	if reenterIndex == -1 {
		t.Fatalf("expected php entrypoint to re-enter as remapped www-data, got:\n%s", content)
	}
	if chownIndex == -1 {
		t.Fatalf("expected php entrypoint to contain CHOWN_DIR_LIST handling, got:\n%s", content)
	}
	if reenterIndex > chownIndex {
		t.Fatalf("expected php entrypoint to re-enter before recursive chown, got:\n%s", content)
	}
}

func TestPHPEntrypointStartsCrondBestEffort(t *testing.T) {
	content := readProjectFileForTest(t, filepath.Join("docker", "php", "etc", "entrypoint.sh"))
	if !strings.Contains(content, "sudo crond 2>/dev/null || true") {
		t.Fatalf("expected php entrypoint to start crond in best-effort mode, got:\n%s", content)
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
