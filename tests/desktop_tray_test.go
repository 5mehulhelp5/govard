package tests

import (
	"context"
	"testing"

	"govard/internal/desktop"
)

func TestDesktopLaunchOptionsEnableBackgroundFromFlag(t *testing.T) {
	options := desktop.ResolveLaunchOptionsForTest([]string{"--background"}, "")
	if !options.Background {
		t.Fatalf("expected background mode enabled from --background flag")
	}
}

func TestDesktopLaunchOptionsEnableBackgroundFromEnv(t *testing.T) {
	options := desktop.ResolveLaunchOptionsForTest(nil, "yes")
	if !options.Background {
		t.Fatalf("expected background mode enabled from env")
	}
}

func TestDesktopLaunchOptionsDisableBackgroundByDefault(t *testing.T) {
	options := desktop.ResolveLaunchOptionsForTest(nil, "")
	if options.Background {
		t.Fatalf("expected background mode disabled by default")
	}
}

func TestDesktopWailsOptionsBackgroundModeEnablesHideAndSingleInstance(t *testing.T) {
	desktop.ResetStateForTest()
	app := desktop.NewApp()
	wailsOptions := desktop.BuildWailsOptionsForTest(app, desktop.LaunchOptions{Background: true})

	if !wailsOptions.StartHidden {
		t.Fatalf("expected StartHidden enabled in background mode")
	}
	if !wailsOptions.HideWindowOnClose {
		t.Fatalf("expected HideWindowOnClose enabled in background mode")
	}
	if wailsOptions.SingleInstanceLock == nil {
		t.Fatalf("expected single instance lock in background mode")
	}
	if wailsOptions.OnBeforeClose == nil {
		t.Fatalf("expected OnBeforeClose handler in background mode")
	}
	if !wailsOptions.OnBeforeClose(context.Background()) {
		t.Fatalf("expected OnBeforeClose to prevent close in background mode")
	}
}

func TestDesktopWailsOptionsStandardModeKeepsDefaultCloseBehavior(t *testing.T) {
	desktop.ResetStateForTest()
	app := desktop.NewApp()
	wailsOptions := desktop.BuildWailsOptionsForTest(app, desktop.LaunchOptions{})

	if wailsOptions.StartHidden {
		t.Fatalf("expected StartHidden disabled by default")
	}
	if wailsOptions.HideWindowOnClose {
		t.Fatalf("expected HideWindowOnClose disabled by default")
	}
	if wailsOptions.SingleInstanceLock != nil {
		t.Fatalf("expected no single instance lock by default")
	}
	if wailsOptions.OnBeforeClose != nil {
		t.Fatalf("expected OnBeforeClose unset by default")
	}
}
