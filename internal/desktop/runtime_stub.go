//go:build !desktop

package desktop

import (
	"context"
	"fmt"
)

var errDesktopNotAvailableFn = func() error {
	return fmt.Errorf("desktop runtime not available")
}

func openURL(ctx context.Context, url string) error {
	_ = ctx
	_ = url
	return errDesktopNotAvailableFn()
}

func emitEvent(ctx context.Context, name string, data interface{}) {
	_ = ctx
	_ = name
	_ = data
}

func chooseDirectory(ctx context.Context, title string, defaultDir string) (string, error) {
	_ = ctx
	_ = title
	_ = defaultDir
	return "", errDesktopNotAvailableFn()
}

func chooseSaveFile(
	ctx context.Context,
	title string,
	defaultDir string,
	defaultFilename string,
) (string, error) {
	_ = ctx
	_ = title
	_ = defaultDir
	_ = defaultFilename
	return "", errDesktopNotAvailableFn()
}

func showApplication(ctx context.Context) {
	_ = ctx
}

func hideApplication(ctx context.Context) {
	_ = ctx
}

func quitApplication(ctx context.Context) {
	_ = ctx
}
