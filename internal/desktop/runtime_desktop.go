//go:build desktop

package desktop

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func openURL(ctx context.Context, url string) error {
	runtime.BrowserOpenURL(ctx, url)
	return nil
}

func emitEvent(ctx context.Context, name string, data interface{}) {
	runtime.EventsEmit(ctx, name, data)
}

func chooseDirectory(ctx context.Context, title string, defaultDir string) (string, error) {
	return runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: defaultDir,
	})
}

func showApplication(ctx context.Context) {
	runtime.Show(ctx)
	runtime.WindowShow(ctx)
	runtime.WindowUnminimise(ctx)
}

func hideApplication(ctx context.Context) {
	runtime.WindowHide(ctx)
	runtime.Hide(ctx)
}
