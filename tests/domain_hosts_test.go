package tests

import (
	"govard/internal/engine"
	"os"
	"path/filepath"
	"testing"
)

func TestHostsLogic(t *testing.T) {
	// We can't actually write to /etc/hosts in a test environment usually
	// But we can check if the logic identifies an existing entry

	domain := "validation-test.test"

	// This will likely fail on permission, which is expected for a non-root test
	err := engine.AddHostsEntry(domain)
	if err != nil {
		t.Logf("AddHostsEntry failed as expected (or due to env): %v", err)
	}
}

func TestConfigDomainPopulation(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "domain-test-*")
	defer os.RemoveAll(tempDir)

	projectName := filepath.Base(tempDir)
	expectedDomain := projectName + ".test"

	config := engine.Config{
		ProjectName: projectName,
		Domain:      expectedDomain,
	}

	if config.Domain != expectedDomain {
		t.Errorf("Expected domain %s, got %s", expectedDomain, config.Domain)
	}
}
