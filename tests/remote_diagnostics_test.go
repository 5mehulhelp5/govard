package tests

import (
	"errors"
	"testing"

	"govard/internal/engine/remote"
)

func TestClassifyFailureHostKey(t *testing.T) {
	details := remote.ClassifyFailure(errors.New("exit status 255"), "Host key verification failed.")
	if details.Category != remote.FailureCategoryHostKey {
		t.Fatalf("expected host_key, got %s", details.Category)
	}
}

func TestClassifyFailureAuth(t *testing.T) {
	details := remote.ClassifyFailure(errors.New("exit status 255"), "Permission denied (publickey).")
	if details.Category != remote.FailureCategoryAuth {
		t.Fatalf("expected auth, got %s", details.Category)
	}
}

func TestClassifyFailureNetwork(t *testing.T) {
	details := remote.ClassifyFailure(errors.New("exit status 255"), "ssh: connect to host example.com port 22: Connection timed out")
	if details.Category != remote.FailureCategoryNetwork {
		t.Fatalf("expected network, got %s", details.Category)
	}
}

func TestClassifyFailureDependency(t *testing.T) {
	details := remote.ClassifyFailure(errors.New("exit status 127"), "bash: rsync: command not found")
	if details.Category != remote.FailureCategoryDependency {
		t.Fatalf("expected dependency, got %s", details.Category)
	}
}

func TestClassifyFailurePermission(t *testing.T) {
	details := remote.ClassifyFailure(errors.New("exit status 1"), "Permission denied")
	if details.Category != remote.FailureCategoryPermission {
		t.Fatalf("expected permission, got %s", details.Category)
	}
}
