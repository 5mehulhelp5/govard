//go:build !desktop

package desktop

import (
	"context"
	"fmt"
)

func openURL(ctx context.Context, url string) error {
	_ = ctx
	_ = url
	return fmt.Errorf("desktop runtime not available")
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
	return "", fmt.Errorf("desktop runtime not available")
}

func showApplication(ctx context.Context) {
	_ = ctx
}

func hideApplication(ctx context.Context) {
	_ = ctx
}
