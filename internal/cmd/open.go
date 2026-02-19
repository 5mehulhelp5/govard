package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const openSupportedTargets = "admin, db, shell, sftp, elasticsearch, opensearch"

var openEnvironment string

var openCmd = &cobra.Command{
	Use:   "open [target]",
	Short: "Open common service URLs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config := loadFullConfig()
		target := strings.ToLower(strings.TrimSpace(args[0]))

		switch target {
		case "admin":
			return runOpenAdminTarget(config, openEnvironment)
		case "db":
			return runOpenDBTarget(config, openEnvironment)
		case "shell":
			return runOpenShellTarget(config, openEnvironment)
		case "sftp":
			return runOpenSFTPTarget(config, openEnvironment)
		case "elasticsearch", "opensearch":
			return runOpenSearchTarget(config, target, openEnvironment)
		default:
			return fmt.Errorf("unknown target %q. Supported: %s", target, openSupportedTargets)
		}
	},
}

func init() {
	openCmd.Flags().StringVarP(
		&openEnvironment,
		"environment",
		"e",
		"",
		"Environment for open targets (local or configured remote name/env)",
	)
}
