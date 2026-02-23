package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configAutoCmd = &cobra.Command{
	Use:   "auto",
	Short: "Auto-configure Magento env.php",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if configureCmd.RunE == nil {
			return fmt.Errorf("config auto is unavailable")
		}
		return configureCmd.RunE(cmd, args)
	},
}

var configProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Show recommended runtime profile for the detected framework",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if profileCmd.RunE == nil {
			return fmt.Errorf("config profile is unavailable")
		}
		return profileCmd.RunE(cmd, args)
	},
}

var configProfileApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the recommended runtime profile to govard.yml",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if profileApplyCmd.RunE == nil {
			return fmt.Errorf("config profile apply is unavailable")
		}
		return profileApplyCmd.RunE(cmd, args)
	},
}

func init() {
	registerProfileFlags(configProfileCmd)
	configProfileCmd.AddCommand(configProfileApplyCmd)

	configCmd.AddCommand(configAutoCmd)
	configCmd.AddCommand(configProfileCmd)
}
