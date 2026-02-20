package tests

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"govard/internal/cmd"
	"govard/internal/engine"
)

func TestProjectsCommandExists(t *testing.T) {
	root := cmd.RootCommandForTest()

	projectsCommand, _, err := root.Find([]string{"projects"})
	if err != nil {
		t.Fatalf("find projects command: %v", err)
	}
	if projectsCommand == nil || projectsCommand.Use != "projects" {
		t.Fatalf("unexpected projects command: %#v", projectsCommand)
	}

	openCommand, _, err := root.Find([]string{"projects", "open"})
	if err != nil {
		t.Fatalf("find projects open command: %v", err)
	}
	if openCommand == nil || openCommand.Use != "open <query>" {
		t.Fatalf("unexpected projects open command: %#v", openCommand)
	}
}

func TestProjectsOpenCommandPrintsMatchedProjectPath(t *testing.T) {
	t.Setenv(engine.ProjectRegistryPathEnvVar, filepath.Join(t.TempDir(), "projects.json"))
	entries := []engine.ProjectRegistryEntry{
		{
			Path:        "/workspace/demo-store",
			ProjectName: "demo-store",
			Domain:      "demo.test",
			LastSeenAt:  time.Date(2026, 2, 20, 9, 0, 0, 0, time.UTC),
		},
		{
			Path:        "/workspace/platform-billing",
			ProjectName: "billing-api",
			Domain:      "billing.test",
			LastSeenAt:  time.Date(2026, 2, 20, 10, 0, 0, 0, time.UTC),
		},
	}
	for _, entry := range entries {
		if err := engine.UpsertProjectRegistryEntry(entry); err != nil {
			t.Fatalf("upsert registry entry: %v", err)
		}
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := cmd.RootCommandForTest()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"projects", "open", "bill"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute projects open: %v\nstderr=%s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output != "/workspace/platform-billing" {
		t.Fatalf("unexpected matched path output: %q", output)
	}
}

func TestProjectsOpenCommandReturnsErrorWhenNoMatch(t *testing.T) {
	t.Setenv(engine.ProjectRegistryPathEnvVar, filepath.Join(t.TempDir(), "projects.json"))
	if err := engine.UpsertProjectRegistryEntry(engine.ProjectRegistryEntry{
		Path:        "/workspace/demo-store",
		ProjectName: "demo-store",
		LastSeenAt:  time.Date(2026, 2, 20, 9, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("upsert registry entry: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := cmd.RootCommandForTest()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"projects", "open", "unknown"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected projects open to fail when no match exists")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "no project matches") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectsOpenCommandReturnsErrorWhenRegistryMissing(t *testing.T) {
	t.Setenv(engine.ProjectRegistryPathEnvVar, filepath.Join(t.TempDir(), "projects.json"))

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := cmd.RootCommandForTest()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"projects", "open", "demo"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected projects open to fail when registry is empty")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "project registry is empty") {
		t.Fatalf("unexpected empty registry error: %v", err)
	}
}
