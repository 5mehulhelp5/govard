package tests

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

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

func TestBuildUpReadinessChecksForPHPRuntime(t *testing.T) {
	checks := cmd.BuildUpReadinessChecksForTest(engine.Config{
		ProjectName: "demo",
		Framework:   "laravel",
		Stack: engine.Stack{
			Features: engine.Features{
				Xdebug: true,
			},
			Services: engine.Services{
				Cache: "none",
			},
		},
	})

	expected := []cmd.UpReadinessCheckForTest{
		{Service: "php", ContainerName: "demo-php-1"},
		{Service: "php-debug", ContainerName: "demo-php-debug-1"},
	}

	if !reflect.DeepEqual(checks, expected) {
		t.Fatalf("expected readiness checks %v, got %v", expected, checks)
	}
}

func TestBuildUpReadinessChecksForNonPHPRuntime(t *testing.T) {
	checks := cmd.BuildUpReadinessChecksForTest(engine.Config{
		ProjectName: "demo",
		Framework:   "nextjs",
	})
	if len(checks) != 0 {
		t.Fatalf("expected no readiness checks for non-PHP runtime, got %v", checks)
	}
}

func TestWaitForUpRuntimeReadinessRetriesUntilSuccess(t *testing.T) {
	attempts := 0

	restoreRunner := cmd.SetUpReadinessProbeRunnerForTest(func(containerName string, probeArgs []string) error {
		if containerName != "demo-php-1" {
			t.Fatalf("unexpected container %q", containerName)
		}
		attempts++
		if attempts < 3 {
			return errors.New("not ready")
		}
		return nil
	})
	defer restoreRunner()

	restoreInterval := cmd.SetUpReadinessProbeIntervalForTest(1 * time.Millisecond)
	defer restoreInterval()

	restoreSleep := cmd.SetUpReadinessSleepForTest(func(time.Duration) {})
	defer restoreSleep()

	err := cmd.WaitForUpRuntimeReadinessForTest(engine.Config{
		ProjectName: "demo",
		Framework:   "laravel",
		Stack: engine.Stack{
			Services: engine.Services{
				Cache: "none",
			},
		},
	}, 3*time.Millisecond)
	if err != nil {
		t.Fatalf("expected readiness wait to succeed, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 probe attempts, got %d", attempts)
	}
}

func TestWaitForUpRuntimeReadinessReturnsErrorAfterTimeout(t *testing.T) {
	restoreState := cmd.SetUpContainerStateRunnerForTest(func(containerName string) (string, error) {
		return "running|0|false|", nil
	})
	defer restoreState()

	restoreRunner := cmd.SetUpReadinessProbeRunnerForTest(func(containerName string, probeArgs []string) error {
		return errors.New("still booting")
	})
	defer restoreRunner()

	restoreInterval := cmd.SetUpReadinessProbeIntervalForTest(1 * time.Millisecond)
	defer restoreInterval()

	restoreSleep := cmd.SetUpReadinessSleepForTest(func(time.Duration) {})
	defer restoreSleep()

	err := cmd.WaitForUpRuntimeReadinessForTest(engine.Config{
		ProjectName: "demo",
		Framework:   "laravel",
		Stack: engine.Stack{
			Services: engine.Services{
				Cache: "none",
			},
		},
	}, 2*time.Millisecond)
	if err == nil {
		t.Fatal("expected readiness wait to fail")
	}
	if !strings.Contains(err.Error(), "php runtime did not become ready") {
		t.Fatalf("expected php readiness error, got %v", err)
	}
}

func TestWaitForUpRuntimeReadinessFailsFastWhenContainerExited(t *testing.T) {
	restoreState := cmd.SetUpContainerStateRunnerForTest(func(containerName string) (string, error) {
		if containerName != "demo-php-1" {
			t.Fatalf("unexpected container %q", containerName)
		}
		return "exited|1|false|permission denied", nil
	})
	defer restoreState()

	restoreRunner := cmd.SetUpReadinessProbeRunnerForTest(func(containerName string, probeArgs []string) error {
		t.Fatal("readiness probe should not run after exited state is detected")
		return nil
	})
	defer restoreRunner()

	err := cmd.WaitForUpRuntimeReadinessForTest(engine.Config{
		ProjectName: "demo",
		Framework:   "laravel",
		Stack: engine.Stack{
			Services: engine.Services{
				Cache: "none",
			},
		},
	}, 30*time.Second)
	if err == nil {
		t.Fatal("expected readiness wait to fail fast")
	}
	if !strings.Contains(err.Error(), "container demo-php-1 is exited") {
		t.Fatalf("expected exited container error, got %v", err)
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected container error detail, got %v", err)
	}
}

