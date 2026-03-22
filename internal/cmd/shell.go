package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:                "shell",
	Short:              "Enter the application container",
	DisableFlagParsing: true,
	SilenceUsage:       true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
			return cmd.Help()
		}
		config := loadConfig()
		containerName := fmt.Sprintf("%s-php-1", config.ProjectName)
		user := ResolveProjectExecUser(config, "www-data")

		// If no arguments, we're starting an interactive session.
		// We use a subshell trick to set a colored PS1 (cyan for user@host)
		// which matches Warden's style.
		if len(args) == 0 {
			coloredPS1 := "\\[\\033[01;36m\\]\\u@\\h\\[\\033[00m\\]:\\w\\$ "
			bashCmd := fmt.Sprintf("export PS1='%s'; exec bash", coloredPS1)
			err := RunInContainer(containerName, user, "bash", []string{"-c", bashCmd})
			if err == nil {
				return nil
			}

			if exitErr, ok := err.(*exec.ExitError); ok {
				code := exitErr.ExitCode()
				if code == 126 || code == 127 {
					// bash not found — try sh as plain fallback
					err = RunInContainer(containerName, user, "sh", args)
				} else {
					os.Exit(code)
				}
			}
			return err
		}

		// Try bash first; only fallback to sh if bash is not found (126/127)
		err := RunInContainer(containerName, user, "bash", args)
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code == 126 || code == 127 {
				// bash not found or not executable — try sh
				err = RunInContainer(containerName, user, "sh", args)
			} else {
				// Passthrough the exit code without printing Cobra error/usage
				os.Exit(code)
			}
		}

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			return err
		}
		return nil
	},
}

func init() {
}
