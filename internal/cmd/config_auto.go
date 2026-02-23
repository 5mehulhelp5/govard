package cmd

import (
	"fmt"
	"govard/internal/engine"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var configAutoCmd = &cobra.Command{
	Use:   "auto",
	Short: "Auto-configure framework runtime files",
	RunE: func(cmd *cobra.Command, args []string) error {
		pterm.DefaultHeader.Println("Govard Auto-Configuration")

		config := loadFullConfig()
		if err := applyRecipeAutoConfiguration(config); err != nil {
			return fmt.Errorf("configuration failed: %w", err)
		}
		return nil
	},
}

func applyRecipeAutoConfiguration(config engine.Config) error {
	switch config.Recipe {
	case "magento2":
		return engine.ConfigureMagento(config.ProjectName, config)
	default:
		pterm.Warning.Printf(
			"Auto configuration is not supported for recipe %q yet.\n",
			config.Recipe,
		)
		return nil
	}
}

func init() {
	configCmd.AddCommand(configAutoCmd)
}