func TestRefreshCrossProjectRuntimeHostsRefreshesOtherPHPRuntimes(t *testing.T) {
	var renderedRoots []string
	var composeCalls []engine.ComposeOptions

	restore := cmd.SetUpCrossProjectRefreshDependenciesForTest(cmd.UpCrossProjectRefreshDependenciesForTest{
		GetRunningProjectNames: func(_ context.Context) ([]string, error) {
			return []string{"project-a", "project-b", "next-app"}, nil
		},
		ReadProjectRegistryEntries: func() ([]engine.ProjectRegistryEntry, error) {
			return []engine.ProjectRegistryEntry{
				{Path: "/workspace/project-a", ProjectName: "project-a", Domain: "project-a.test"},
				{Path: "/workspace/project-b", ProjectName: "project-b", Domain: "project-b.test"},
				{Path: "/workspace/next-app", ProjectName: "next-app", Domain: "next-app.test"},
			}, nil
		},
		LoadConfigFromDir: func(path string, _ bool) (engine.Config, []string, error) {
			switch path {
			case "/workspace/project-b":
				return engine.Config{
					ProjectName:    "project-b",
					Framework:      "laravel",
					Domain:         "project-b.test",
					LinkedProjects: []string{"project-a"},
					Stack: engine.Stack{
						Features: engine.Features{
							Xdebug: true,
						},
					},
				}, nil, nil
			case "/workspace/next-app":
				return engine.Config{
					ProjectName: "next-app",
					Framework:   "nextjs",
					Domain:      "next-app.test",
				}, nil, nil
			default:
				return engine.Config{}, nil, errors.New("unexpected project path")
			}
		},
		RenderBlueprint: func(root string, _ engine.Config) error {
			renderedRoots = append(renderedRoots, root)
			return nil
		},
		RunCompose: func(_ context.Context, opts engine.ComposeOptions) error {
			composeCalls = append(composeCalls, opts)
			return nil
		},
	})
	defer restore()

	err := cmd.RefreshCrossProjectRuntimeHostsForTest(context.Background(), "/workspace/project-a", engine.Config{
		ProjectName: "project-a",
		Framework:   "laravel",
		Domain:      "project-a.test",
	})
	if err != nil {
		t.Fatalf("refresh cross-project runtime hosts: %v", err)
	}

	if !reflect.DeepEqual(renderedRoots, []string{"/workspace/project-b"}) {
		t.Fatalf("rendered roots = %#v, want %#v", renderedRoots, []string{"/workspace/project-b"})
	}

	if len(composeCalls) != 1 {
		t.Fatalf("expected 1 compose refresh call, got %d", len(composeCalls))
	}

	got := composeCalls[0]
	if got.ProjectDir != "/workspace/project-b" {
		t.Fatalf("compose project dir = %q, want %q", got.ProjectDir, "/workspace/project-b")
	}
	if got.ProjectName != "project-b" {
		t.Fatalf("compose project name = %q, want %q", got.ProjectName, "project-b")
	}
	if !reflect.DeepEqual(got.Args, []string{"up", "-d", "--no-deps", "php", "php-debug"}) {
		t.Fatalf("compose args = %#v, want %#v", got.Args, []string{"up", "-d", "--no-deps", "php", "php-debug"})
	}
}

func TestApplyQuickstartProfileDisablesOptionalServices(t *testing.T) {
	config := engine.Config{
		Stack: engine.Stack{
			Features: engine.Features{
				Xdebug:  true,
				Varnish: true,
				Cache:   true,
				Search:  true,
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
	if config.Stack.Services.Cache != "none" || config.Stack.CacheVersion != "" || config.Stack.Features.Cache {
		t.Fatalf("expected cache disabled, got service=%s version=%s cache=%t", config.Stack.Services.Cache, config.Stack.CacheVersion, config.Stack.Features.Cache)
	}
	if config.Stack.Services.Search != "none" || config.Stack.SearchVersion != "" || config.Stack.Features.Search {
		t.Fatalf("expected search disabled, got service=%s version=%s search=%t", config.Stack.Services.Search, config.Stack.SearchVersion, config.Stack.Features.Search)
	}
	if config.Stack.Services.Queue != "none" || config.Stack.QueueVersion != "" {
		t.Fatalf("expected queue disabled, got service=%s version=%s", config.Stack.Services.Queue, config.Stack.QueueVersion)
	}
}

func TestCheckMagentoRuntimeSyncReturnsWarnings(t *testing.T) {
	// Setup a temporary project environment so LoadRawConfigFromDir finds an empty config
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(origWd)
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create a raw config that lacks explicit versions (so it's not 'intentional')
	rawContent := `project_name: test-sync
framework: magento2
`
	if err := os.WriteFile(".govard.yml", []byte(rawContent), 0644); err != nil {
		t.Fatal(err)
	}

	config := engine.Config{
		Framework: "magento2",
		Stack: engine.Stack{
			PHPVersion: "8.1", // This would be the normalized result if we were testing normalization
			Services: engine.Services{
				Search: "elasticsearch",
			},
		},
	}

	warnings := cmd.CheckMagentoRuntimeSync(config, engine.ProjectMetadata{
		Framework: "magento2",
		Version:   "2.4.7-p3",
	})

	if len(warnings) == 0 {
		t.Fatal("expected warnings for out of sync profile when versions are not explicitly set in raw config")
	}

	warningMsg := warnings[0]
	// Magento 2.4.7-p3 expects PHP 8.3.
	// Since raw config is empty, CheckMagentoRuntimeSync should warn about the mismatch.
	if !strings.Contains(warningMsg, "PHP") {
		t.Errorf("expected warning about PHP mismatch, got: %s", warningMsg)
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
