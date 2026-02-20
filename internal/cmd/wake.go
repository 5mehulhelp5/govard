package cmd

import "github.com/spf13/cobra"

var wakeCmd = &cobra.Command{
	Use:   "wake",
	Short: "Start projects recorded in sleep state",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWake()
	},
}
