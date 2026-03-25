package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:                "shell",
	Aliases:            []string{"sh"},
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

		var err error
		if len(args) == 0 {
			// Interactive session with colored PS1 trick (Cyan user@host to match Warden)
			coloredPS1 := "\\[\\033[01;36m\\]\\u@\\h\\[\\033[00m\\]:\\w\\$ "
			bashCmd := fmt.Sprintf("export PS1='%s'; exec bash", coloredPS1)
			err = RunInContainer(containerName, user, "bash", []string{"-c", bashCmd})
		} else {
			err = RunInContainer(containerName, user, "bash", args)
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code == 126 || code == 127 {
				// bash not found or not executable — try sh
				err = RunInContainer(containerName, user, "sh", args)
			}
		}

		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
}
