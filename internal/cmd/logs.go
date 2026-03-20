package cmd

import (
	"fmt"
	"govard/internal/engine"
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	errorFilter bool
	tailCount   int
)

var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View project logs",
	Long: `Streams real-time logs from all services or a specific service in the current project environment.
It automatically aggregates logs from PHP, Web, MySQL, Redis, and other active containers.

Case Studies:
- Real-time Debugging: Watch the log stream while browsing the site to catch errors as they happen.
- Post-Mortem: Check the last 100 lines (default) to see why a container crashed.
- Error Hunting: Use --errors to filter out noise and only show critical failure messages.
- Service Focus: Use 'govard env logs php' to only see PHP-related logs.`,
	Example: `  # Follow all project logs
  govard env logs

  # Follow logs for a specific service
  govard env logs php

  # Show last 200 lines and follow
  govard env logs --tail 200

  # Show only error messages
  govard env logs --errors`,
	RunE: func(cmd *cobra.Command, args []string) error {
		service := ""
		if len(args) > 0 {
			service = args[0]
		}

		pterm.DefaultHeader.Println("Govard Log Stream")
		config := loadConfig()
		cwd, _ := os.Getwd()
		composePath := engine.ComposeFilePath(cwd, config.ProjectName)

		if errorFilter {
			pterm.Info.Println("Filtering for errors...")
			servicePart := ""
			if service != "" {
				servicePart = " " + shellQuote(service)
			}
			// Simple grep for error-like strings
			filterCommand := fmt.Sprintf(
				"docker compose --project-directory %q -p %q -f %q logs -f --tail=%d%s | grep -iE 'error|critical|fail|exception'",
				cwd,
				config.ProjectName,
				composePath,
				tailCount,
				servicePart,
			)
			c := exec.Command("sh", "-c", filterCommand)
			c.Stdout, c.Stderr = os.Stdout, os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("stream filtered logs: %w", err)
			}
			return nil
		}

		dockerArgs := []string{
			"compose",
			"--project-directory", cwd,
			"-p", config.ProjectName,
			"-f", composePath,
			"logs", "-f",
			fmt.Sprintf("--tail=%d", tailCount),
		}
		if service != "" {
			dockerArgs = append(dockerArgs, service)
		}

		c := exec.Command("docker", dockerArgs...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("stream logs: %w", err)
		}
		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVar(&errorFilter, "errors", false, "Filter logs for error messages")
	logsCmd.Flags().IntVar(&tailCount, "tail", 100, "Number of lines to show from the end of the logs")
}
