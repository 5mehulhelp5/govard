package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestValidateDBCommandOptions(t *testing.T) {
	testCases := []struct {
		name      string
		sub       string
		options   cmd.DBCommandOptions
		expectErr bool
	}{
		{
			name:      "connect rejects file",
			sub:       "connect",
			options:   cmd.DBCommandOptions{Environment: "local", File: "dump.sql"},
			expectErr: true,
		},
		{
			name:      "dump rejects stream flag",
			sub:       "dump",
			options:   cmd.DBCommandOptions{Environment: "local", StreamDB: true},
			expectErr: true,
		},
		{
			name:      "import rejects stream with local source",
			sub:       "import",
			options:   cmd.DBCommandOptions{Environment: "local", StreamDB: true},
			expectErr: true,
		},
		{
			name:      "import allows remote stream",
			sub:       "import",
			options:   cmd.DBCommandOptions{Environment: "staging", StreamDB: true},
			expectErr: false,
		},
		{
			name:      "dump allows full mode",
			sub:       "dump",
			options:   cmd.DBCommandOptions{Environment: "local", Full: true},
			expectErr: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			err := cmd.ValidateDBCommandOptions(testCase.sub, testCase.options)
			if testCase.expectErr && err == nil {
				t.Fatal("expected error")
			}
			if !testCase.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
