package tests

import (
	"testing"

	"govard/internal/desktop"
)

func TestDesktopPkgCalculateCPUPercentForTest(t *testing.T) {
	value := desktop.CalculateCPUPercentForTest(200, 100, 400, 200, 2, 0)
	if value <= 0 {
		t.Fatalf("expected positive CPU percent, got %f", value)
	}

	zero := desktop.CalculateCPUPercentForTest(100, 100, 400, 200, 2, 0)
	if zero != 0 {
		t.Fatalf("expected zero CPU percent when no delta, got %f", zero)
	}
}

func TestDesktopPkgBuildMetricsWarningsForTest(t *testing.T) {
	warnings := desktop.BuildMetricsWarningsForTest([]desktop.ProjectResourceMetric{
		{Project: "demo", OOMKilled: true},
		{Project: "shop", OOMKilled: true},
		{Project: "api", OOMKilled: false},
	}, nil)

	if len(warnings) == 0 {
		t.Fatal("expected OOM warning")
	}

	if warnings[0] != "OOM kill detected in: demo, shop" {
		t.Fatalf("unexpected OOM warning: %s", warnings[0])
	}
}

func TestDesktopPkgBytesToMBForTest(t *testing.T) {
	value := desktop.BytesToMBForTest(1048576)
	if value != 1 {
		t.Fatalf("expected 1 MB, got %f", value)
	}
}
