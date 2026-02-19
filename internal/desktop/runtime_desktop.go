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
