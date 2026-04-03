package tests

import (
	"govard/internal/cmd"
	"testing"
)

func TestNeedsRemoteEnvironmentImproved(t *testing.T) {
	tests := []struct {
		name     string
		opts     cmd.BootstrapRuntimeOptions
		expected bool
	}{
		{
			name:     "Fresh install should NOT need remote",
			opts:     cmd.BootstrapRuntimeOptions{Fresh: true},
			expected: false,
		},
		{
			name:     "Plan should need remote (for resolution)",
			opts:     cmd.BootstrapRuntimeOptions{Plan: true},
			expected: true,
		},
		{
			name:     "Clone should need remote",
			opts:     cmd.BootstrapRuntimeOptions{Clone: true},
			expected: true,
		},
		{
			name:     "DB Import (from remote) should need remote",
			opts:     cmd.BootstrapRuntimeOptions{DBImport: true, DBDump: ""},
			expected: true,
		},
		{
			name:     "DB Import (from local dump) should NOT need remote",
			opts:     cmd.BootstrapRuntimeOptions{DBImport: true, DBDump: "dump.sql"},
			expected: false,
		},
		{
			name:     "Media sync should need remote",
			opts:     cmd.BootstrapRuntimeOptions{MediaSync: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cmd.NeedsRemoteEnvironmentForTest(tt.opts); got != tt.expected {
				t.Errorf("NeedsRemoteEnvironment() = %v; want %v", got, tt.expected)
			}
		})
	}
}
