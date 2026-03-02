//go:build desktop

package main

import (
	"log"
	"os"

	"govard/desktop/frontend"
	"govard/internal/desktop"

	"github.com/wailsapp/wails/v2"
)

func main() {
	app := desktop.NewApp()

	launchOptions := desktop.ResolveLaunchOptions(os.Args[1:], os.Getenv(desktop.DesktopBackgroundEnvVar))
	err := wails.Run(desktop.BuildWailsOptions(app, frontend.Assets, launchOptions))
	if err != nil {
		log.Fatalf("Failed to start Govard Desktop: %v", err)
	}
}
