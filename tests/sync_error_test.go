package tests

import (
	"fmt"
	"os/exec"
	"testing"

	"govard/internal/cmd"
)

func TestHandleRsyncError(t *testing.T) {
	// Helper to get real exec.ExitError
	getExitErr := func(code int) error {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("exit %d", code))
		return cmd.Run()
	}

	tests := []struct {
		name          string
		err           error
		scope         string
		expectedCont  bool
		expectedError bool
	}{
		{"nil error", nil, "Media", true, false},
		{"Media exit 23", getExitErr(23), "Media", true, false},
		{"Media exit 24", getExitErr(24), "Media", true, false},
		{"Media fatal 1", getExitErr(1), "Media", false, true},
		{"Files exit 23", getExitErr(23), "Files", false, true},
		{"Files exit 24", getExitErr(24), "Files", false, true},
		{"Files fatal 1", getExitErr(1), "Files", false, true},
		{"Generic error", fmt.Errorf("random error"), "Media", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cont, err := cmd.HandleRsyncErrorForTest(tt.err, tt.scope)
			if cont != tt.expectedCont {
				t.Errorf("cont = %v, want %v", cont, tt.expectedCont)
			}
			if (err != nil) != tt.expectedError {
				t.Errorf("err = %v, want error: %v", err, tt.expectedError)
			}
		})
	}
}
