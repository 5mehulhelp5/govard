package tests

import (
	"strings"
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"
)

func TestUpCommandQuickstartFlagExists(t *testing.T) {
	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"env", "up"})
	if err != nil {
		t.Fatalf("find env up: %v", err)
	}
	if command.Flags().Lookup("quickstart") == nil {
		t.Fatal("expected --quickstart flag on env up command")
	}
}

func TestUpCommandPullFlagExists(t *testing.T) {
	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"env", "up"})
	if err != nil {
		t.Fatalf("find env up: %v", err)
	}
	if command.Flags().Lookup("pull") == nil {
		t.Fatal("expected --pull flag on env up command")
	}
}

func TestUpCommandFallbackLocalBuildFlagExists(t *testing.T) {
	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"env", "up"})
	if err != nil {
		t.Fatalf("find env up: %v", err)
	}
	if command.Flags().Lookup("fallback-local-build") == nil {
		t.Fatal("expected --fallback-local-build flag on env up command")
	}
}

func TestUpCommandRemoveOrphansFlagExists(t *testing.T) {
	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"env", "up"})
	if err != nil {
		t.Fatalf("find env up: %v", err)
	}
	if command.Flags().Lookup("remove-orphans") == nil {
		t.Fatal("expected --remove-orphans flag on env up command")
	}
}

func TestUpCommandUsesRunE(t *testing.T) {
	root := cmd.RootCommandForTest()
	command, _, err := root.Find([]string{"env", "up"})
	if err != nil {
		t.Fatalf("find env up: %v", err)
	}
	if command.RunE == nil {
		t.Fatal("expected env up command to use RunE so failures return a non-zero exit code")
	}
}

func TestResolveUpProxyTargetDefaultWeb(t *testing.T) {
	target := cmd.ResolveUpProxyTarget(engine.Config{
		ProjectName: "demo",
		Stack: engine.Stack{
			Features: engine.Features{
				Varnish: false,
			},
		},
	})
	if target != "demo-web-1" {
		t.Fatalf("expected demo-web-1, got %s", target)
	}
}

func TestResolveUpProxyTargetWithVarnish(t *testing.T) {
	target := cmd.ResolveUpProxyTarget(engine.Config{
		ProjectName: "demo",
		Stack: engine.Stack{
			Features: engine.Features{
				Varnish: true,
			},
		},
	})
	if target != "demo-varnish-1" {
		t.Fatalf("expected demo-varnish-1, got %s", target)
	}
}

func TestApplyQuickstartProfileDisablesOptionalServices(t *testing.T) {
	config := engine.Config{
		Stack: engine.Stack{
			Features: engine.Features{
				Xdebug:        true,
				Varnish:       true,
				Redis:         true,
				Elasticsearch: true,
			},
			Services: engine.Services{
				Cache:  "redis",
				Search: "opensearch",
				Queue:  "rabbitmq",
			},
			CacheVersion:  "7.4",
			SearchVersion: "3.4.0",
			QueueVersion:  "3.13.7",
		},
	}

	cmd.ApplyQuickstartProfile(&config)

	if config.Stack.Features.Xdebug {
		t.Fatal("expected xdebug disabled by quickstart")
	}
	if config.Stack.Features.Varnish {
		t.Fatal("expected varnish disabled by quickstart")
	}
	if config.Stack.Services.Cache != "none" || config.Stack.CacheVersion != "" || config.Stack.Features.Redis {
		t.Fatalf("expected cache disabled, got service=%s version=%s redis=%t", config.Stack.Services.Cache, config.Stack.CacheVersion, config.Stack.Features.Redis)
	}
	if config.Stack.Services.Search != "none" || config.Stack.SearchVersion != "" || config.Stack.Features.Elasticsearch {
		t.Fatalf("expected search disabled, got service=%s version=%s elastic=%t", config.Stack.Services.Search, config.Stack.SearchVersion, config.Stack.Features.Elasticsearch)
	}
	if config.Stack.Services.Queue != "none" || config.Stack.QueueVersion != "" {
		t.Fatalf("expected queue disabled, got service=%s version=%s", config.Stack.Services.Queue, config.Stack.QueueVersion)
	}
}

func TestCheckMagentoRuntimeSyncReturnsWarnings(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			PHPVersion: "8.1", // Out of sync (expected 8.3)
			Services: engine.Services{
				Search: "elasticsearch", // Out of sync (expected opensearch)
			},
		},
	}

	warnings := cmd.CheckMagentoRuntimeSync(config, engine.ProjectMetadata{
		Framework: "magento2",
		Version:   "2.4.7-p3",
	})

	if len(warnings) == 0 {
		t.Fatal("expected warnings for out of sync profile")
	}

	warningMsg := warnings[0]
	if !strings.Contains(warningMsg, "PHP 8.1 (expected 8.3)") {
		t.Errorf("expected warning about PHP mismatch, got: %s", warningMsg)
	}
	if !strings.Contains(warningMsg, "Search elasticsearch (expected opensearch)") {
		t.Errorf("expected warning about Search mismatch, got: %s", warningMsg)
	}
}

func TestCheckMagentoRuntimeSyncReturnsNilWhenSynced(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			PHPVersion: "8.3",
			Services: engine.Services{
				Search: "opensearch",
			},
		},
	}

	warnings := cmd.CheckMagentoRuntimeSync(config, engine.ProjectMetadata{
		Framework: "magento2",
		Version:   "2.4.7-p3",
	})

	if len(warnings) > 0 {
		t.Fatalf("expected no warnings for synced profile, got: %v", warnings)
	}
}
