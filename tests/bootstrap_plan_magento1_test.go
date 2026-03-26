package tests

import (
	"strings"
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"
)

func TestBuildBootstrapRemotePlanIncludesConfigAutoForMagento1(t *testing.T) {
	plan, err := cmd.BuildBootstrapRemotePlanForTest(engine.Config{
		Framework:   "magento1",
		ProjectName: "sample-project",
		Domain:      "sample.test",
	}, cmd.DefaultBootstrapRuntimeOptionsForTest())
	if err != nil {
		t.Fatalf("build bootstrap remote plan: %v", err)
	}

	found := false
	for _, command := range plan.Commands {
		if strings.Contains(command, "govard config auto") {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected Magento 1 bootstrap plan to include govard config auto")
	}
}
