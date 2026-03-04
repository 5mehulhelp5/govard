package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var redisCmd = &cobra.Command{
	Use:   "redis [args]",
	Short: "Interact with the redis container using redis-cli",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := loadConfig()
		containerName := fmt.Sprintf("%s-redis-1", config.ProjectName)
		serviceLabel := "Redis"
		cliBinary := "redis-cli"
		if config.Stack.Services.Cache == "valkey" {
			cliBinary = "valkey-cli"
			serviceLabel = "Valkey"
		}

		if err := ensureContainerReadyForExec(containerName, serviceLabel); err != nil {
			return err
		}

		pterm.Info.Printf("Connecting to %s on %s...\n", serviceLabel, containerName)

		dockerArgs := dockerExecBaseArgs()
		dockerArgs = append(dockerArgs, containerName, cliBinary)
		dockerArgs = append(dockerArgs, args...)

		c := exec.Command("docker", dockerArgs...)
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr

		if err := c.Run(); err != nil {
			if stateErr := ensureContainerReadyForExec(containerName, serviceLabel); stateErr != nil {
				return fmt.Errorf("%s CLI failed: %w", serviceLabel, stateErr)
			}
			pterm.Debug.Printf("%s CLI exited with error: %v\n", serviceLabel, err)
			return fmt.Errorf("%s CLI failed: %w", serviceLabel, err)
		}
		return nil
	},
}
