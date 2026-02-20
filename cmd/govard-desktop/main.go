//go:build desktop

package main

import (
	"log"
	"os"

	"govard/internal/desktop"

	"github.com/wailsapp/wails/v2"
)

func main() {
	app := desktop.NewApp()
	assets, err := desktop.ResolveAssets()
	if err != nil {
		log.Fatalf("Failed to locate frontend assets: %v", err)
	}

	launchOptions := desktop.ResolveLaunchOptions(os.Args[1:], os.Getenv(desktop.DesktopBackgroundEnvVar))
	err = wails.Run(desktop.BuildWailsOptions(app, assets, launchOptions))
	if err != nil {
		log.Fatalf("Failed to start Govard Desktop: %v", err)
	}
}
