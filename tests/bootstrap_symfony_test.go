package tests

import (
	"testing"

	"govard/internal/engine/bootstrap"
)

func TestBootstrapPkgSymfonyFreshCommands(t *testing.T) {
	cases := []struct {
		version  string
		expected string
	}{
		{"7.0", "symfony/skeleton"},
		{"6.4", "symfony/skeleton:^6.0"},
		{"5.4", "symfony/website-skeleton:^5.0"},
		{"", "symfony/skeleton"},
	}

	for _, tc := range cases {
		opts := bootstrap.Options{Version: tc.version}
		symfony := bootstrap.NewSymfonyBootstrap(opts)
		cmds := symfony.FreshCommands()

		if len(cmds) == 0 {
			t.Fatalf("expected commands for version %s, got none", tc.version)
		}

		if !containsSubstring(cmds[0], tc.expected) {
			t.Errorf("expected command to contain %q for version %s, got %q", tc.expected, tc.version, cmds[0])
		}
	}
}

func TestBootstrapPkgSymfonyRun(t *testing.T) {
	opts := bootstrap.Options{Version: "7.0"}

	err := bootstrap.BootstrapSymfony(opts)
	if err != nil {
		t.Fatalf("BootstrapSymfony failed: %v", err)
	}
}

func TestBootstrapDispatcherSymfony(t *testing.T) {
	opts := bootstrap.DefaultOptions()
	opts.Version = "7.0"

	err := bootstrap.Run("symfony", opts)
	if err != nil {
		t.Fatalf("Run(symfony) failed: %v", err)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
