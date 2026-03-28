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

func TestProjectCommandExists(t *testing.T) {
	root := cmd.RootCommandForTest()

	projectCommand, _, err := root.Find([]string{"project"})
	if err != nil {
		t.Fatalf("find project command: %v", err)
	}
	if projectCommand == nil || projectCommand.Use != "project" {
		t.Fatalf("unexpected project command: %#v", projectCommand)
	}

	openCommand, _, err := root.Find([]string{"project", "open"})
	if err != nil {
		t.Fatalf("find project open command: %v", err)
	}
	if openCommand == nil || openCommand.Use != "open <query>" {
		t.Fatalf("unexpected project open command: %#v", openCommand)
	}
}

func TestProjectOpenCommandPrintsMatchedProjectPath(t *testing.T) {
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
	root.SetArgs([]string{"project", "open", "bill"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute project open: %v\nstderr=%s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output != "/workspace/platform-billing" {
		t.Fatalf("unexpected matched path output: %q", output)
	}
}

func TestProjectOpenCommandReturnsErrorWhenNoMatch(t *testing.T) {
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
	root.SetArgs([]string{"project", "open", "unknown"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected project open to fail when no match exists")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "no project matches") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectOpenCommandReturnsErrorWhenRegistryMissing(t *testing.T) {
	t.Setenv(engine.ProjectRegistryPathEnvVar, filepath.Join(t.TempDir(), "projects.json"))

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := cmd.RootCommandForTest()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"project", "open", "demo"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected project open to fail when registry is empty")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "project registry is empty") {
		t.Fatalf("unexpected empty registry error: %v", err)
	}
}
