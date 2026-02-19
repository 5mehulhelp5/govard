package tests

import (
	"testing"

	"govard/internal/engine/bootstrap"
)

func TestBootstrapPkgLaravelFreshCommands(t *testing.T) {
	cases := []struct {
		version  string
		expected string
	}{
		{"11", "laravel/laravel"},
		{"10", "laravel/laravel:^10.0"},
		{"9", "laravel/laravel:^9.0"},
		{"", "laravel/laravel"},
	}

	for _, tc := range cases {
		opts := bootstrap.Options{Version: tc.version}
		laravel := bootstrap.NewLaravelBootstrap(opts)
		cmds := laravel.FreshCommands()

		if len(cmds) == 0 {
			t.Fatalf("expected commands for version %s, got none", tc.version)
		}

		if !containsSubstring(cmds[0], tc.expected) {
			t.Errorf("expected command to contain %q for version %s, got %q", tc.expected, tc.version, cmds[0])
		}
	}
}

func TestBootstrapPkgLaravelRun(t *testing.T) {
	opts := bootstrap.Options{Version: "11"}

	err := bootstrap.BootstrapLaravel(opts)
	if err != nil {
		t.Fatalf("BootstrapLaravel failed: %v", err)
	}
}

func TestBootstrapPkgDrupalFreshCommands(t *testing.T) {
	cases := []struct {
		version  string
		expected string
	}{
		{"11", "drupal/recommended-project"},
		{"10", "drupal/recommended-project:^10"},
		{"9", "drupal/recommended-project:^9"},
		{"", "drupal/recommended-project"},
	}

	for _, tc := range cases {
		opts := bootstrap.Options{Version: tc.version}
		drupal := bootstrap.NewDrupalBootstrap(opts)
		cmds := drupal.FreshCommands()

		if len(cmds) == 0 {
			t.Fatalf("expected commands for version %s, got none", tc.version)
		}

		if !containsSubstring(cmds[0], tc.expected) {
			t.Errorf("expected command to contain %q for version %s, got %q", tc.expected, tc.version, cmds[0])
		}
	}
}

func TestBootstrapPkgDrupalRun(t *testing.T) {
	opts := bootstrap.Options{Version: "10"}

	err := bootstrap.BootstrapDrupal(opts)
	if err != nil {
		t.Fatalf("BootstrapDrupal failed: %v", err)
	}
}

func TestBootstrapPkgWordPressRun(t *testing.T) {
	opts := bootstrap.Options{}

	err := bootstrap.BootstrapWordPress(opts)
	if err != nil {
		t.Fatalf("BootstrapWordPress failed: %v", err)
	}
}

func TestBootstrapPkgNextJSRun(t *testing.T) {
	opts := bootstrap.Options{}

	err := bootstrap.BootstrapNextJS(opts)
	if err != nil {
		t.Fatalf("BootstrapNextJS failed: %v", err)
	}
}

func TestBootstrapPkgShopwareRun(t *testing.T) {
	opts := bootstrap.Options{}

	err := bootstrap.BootstrapShopware(opts)
	if err != nil {
		t.Fatalf("BootstrapShopware failed: %v", err)
	}
}

func TestBootstrapPkgCakePHPRun(t *testing.T) {
	opts := bootstrap.Options{}

	err := bootstrap.BootstrapCakePHP(opts)
	if err != nil {
		t.Fatalf("BootstrapCakePHP failed: %v", err)
	}
}

func TestBootstrapPkgMagento1Run(t *testing.T) {
	opts := bootstrap.Options{}

	err := bootstrap.BootstrapMagento1(opts)
	if err != nil {
		t.Fatalf("BootstrapMagento1 failed: %v", err)
	}
}

func TestBootstrapPkgOpenMageRun(t *testing.T) {
	opts := bootstrap.Options{}

	err := bootstrap.BootstrapOpenMage(opts)
	if err != nil {
		t.Fatalf("BootstrapOpenMage failed: %v", err)
	}
}

func TestBootstrapPkgOpenMageFreshCommands(t *testing.T) {
	opts := bootstrap.Options{}
	openmage := bootstrap.NewOpenMageBootstrap(opts)
	cmds := openmage.FreshCommands()

	if len(cmds) == 0 {
		t.Fatal("expected commands for OpenMage, got none")
	}

	if !containsSubstring(cmds[0], "openmage/magento-lts") {
		t.Errorf("expected command to contain 'openmage/magento-lts', got %q", cmds[0])
	}
}

func TestBootstrapPkgNextJSFreshInstallSupport(t *testing.T) {
	opts := bootstrap.Options{}
	nextjs := bootstrap.NewNextJSBootstrap(opts)

	if !nextjs.SupportsFreshInstall() {
		t.Error("expected Next.js to support fresh install")
	}

	if !nextjs.SupportsClone() {
		t.Error("expected Next.js to support clone")
	}
}

func TestBootstrapDispatcherAllFrameworks(t *testing.T) {
	frameworks := []string{
		"magento2",
		"magento1",
		"openmage",
		"symfony",
		"laravel",
		"drupal",
		"wordpress",
		"nextjs",
		"shopware",
		"cakephp",
	}

	opts := bootstrap.DefaultOptions()

	for _, fw := range frameworks {
		t.Run(fw, func(t *testing.T) {
			err := bootstrap.Run(fw, opts)
			if err != nil {
				t.Fatalf("Run(%s) failed: %v", fw, err)
			}
		})
	}
}
