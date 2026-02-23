package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"govard/internal/cmd"
	"govard/internal/engine"
)

type sleepStateFixture struct {
	Version  int                 `json:"version"`
	Projects []sleepProjectEntry `json:"projects"`
}

type sleepProjectEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func TestSleepWakeCommandsExist(t *testing.T) {
	root := cmd.RootCommandForTest()

	sleepCommand, _, err := root.Find([]string{"svc", "sleep"})
	if err != nil {
		t.Fatalf("find svc sleep command: %v", err)
	}
	if sleepCommand == nil || sleepCommand.Use != "sleep" {
		t.Fatalf("unexpected sleep command: %#v", sleepCommand)
	}

	wakeCommand, _, err := root.Find([]string{"svc", "wake"})
	if err != nil {
		t.Fatalf("find svc wake command: %v", err)
	}
	if wakeCommand == nil || wakeCommand.Use != "wake" {
		t.Fatalf("unexpected wake command: %#v", wakeCommand)
	}
}

func TestRunSleepForTestStopsRunningProjectsAndWritesState(t *testing.T) {
	t.Setenv(cmd.SleepStatePathEnvVar, filepath.Join(t.TempDir(), "sleep-state.json"))
	t.Setenv(engine.OperationsLogPathEnvVar, filepath.Join(t.TempDir(), "operations.log"))

	restoreDiscover := cmd.SetDiscoverRunningGovardProjectsForSleepForTest(func() ([]string, error) {
		return []string{"shop", "demo"}, nil
	})
	defer restoreDiscover()

	restoreRegistry := cmd.SetReadProjectRegistryEntriesForSleepForTest(func() ([]engine.ProjectRegistryEntry, error) {
		return []engine.ProjectRegistryEntry{
			{ProjectName: "demo", Path: "/workspace/demo"},
			{ProjectName: "shop", Path: "/workspace/shop"},
		}, nil
	})
	defer restoreRegistry()

	var calls []string
	restoreRunner := cmd.SetRunProjectGovardCommandForSleepForTest(func(projectPath string, args ...string) error {
		calls = append(calls, fmt.Sprintf("%s|%v", projectPath, args))
		return nil
	})
	defer restoreRunner()

	if err := cmd.RunSleepForTest(); err != nil {
		t.Fatalf("run sleep: %v", err)
	}

	expectedCalls := []string{
		"/workspace/demo|[stop]",
		"/workspace/shop|[stop]",
	}
	if !reflect.DeepEqual(calls, expectedCalls) {
		t.Fatalf("unexpected runner calls:\nexpected=%v\nactual=%v", expectedCalls, calls)
	}

	data, err := os.ReadFile(cmd.SleepStatePathForTest())
	if err != nil {
		t.Fatalf("read sleep state file: %v", err)
	}

	var state sleepStateFixture
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("parse sleep state: %v", err)
	}
	if len(state.Projects) != 2 {
		t.Fatalf("expected 2 sleep-state projects, got %d", len(state.Projects))
	}
	if state.Projects[0].Name != "demo" || state.Projects[0].Path != "/workspace/demo" {
		t.Fatalf("unexpected first sleep-state project: %#v", state.Projects[0])
	}
	if state.Projects[1].Name != "shop" || state.Projects[1].Path != "/workspace/shop" {
		t.Fatalf("unexpected second sleep-state project: %#v", state.Projects[1])
	}
}

func TestRunWakeForTestStartsProjectsAndClearsStateOnSuccess(t *testing.T) {
	t.Setenv(cmd.SleepStatePathEnvVar, filepath.Join(t.TempDir(), "sleep-state.json"))
	t.Setenv(engine.OperationsLogPathEnvVar, filepath.Join(t.TempDir(), "operations.log"))

	state := sleepStateFixture{
		Version: 1,
		Projects: []sleepProjectEntry{
			{Name: "demo", Path: "/workspace/demo"},
			{Name: "shop", Path: "/workspace/shop"},
		},
	}
	payload, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal state fixture: %v", err)
	}
	if err := os.WriteFile(cmd.SleepStatePathForTest(), payload, 0o600); err != nil {
		t.Fatalf("write state fixture: %v", err)
	}

	var calls []string
	restoreRunner := cmd.SetRunProjectGovardCommandForSleepForTest(func(projectPath string, args ...string) error {
		calls = append(calls, fmt.Sprintf("%s|%v", projectPath, args))
		return nil
	})
	defer restoreRunner()

	if err := cmd.RunWakeForTest(); err != nil {
		t.Fatalf("run wake: %v", err)
	}

	expectedCalls := []string{
		"/workspace/demo|[up]",
		"/workspace/shop|[up]",
	}
	if !reflect.DeepEqual(calls, expectedCalls) {
		t.Fatalf("unexpected runner calls:\nexpected=%v\nactual=%v", expectedCalls, calls)
	}

	if _, err := os.Stat(cmd.SleepStatePathForTest()); !os.IsNotExist(err) {
		t.Fatalf("expected sleep state to be removed after wake success, stat err=%v", err)
	}
}

func TestRunWakeForTestRetainsFailedProjectsInState(t *testing.T) {
	t.Setenv(cmd.SleepStatePathEnvVar, filepath.Join(t.TempDir(), "sleep-state.json"))
	t.Setenv(engine.OperationsLogPathEnvVar, filepath.Join(t.TempDir(), "operations.log"))

	state := sleepStateFixture{
		Version: 1,
		Projects: []sleepProjectEntry{
			{Name: "demo", Path: "/workspace/demo"},
			{Name: "shop", Path: "/workspace/shop"},
		},
	}
	payload, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal state fixture: %v", err)
	}
	if err := os.WriteFile(cmd.SleepStatePathForTest(), payload, 0o600); err != nil {
		t.Fatalf("write state fixture: %v", err)
	}

	restoreRunner := cmd.SetRunProjectGovardCommandForSleepForTest(func(projectPath string, args ...string) error {
		if projectPath == "/workspace/shop" {
			return fmt.Errorf("start failure")
		}
		return nil
	})
	defer restoreRunner()

	err = cmd.RunWakeForTest()
	if err == nil {
		t.Fatal("expected wake error when one project fails")
	}

	data, err := os.ReadFile(cmd.SleepStatePathForTest())
	if err != nil {
		t.Fatalf("read sleep state after partial wake failure: %v", err)
	}
	var remaining sleepStateFixture
	if err := json.Unmarshal(data, &remaining); err != nil {
		t.Fatalf("parse sleep state after partial wake failure: %v", err)
	}
	if len(remaining.Projects) != 1 {
		t.Fatalf("expected 1 remaining project, got %d", len(remaining.Projects))
	}
	if remaining.Projects[0].Name != "shop" {
		t.Fatalf("expected remaining project shop, got %#v", remaining.Projects[0])
	}
}
