package desktop

import (
	"testing/fstest"

	"github.com/wailsapp/wails/v2/pkg/options"
)

// ResolveLaunchOptionsForTest exposes desktop launch option normalization for tests.
func ResolveLaunchOptionsForTest(args []string, envBackground string) LaunchOptions {
	return ResolveLaunchOptions(args, envBackground)
}

// BuildWailsOptionsForTest exposes Wails option wiring for tests.
func BuildWailsOptionsForTest(app *App, launch LaunchOptions) *options.App {
	return BuildWailsOptions(app, fstest.MapFS{}, launch)
}
