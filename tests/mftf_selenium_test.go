package tests

import (
	"strings"
	"testing"

	"govard/internal/engine"
)

func TestRenderBlueprintMFTFDisabledNoSelenium(t *testing.T) {
	content := renderComposeWithConfig(t, engine.Config{
		ProjectName: "mftf-off",
		Framework:   "magento2",
		Domain:      "mftf-off.test",
		Stack: engine.Stack{
			Features: engine.Features{MFTF: false},
		},
	})

	if strings.Contains(content, "selenium") {
		t.Fatal("expected no selenium service when MFTF is disabled")
	}
}

func TestRenderBlueprintMFTFEnabledHasSelenium(t *testing.T) {
	content := renderComposeWithConfig(t, engine.Config{
		ProjectName: "mftf-on",
		Framework:   "magento2",
		Domain:      "mftf-on.test",
		Stack: engine.Stack{
			Features: engine.Features{MFTF: true},
		},
	})

	if !strings.Contains(content, "selenium") {
		t.Fatal("expected selenium service when MFTF is enabled")
	}
	if !strings.Contains(content, "selenium/standalone-chrome") {
		t.Fatal("expected selenium/standalone-chrome image")
	}
}

func TestMFTFConfigParseDefault(t *testing.T) {
	cfg := engine.Config{}
	if cfg.Stack.Features.MFTF {
		t.Fatal("expected MFTF to default to false")
	}
}

func TestMagento2FrameworkConfigIncludesSelenium(t *testing.T) {
	fwConfig, ok := engine.GetFrameworkConfig("magento2")
	if !ok {
		t.Fatal("expected magento2 framework config to exist")
	}

	found := false
	for _, inc := range fwConfig.Includes {
		if strings.Contains(inc, "selenium") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected magento2 includes to contain selenium.yml")
	}
}
