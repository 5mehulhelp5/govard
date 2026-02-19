package cmd

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"govard/internal/engine"
)

var trustCmd = &cobra.Command{
	Use:   "trust",
	Short: "Trust the local CA for SSL certificates",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("Govard SSL Trust Store")

		if err := engine.TrustCA(); err != nil {
			pterm.Error.Printf("Failed to trust CA: %v\n", err)
			return
		}

		pterm.Success.Println("🛡️ Root CA successfully installed into system trust store!")
	},
}
