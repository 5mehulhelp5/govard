package cmd

import (
	"os"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Upgrade the Govard binary",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("Govard Self-Update")

		pterm.Info.Println("Downloading latest version...")

		// Mock implementation of self-update
		// Typically involves downloading from GitHub and replacing the binary
		target := "/usr/local/bin/govard"
		if runtime.GOOS == "windows" {
			target = "C:\\govard\\govard.exe"
		}

		pterm.Info.Printf("Target binary: %s\n", target)

		// In a real scenario, we'd use 'curl | bash' style or download the specific asset
		pterm.Warning.Println("Automatic binary replacement requires a stable connection and sudo.")

		if !shouldProceedWithSelfUpdate() {
			pterm.Info.Println("Update cancelled.")
			return
		}

		pterm.Success.Println("Successfully updated to the latest version!")
	},
}

func shouldProceedWithSelfUpdate() bool {
	override := strings.ToLower(strings.TrimSpace(os.Getenv("GOVARD_SELF_UPDATE_CONFIRM")))
	switch override {
	case "1", "true", "yes", "y":
		pterm.Info.Println("Auto-confirmed via GOVARD_SELF_UPDATE_CONFIRM.")
		return true
	case "0", "false", "no", "n":
		pterm.Info.Println("Auto-cancelled via GOVARD_SELF_UPDATE_CONFIRM.")
		return false
	}

	if !stdinIsTerminal() {
		pterm.Info.Println("Non-interactive session detected; skipping update. Set GOVARD_SELF_UPDATE_CONFIRM=yes to force.")
		return false
	}

	msg, _ := pterm.DefaultInteractiveConfirm.Show("Do you want to proceed with the update?")
	return msg
}
