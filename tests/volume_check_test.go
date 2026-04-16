package tests

import (
	"fmt"
	"govard/internal/engine"
	"os/exec"
	"testing"
)

func TestDockerVolumeHelpers(t *testing.T) {
	sourceVol := "govard-test-vol-src"
	targetVol := "govard-test-vol-dest"

	// Cleanup
	defer func() {
		if err := exec.Command("docker", "volume", "rm", "-f", sourceVol, targetVol).Run(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	}()

	// 1. Test IsVolumeEmpty on non-existent
	empty, err := engine.IsVolumeEmpty(sourceVol)
	if err != nil {
		t.Fatalf("IsVolumeEmpty for non-existent: %v", err)
	}
	if !empty {
		t.Error("IsVolumeEmpty for non-existent should be true")
	}

	// 2. Test IsVolumeEmpty on fresh volume
	_ = exec.Command("docker", "volume", "create", sourceVol).Run()
	empty, err = engine.IsVolumeEmpty(sourceVol)
	if err != nil {
		t.Fatalf("IsVolumeEmpty for fresh: %v", err)
	}
	if !empty {
		t.Error("IsVolumeEmpty for fresh volume should be true")
	}

	// 3. Test IsVolumeEmpty on non-empty
	_ = exec.Command("docker", "run", "--rm", "-v", fmt.Sprintf("%s:/data", sourceVol), "alpine", "touch", "/data/test.txt").Run()
	empty, err = engine.IsVolumeEmpty(sourceVol)
	if err != nil {
		t.Fatalf("IsVolumeEmpty for non-empty: %v", err)
	}
	if empty {
		t.Error("IsVolumeEmpty for non-empty volume should be false")
	}

	// 4. Test CloneVolume
	err = engine.CloneVolume(sourceVol, targetVol)
	if err != nil {
		t.Fatalf("CloneVolume failed: %v", err)
	}

	// Verify dest has data
	output, _ := exec.Command("docker", "run", "--rm", "-v", fmt.Sprintf("%s:/data", targetVol), "alpine", "ls", "/data").CombinedOutput()
	if string(output) != "test.txt\n" {
		t.Errorf("CloneVolume target data mismatch: %q", string(output))
	}
}
