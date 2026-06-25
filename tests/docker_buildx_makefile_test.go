package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMakefileUsesBuildxWrapperForImageTargets(t *testing.T) {
	content := readProjectFileForTest(t, "Makefile")
	if !strings.Contains(content, "./scripts/docker-buildx-bake.sh") {
		t.Fatalf("expected Makefile image targets to use buildx wrapper, got:\n%s", content)
	}
}

func TestDockerBuildxBakeWrapperPreparesMultiPlatformBuilder(t *testing.T) {
	content := readProjectFileForTest(t, filepath.Join("scripts", "docker-buildx-bake.sh"))
	for _, expected := range []string{
		"DOCKER_PLATFORMS",
		"govard-multiarch",
		"docker buildx create",
		"--driver docker-container",
		"docker buildx inspect \"${BUILDER_NAME}\" --bootstrap",
		"does not support requested platform",
		"--builder",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected buildx wrapper to contain %q, got:\n%s", expected, content)
		}
	}
}

func TestDockerBuildxBakeWrapperIsExecutable(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location")
	}
	projectRoot := filepath.Join(filepath.Dir(filename), "..")
	scriptPath := filepath.Join(projectRoot, "scripts", "docker-buildx-bake.sh")
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat %s: %v", scriptPath, err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("expected %s to be executable, mode is %s", scriptPath, info.Mode())
	}
}
