package tests

import (
	"govard/internal/cmd"
	"govard/internal/engine"
	"testing"
)

func TestDoctorFixMultiPassSkip(t *testing.T) {
	// 1. Setup a report with two failing checks
	report := engine.DoctorReport{
		Checks: []engine.DoctorCheck{
			{
				ID:     "check.a",
				Title:  "Check A",
				Status: engine.DoctorStatusWarn,
			},
			{
				ID:     "check.b",
				Title:  "Check B",
				Status: engine.DoctorStatusWarn,
			},
		},
	}

	// 2. Run first pass with an empty skipped map.
	// Since we can't easily mock the handlers to return Skip/Apply in this unit test without refactoring how handlers are registered,
	// we will at least verify that the map is being passed and respected by the wrapper.

	skipped := make(map[string]bool)

	// Pass 1: "check.a" is encountered, but no handler exists, so it should be "Unavailable"
	// and NOT added to skipped (only Skipped status is added).
	results := cmd.ApplyDoctorSafeFixesForTest(report, skipped)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// 3. Manually simulate "check.b" being skipped in the map
	skipped["check.b"] = true

	// Pass 2: "check.b" should now be bypassed completely
	results2 := cmd.ApplyDoctorSafeFixesForTest(report, skipped)

	foundB := false
	for _, res := range results2 {
		if res.CheckID == "check.b" {
			foundB = true
		}
	}

	if foundB {
		t.Errorf("expected check.b to be skipped by the map check, but it appeared in results")
	}
}
