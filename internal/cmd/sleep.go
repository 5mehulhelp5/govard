package cmd

import "github.com/spf13/cobra"

var sleepCmd = &cobra.Command{
	Use:   "sleep",
	Short: "Stop running projects and persist sleep state",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSleep()
	},
}
