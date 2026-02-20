package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var redisCmd = &cobra.Command{
	Use:   "redis [args]",
	Short: "Interact with the redis container using redis-cli",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := loadConfig()
		containerName := fmt.Sprintf("%s-redis-1", config.ProjectName)
		cliBinary := "redis-cli"
		if config.Stack.Services.Cache == "valkey" {
			cliBinary = "valkey-cli"
		}

		// Check if container exists and is running
		check := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
		if output, err := check.Output(); err != nil || strings.TrimSpace(string(output)) != "true" {
			return fmt.Errorf("redis container %s is not running", containerName)
		}

		pterm.Info.Printf("Connecting to Redis on %s...\n", containerName)

		// Prepare redis-cli command
		dockerArgs := []string{"exec", "-it", containerName, cliBinary}
		dockerArgs = append(dockerArgs, args...)

		c := exec.Command("docker", dockerArgs...)
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr

		if err := c.Run(); err != nil {
			pterm.Debug.Printf("redis-cli exited with error: %v\n", err)
			return fmt.Errorf("redis cli failed: %w", err)
		}
		return nil
	},
}
