package tests

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"govard/internal/engine"
)

func TestDiscoverMergedCommandsProjectWinsOnConflict(t *testing.T) {
	projectRoot := t.TempDir()
	projectCommandsDir := filepath.Join(projectRoot, ".govard", "commands")
	if err := os.MkdirAll(projectCommandsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectCommandsDir, "hello"), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	globalCommandsDir := filepath.Join(t.TempDir(), "global-commands")
	if err := os.MkdirAll(globalCommandsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalCommandsDir, "hello"), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalCommandsDir, "deploy"), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalCommandsDir, "_ignored"), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv(engine.GlobalCommandsDirEnvVar, globalCommandsDir)
	commands, err := engine.DiscoverMergedCommands(projectRoot)
	if err != nil {
		t.Fatalf("discover merged commands: %v", err)
	}
	if len(commands) != 2 {
		t.Fatalf("expected 2 merged commands, got %d", len(commands))
	}

	names := []string{commands[0].Name, commands[1].Name}
	if !reflect.DeepEqual(names, []string{"deploy", "hello"}) {
		t.Fatalf("unexpected merged command names: %#v", names)
	}

	if commands[1].Path != filepath.Join(projectCommandsDir, "hello") {
		t.Fatalf("expected project hello command to win, got %s", commands[1].Path)
	}
	if commands[0].Path != filepath.Join(globalCommandsDir, "deploy") {
		t.Fatalf("expected global deploy command path, got %s", commands[0].Path)
	}
}

func TestDiscoverGlobalCommandsDefaultsToGovardHome(t *testing.T) {
	homeDir := filepath.Join(t.TempDir(), "govard-home")
	t.Setenv("GOVARD_HOME_DIR", homeDir)
	commandsDir := filepath.Join(homeDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "build"), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	commands, err := engine.DiscoverGlobalCommands()
	if err != nil {
		t.Fatalf("discover global commands: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("expected 1 global command, got %d", len(commands))
	}
	if commands[0].Name != "build" {
		t.Fatalf("expected global command build, got %s", commands[0].Name)
	}
}
