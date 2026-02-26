package tests

import (
	"govard/internal/engine"
	"strings"
	"testing"
)

func TestBuildMagentoCommandsStoreDomains(t *testing.T) {
	config := engine.Config{
		ProjectName: "testproject",
		Domain:      "main.test",
		StoreDomains: map[string]string{
			"brand-b.test": "brand_b",
		},
	}

	commands := engine.MagentoConfigCommandsForTest("testproject", config)

	foundUnsecure := false
	foundSecure := false

	for _, cmd := range commands {
		cmdStr := strings.Join(cmd.Args, " ")
		if strings.Contains(cmdStr, "config:set --scope=stores --scope-code=brand_b web/unsecure/base_url https://brand-b.test/") {
			foundUnsecure = true
			if !cmd.Optional {
				t.Error("Multistore config command should be optional")
			}
		}
		if strings.Contains(cmdStr, "config:set --scope=stores --scope-code=brand_b web/secure/base_url https://brand-b.test/") {
			foundSecure = true
			if !cmd.Optional {
				t.Error("Multistore secure config command should be optional")
			}
		}
	}

	if !foundUnsecure {
		t.Error("Did not find command to set unsecure base URL for brand_b")
	}
	if !foundSecure {
		t.Error("Did not find command to set secure base URL for brand_b")
	}
}

func TestBuildMagentoCommandsNoStoreDomains(t *testing.T) {
	config := engine.Config{
		ProjectName: "testproject",
		Domain:      "main.test",
		// StoreDomains is nil
	}

	commands := engine.MagentoConfigCommandsForTest("testproject", config)

	for _, cmd := range commands {
		if strings.Contains(strings.Join(cmd.Args, " "), "--scope=stores") {
			t.Errorf("Found unexpected store scope command: %s", cmd.Desc)
		}
	}
}
