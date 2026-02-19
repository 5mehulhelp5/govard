package tests

import (
	"testing"

	"govard/internal/engine"
)

func TestBuildXdebugSessionPattern(t *testing.T) {
	tests := map[string]string{
		"PHPSTORM":          "PHPSTORM",
		"PHPSTORM, VSCODE":  "(PHPSTORM|VSCODE)",
		" PHPSTORM , ,IDEA": "(PHPSTORM|IDEA)",
		"":                  "PHPSTORM",
	}

	for raw, expected := range tests {
		actual := engine.BuildXdebugSessionPatternForTest(raw)
		if actual != expected {
			t.Fatalf("Expected %q for %q, got %q", expected, raw, actual)
		}
	}
}
