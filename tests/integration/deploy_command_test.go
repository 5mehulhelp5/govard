//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeployHooksExecute(t *testing.T) {
	env := NewTestEnvironment(t)
	projectDir := env.CreateProjectFromFixture(t, "magento2/options-local", "deploy-hooks")

	localOverride := `hooks:
  pre_deploy:
    - name: pre deploy marker
      run: "echo pre >> .govard-deploy-hooks.log"
  post_deploy:
    - name: post deploy marker
      run: "echo post >> .govard-deploy-hooks.log"
`
	overridePath := filepath.Join(projectDir, ".govard.local.yml")
	if err := os.WriteFile(overridePath, []byte(localOverride), 0o644); err != nil {
		t.Fatalf("failed to write .govard.local.yml: %v", err)
	}

	result := env.RunGovard(t, projectDir, "deploy")
	result.AssertSuccess(t)

	content, err := os.ReadFile(filepath.Join(projectDir, ".govard-deploy-hooks.log"))
	if err != nil {
		t.Fatalf("failed to read deploy hook log: %v", err)
	}
	out := strings.TrimSpace(string(content))
	if out != "pre\npost" {
		t.Fatalf("expected deploy hook order pre->post, got:\n%s", out)
	}
}
