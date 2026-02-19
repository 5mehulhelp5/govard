package cmd

import (
	"fmt"
	"govard/internal/engine"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Auto-configure Magento env.php",
	RunE: func(cmd *cobra.Command, args []string) error {
		pterm.DefaultHeader.Println("Govard Auto-Configuration")

		// 1. Load layered config
		config := loadFullConfig()

		if config.Recipe != "magento2" {
			pterm.Warning.Printf("Configuration injection is currently only supported for Magento 2. Detected: %s\n", config.Recipe)
			return nil
		}

		if err := engine.ConfigureMagento(config.ProjectName, config); err != nil {
			return fmt.Errorf("configuration failed: %w", err)
		}
		return nil
	},
}
