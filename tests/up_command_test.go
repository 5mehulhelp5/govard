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

func TestAutoTuneMagentoRuntimeAppliesVersionProfile(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			PHPVersion: "8.4",
			DBType:     "mariadb",
			DBVersion:  "11.4",
			Services: engine.Services{
				Cache:  "valkey",
				Search: "opensearch",
				Queue:  "none",
			},
			CacheVersion:  "8.0.0",
			SearchVersion: "2.19.0",
		},
	}

	notes := cmd.AutoTuneMagentoRuntime(&config, engine.ProjectMetadata{
		Framework: "magento2",
		Version:   "2.4.7-p3",
	})

	if len(notes) == 0 {
		t.Fatal("expected autotune notes")
	}
	if config.Stack.PHPVersion != "8.3" {
		t.Fatalf("expected autotuned PHP 8.3, got %s", config.Stack.PHPVersion)
	}
	if config.Stack.DBVersion != "11.4" {
		t.Fatalf("expected existing DB 11.4 preserved, got %s", config.Stack.DBVersion)
	}
	if config.Stack.Services.Cache != "redis" {
		t.Fatalf("expected autotuned cache redis, got %s", config.Stack.Services.Cache)
	}
	if config.Stack.Services.Queue != "rabbitmq" {
		t.Fatalf("expected autotuned queue rabbitmq, got %s", config.Stack.Services.Queue)
	}
	if config.Stack.SearchVersion != "2.12.0" {
		t.Fatalf("expected autotuned search 2.12.0, got %s", config.Stack.SearchVersion)
	}
	hasPreserveNote := false
	for _, note := range notes {
		if strings.Contains(note, "kept existing DB version") {
			hasPreserveNote = true
			break
		}
	}
	if !hasPreserveNote {
		t.Fatalf("expected preserve DB note, got: %v", notes)
	}
}

func TestAutoTuneMagentoRuntimeAllowsDBUpgrade(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			DBType:    "mariadb",
			DBVersion: "10.4",
		},
	}

	cmd.AutoTuneMagentoRuntime(&config, engine.ProjectMetadata{
		Framework: "magento2",
		Version:   "2.4.7-p3",
	})

	if config.Stack.DBVersion != "10.6" {
		t.Fatalf("expected DB upgrade to 10.6, got %s", config.Stack.DBVersion)
	}
}

func TestAutoTuneMagentoRuntimePreservesConfiguredApacheWebServer(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			Services: engine.Services{
				WebServer: "apache",
			},
		},
	}

	notes := cmd.AutoTuneMagentoRuntime(&config, engine.ProjectMetadata{
		Framework: "magento2",
		Version:   "2.4.8-p3",
	})

	if config.Stack.Services.WebServer != "apache" {
		t.Fatalf("expected configured apache web server to be preserved, got %s", config.Stack.Services.WebServer)
	}

	foundPreserveNote := false
	for _, note := range notes {
		if strings.Contains(note, "kept configured web server") {
			foundPreserveNote = true
			break
		}
	}
	if !foundPreserveNote {
		t.Fatalf("expected preserve web server note, got: %v", notes)
	}
}

func TestAutoTuneMagentoRuntimePreservesConfiguredHybridWebServer(t *testing.T) {
	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			Services: engine.Services{
				WebServer: "hybrid",
			},
		},
	}

	notes := cmd.AutoTuneMagentoRuntime(&config, engine.ProjectMetadata{
		Framework: "magento2",
		Version:   "2.4.8-p3",
	})

	if config.Stack.Services.WebServer != "hybrid" {
		t.Fatalf("expected configured hybrid web server to be preserved, got %s", config.Stack.Services.WebServer)
	}

	foundPreserveNote := false
	for _, note := range notes {
		if strings.Contains(note, "kept configured web server") {
			foundPreserveNote = true
			break
		}
	}
	if !foundPreserveNote {
		t.Fatalf("expected preserve web server note, got: %v", notes)
	}
}
