package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"govard/internal/engine"
	"govard/internal/engine/remote"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var redisCmd = &cobra.Command{
	Use:   "redis [command]",
	Short: "Control the redis cache service",
	Long: `Interact with the Redis or Valkey cache service. 
Supports both custom utility commands (flush, info, cli) and standard Docker Compose maintenance commands (ps, logs, stop, start, etc.).`,
	Example: `  # Open a redis CLI
  govard env redis cli

  # Flush all keys
  govard env redis flush

  # View redis logs
  govard env redis logs -f

  # Check redis status
  govard v redis ps`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
				return cmd.Help()
			}
			if isComposeMaintenanceCommand(args[0]) {
				return proxyServiceToCompose(cmd, "redis", args)
			}
			// Fallback to direct command execution (e.g. "govard env redis PING")
			return runRedisCommand(cmd, args)
		}
		return cmd.Help()
	},
}

var redisFlushCmd = &cobra.Command{
	Use:   "flush",
	Short: "Flush all keys from the cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRedisCommand(cmd, []string{"flushall"})
	},
}

var redisInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display cache information",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRedisCommand(cmd, []string{"info"})
	},
}

var redisCliCmd = &cobra.Command{
	Use:                "cli [args]",
	Short:              "Open an interactive CLI or run a command",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRedisCommand(cmd, args)
	},
}

func init() {
	redisCmd.AddCommand(redisFlushCmd)
	redisCmd.AddCommand(redisInfoCmd)
	redisCmd.AddCommand(redisCliCmd)
}

func runRedisCommand(cmd *cobra.Command, args []string) error {
	environment, _ := cmd.Flags().GetString("environment")
	environment = strings.ToLower(strings.TrimSpace(environment))
	if environment == "" {
		environment = "local"
	}

	config := loadConfig()
	cliBinary := "redis-cli"
	serviceLabel := "Redis"
	if config.Stack.Services.Cache == "valkey" {
		cliBinary = "valkey-cli"
		serviceLabel = "Valkey"
	}

	if environment == "local" {
		containerName := fmt.Sprintf("%s-redis-1", config.ProjectName)
		if err := ensureContainerReadyForExec(containerName, serviceLabel); err != nil {
			return err
		}

		dockerArgs := dockerExecBaseArgs()
		dockerArgs = append(dockerArgs, containerName, cliBinary)
		dockerArgs = append(dockerArgs, args...)

		c := exec.Command("docker", dockerArgs...)
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		return c.Run()
	}

	// Remote environment
	resolvedName, remoteCfg, err := ensureRemoteKnown(config, environment)
	if err != nil {
		return err
	}
	environment = resolvedName

	if !engine.RemoteCapabilityEnabled(remoteCfg, engine.RemoteCapabilityCache) {
		return fmt.Errorf("remote '%s' does not allow cache operations", environment)
	}

	cmdStr := cliBinary
	for _, arg := range args {
		cmdStr += " " + engine.ShellQuote(arg)
	}

	pterm.Info.Printf("Running %s on remote %s...\n", cmdStr, environment)
	sshCmd := remote.BuildSSHExecCommand(environment, remoteCfg, true, cmdStr)
	sshCmd.Stdin, sshCmd.Stdout, sshCmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return sshCmd.Run()
}
