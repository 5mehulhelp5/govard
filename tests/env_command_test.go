package tests

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"

	"github.com/spf13/cobra"
)

func TestEnvRestartReappliesProjectDomainsAfterComposeUp(t *testing.T) {
	tempDir := t.TempDir()
	chdirForTest(t, tempDir)
	writeRuntimeConfig(t, tempDir, `project_name: wordpress
domain: wordpress.test
framework: wordpress
`)

	var composeArgs [][]string
	var registeredDomains []string
	var registeredTarget string
	var mappedHosts []string

	restore := cmd.SetEnvDependenciesForTest(cmd.EnvDependenciesForTest{
		RunCompose: func(_ context.Context, opts engine.ComposeOptions) error {
			captured := append([]string{}, opts.Args...)
			composeArgs = append(composeArgs, captured)
			return nil
		},
		RegisterDomains: func(domains []string, target string) error {
			registeredDomains = append([]string{}, domains...)
			registeredTarget = target
			return nil
		},
		UnregisterDomain: func(string) error { return nil },
		AddHostsEntry: func(domain string) error {
			mappedHosts = append(mappedHosts, domain)
			return nil
		},
		RemoveHostsEntry:          func(string) error { return nil },
		IsDomainResolvableLocally: func(string) bool { return false },
		RunHooks:                  func(engine.Config, string, io.Writer, io.Writer) error { return nil },
		RefreshPMAActiveProjects:  func() error { return nil },
	})
	defer restore()

	command := &cobra.Command{}
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	if err := cmd.ProxyEnvToComposeForTest(command, []string{"restart"}); err != nil {
		t.Fatalf("execute env restart: %v", err)
	}

	wantComposeArgs := [][]string{
		{"stop"},
		{"up", "-d"},
	}
	if !reflect.DeepEqual(composeArgs, wantComposeArgs) {
		t.Fatalf("compose args = %#v, want %#v", composeArgs, wantComposeArgs)
	}

	if !reflect.DeepEqual(registeredDomains, []string{"wordpress.test"}) {
		t.Fatalf("registered domains = %#v, want %#v", registeredDomains, []string{"wordpress.test"})
	}
	if registeredTarget != "wordpress-web-1" {
		t.Fatalf("registered target = %q, want %q", registeredTarget, "wordpress-web-1")
	}
	if !reflect.DeepEqual(mappedHosts, []string{"wordpress.test"}) {
		t.Fatalf("mapped hosts = %#v, want %#v", mappedHosts, []string{"wordpress.test"})
	}
}

func TestInitRejectsDuplicateProjectIdentity(t *testing.T) {
	registryPath := filepath.Join(t.TempDir(), "projects.json")
	t.Setenv(engine.ProjectRegistryPathEnvVar, registryPath)

	if err := engine.UpsertProjectRegistryEntry(engine.ProjectRegistryEntry{
		Path:        "/workspace/existing-wordpress",
		ProjectName: "wordpress",
		Domain:      "wordpress.test",
		Framework:   "wordpress",
	}); err != nil {
		t.Fatalf("seed registry: %v", err)
	}

	parentDir := t.TempDir()
	projectDir := filepath.Join(parentDir, "wordpress")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	chdirForTest(t, projectDir)

	root := cmd.RootCommandForTest()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"init", "--framework", "wordpress", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected duplicate project identity error")
	}
	if got := err.Error(); got == "" || !strings.Contains(strings.ToLower(got), "project_name wordpress is already used") {
		t.Fatalf("unexpected duplicate identity error: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(projectDir, ".govard.yml")); !os.IsNotExist(statErr) {
		t.Fatalf("expected no .govard.yml to be written, stat err = %v", statErr)
	}
}
