package cmd

import (
	"fmt"
	"govard/internal/ui"
	"os"

	"github.com/spf13/cobra"
)

const Version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "govard",
	Short: "Govard is a high-performance local development orchestrator",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "help" {
			ui.PrintBrand(Version)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Govard",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintBrand(Version)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(bootstrapCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&errorFilter, "errors", "e", false, "Filter logs for errors only")
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(fixDepsCmd)
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(trustCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(redisCmd)
	rootCmd.AddCommand(valkeyCmd)
	rootCmd.AddCommand(elasticsearchCmd)
	rootCmd.AddCommand(opensearchCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(desktopCmd)

	// Framework & Tooling Shortcuts
	initFrameworkCommands()

	rootCmd.AddCommand(mailCmd)
	rootCmd.AddCommand(pmaCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(varnishCmd)
	rootCmd.AddCommand(proxyCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(extensionsCmd)
	registerProjectCustomCommands()
	rootCmd.AddCommand(customCmd)
	rootCmd.AddCommand(selfUpdateCmd)
	rootCmd.AddCommand(versionCmd)
}
