//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestRunGovardWithShimsCapturesExternalCalls(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "shim-capture")
	shim := env.SetupRuntimeShims(t, map[string]int{
		"docker": 0,
		"ssh":    0,
		"rsync":  0,
	})

	result := env.RunGovardWithEnv(t, projectDir, shim.Env(), "sync", "--source", "dev", "--destination", "local", "--file")
	result.AssertSuccess(t)

	logs := shim.ReadLog(t)
	if !strings.Contains(logs, "rsync|") {
		t.Fatalf("expected rsync invocation in shim logs, got:\n%s", logs)
	}
}
