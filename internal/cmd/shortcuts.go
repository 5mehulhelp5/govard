package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "Open Mailpit web interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		url := "https://mail.govard.test"
		pterm.Info.Printf("Opening Mailpit: %s\n", url)
		return openURL(url)
	},
}

var pmaCmd = &cobra.Command{
	Use:   "pma",
	Short: "Open PHPMyAdmin interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		url := "https://pma.govard.test"
		pterm.Info.Printf("Opening PHPMyAdmin: %s\n", url)
		return openURL(url)
	},
}

func openURL(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Run()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()
	case "darwin":
		return exec.Command("open", url).Run()
	default:
		return fmt.Errorf("unsupported platform %q", runtime.GOOS)
	}
}
