package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"govard/internal/engine"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env [command]",
	Short: "Control project environment via docker compose",
	Long: `Manage the lifecycle and services of your project's development environment.
All commands are scoped to the project in the current working directory.
It provides wrappers around common Docker Compose operations and specialized service interactions.

Case Studies:
- Maintenance: Use 'govard env stop' to pause work and 'govard env start' to resume later.
- Troubleshooting: Check 'govard env ps' and 'govard env logs' to identify failing services.
- Cache Management: Run 'govard env redis-cli flushall' to clear local cache.
- Cleanup: Run 'govard env down -v' to completely remove the environment and its data.`,
	Example: `  # Start the project environment
  govard env up

  # List running containers for this project
  govard env ps

  # View real-time logs for all services
  govard env logs

  # Enter a Redis shell for the current project
  govard env redis`,
}

var envStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start existing project containers",
	Args:  cobra.NoArgs,
	RunE:  runEnvStart,
}

var envRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the project environment",
	Args:  cobra.NoArgs,
	RunE:  runEnvRestart,
}

var envPsCmd = &cobra.Command{
	Use:   "ps",
	Short: "List project containers",
	Args:  cobra.NoArgs,
	RunE:  runEnvPs,
}

var envPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull latest project images",
	Args:  cobra.NoArgs,
	RunE:  runEnvPull,
}

func runEnvStart(cmd *cobra.Command, args []string) error {
	config := loadConfig()
	cwd, _ := os.Getwd()
	composePath := engine.ComposeFilePath(cwd, config.ProjectName)

	// Use "up -d" instead of "start" to reconcile stale container definitions
	// and recover services that require recreation.
	command := exec.Command(
		"docker",
		"compose",
		"--project-directory",
		cwd,
		"-p",
		config.ProjectName,
		"-f",
		composePath,
		"up",
		"-d",
	)
	command.Stdout = cmd.OutOrStdout()
	command.Stderr = cmd.ErrOrStderr()
	if err := command.Run(); err != nil {
		return fmt.Errorf("start project containers: %w", err)
	}
	pterm.Success.Println("✅ Environment started.")
	return nil
}

func runEnvRestart(cmd *cobra.Command, args []string) error {
	if err := stopCmd.RunE(cmd, args); err != nil {
		return err
	}
	return upCmd.RunE(cmd, args)
}

func runEnvPs(cmd *cobra.Command, args []string) error {
	config := loadConfig()
	cwd, _ := os.Getwd()
	composePath := engine.ComposeFilePath(cwd, config.ProjectName)

	command := exec.Command(
		"docker",
		"compose",
		"--project-directory",
		cwd,
		"-p",
		config.ProjectName,
		"-f",
		composePath,
		"ps",
	)
	command.Stdout = cmd.OutOrStdout()
	command.Stderr = cmd.ErrOrStderr()
	if err := command.Run(); err != nil {
		return fmt.Errorf("list project containers: %w", err)
	}
	return nil
}

func runEnvPull(cmd *cobra.Command, args []string) error {
	config := loadConfig()
	cwd, _ := os.Getwd()
	composePath := engine.ComposeFilePath(cwd, config.ProjectName)

	command := exec.Command(
		"docker",
		"compose",
		"--project-directory",
		cwd,
		"-p",
		config.ProjectName,
		"-f",
		composePath,
		"pull",
	)
	command.Stdout = cmd.OutOrStdout()
	command.Stderr = cmd.ErrOrStderr()
	if err := command.Run(); err != nil {
		return fmt.Errorf("pull project images: %w", err)
	}
	pterm.Success.Println("✅ Images pulled.")
	return nil
}

func init() {
	envCmd.AddCommand(upCmd)
	envCmd.AddCommand(envStartCmd)
	envCmd.AddCommand(stopCmd)
	envCmd.AddCommand(downCmd)
	envCmd.AddCommand(envRestartCmd)
	envCmd.AddCommand(envPsCmd)
	envCmd.AddCommand(envPullCmd)
	envCmd.AddCommand(logsCmd)
	envCmd.AddCommand(redisCmd)
	envCmd.AddCommand(valkeyCmd)
	envCmd.AddCommand(elasticsearchCmd)
	envCmd.AddCommand(opensearchCmd)
	envCmd.AddCommand(varnishCmd)
}
