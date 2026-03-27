package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"govard/internal/engine"
	"govard/internal/proxy"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env [command]",
	Short: "Control project environment via docker compose",
	Long: `Manage the lifecycle and services of your project's development environment.
All commands are scoped to the project in the current working directory.
It provides smart wrappers around Docker Compose operations and specialized service interactions.

Govard intelligently proxies almost all Docker Compose commands. If a command is not
explicitly handled by Govard, it is passed through to 'docker compose' with the 
correct project context.

Aliases: project

Case Studies:
- Maintenance: Use 'govard env stop' to pause work and 'govard env start' to resume later.
- Troubleshooting: Check 'govard env ps' and 'govard env logs' to identify failing services.
- Cache Management: Run 'govard redis flush' to clear local cache.
- Cleanup: Run 'govard env down -v' to completely remove the environment and its data.`,
	Example: `  # Start the project environment
  govard env up

  # View help for all supported compose commands
  govard env --help

  # List running containers for this project
  govard env ps

  # View real-time logs for all services
  govard env logs -f

  # Enter a Redis shell for the current project
  govard redis cli`,
	Args:    cobra.ArbitraryArgs,
	Aliases: []string{"project"},
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			proxyArgs := []string{}
			for i, arg := range os.Args {
				if (arg == "env" || arg == "project") && i+1 < len(os.Args) {
					proxyArgs = os.Args[i+1:]
					break
				}
			}
			if len(proxyArgs) > 0 {
				return proxyEnvToCompose(cmd, proxyArgs)
			}
		}
		return cmd.Help()
	},
}

func proxyEnvToCompose(cmd *cobra.Command, args []string) error {
	subcommand := args[0]
	config := loadConfig()
	cwd, _ := os.Getwd()
	composePath := engine.ComposeFilePath(cwd, config.ProjectName)

	switch subcommand {
	case "start":
		fmt.Println()
		pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack, pterm.Bold).Println(" Starting Govard Project ")
		fmt.Println()
		args = append([]string{"up", "-d"}, args[1:]...)
	case "restart":
		fmt.Println()
		pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack, pterm.Bold).Println(" Restarting Govard Project ")
		fmt.Println()
		if err := proxyEnvToCompose(cmd, []string{"stop"}); err != nil {
			return err
		}
		return proxyEnvToCompose(cmd, append([]string{"up", "-d"}, args[1:]...))
	case "stop", "down":
		action := "Stopping"
		if subcommand == "down" {
			action = "Tearing Down"
		}
		fmt.Println()
		pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack, pterm.Bold).Printf(" %s Govard Environment \n", action)
		fmt.Println()

		if err := engine.RunHooks(config, engine.HookPreStop, cmd.OutOrStdout(), cmd.ErrOrStderr()); err != nil {
			return fmt.Errorf("pre-stop hooks failed: %w", err)
		}

		err := engine.RunCompose(cmd.Context(), engine.ComposeOptions{
			ProjectDir:  cwd,
			ProjectName: config.ProjectName,
			ComposeFile: composePath,
			Args:        args,
			Stdout:      cmd.OutOrStdout(),
			Stderr:      cmd.ErrOrStderr(),
			Stdin:       os.Stdin,
		})
		if err != nil {
			return fmt.Errorf("failed to %s containers: %w", strings.ToLower(subcommand), err)
		}

		for _, domain := range config.AllDomains() {
			if err := proxy.UnregisterDomain(domain); err != nil {
				pterm.Warning.Printf("Could not remove proxy route for %s: %v\n", domain, err)
			}
			if err := engine.RemoveHostsEntry(domain); err != nil {
				pterm.Warning.Printf("Could not remove hosts entry for %s: %v\n", domain, err)
			}
		}

		if err := engine.RunHooks(config, engine.HookPostStop, cmd.OutOrStdout(), cmd.ErrOrStderr()); err != nil {
			return fmt.Errorf("post-stop hooks failed: %w", err)
		}

		pterm.Success.Printf("✅ Environment %s.\n", strings.ToLower(action))
		return nil

	case "logs":
		// Check for --errors flag
		hasErrorFilter := false
		for _, arg := range args {
			if arg == "--errors" {
				hasErrorFilter = true
				break
			}
		}

		if hasErrorFilter {
			fmt.Println()
			pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack, pterm.Bold).Println(" Govard Log Stream (Errors Only) ")
			fmt.Println()
			pterm.Info.Println("Filtering for errors...")

			// Rebuild args without --errors
			filteredArgs := []string{}
			for _, arg := range args {
				if arg != "--errors" {
					filteredArgs = append(filteredArgs, arg)
				}
			}

			// Construct manual grep command
			composeCmd := fmt.Sprintf("docker compose --project-directory %s -p %s -f %s",
				shellQuote(cwd), shellQuote(config.ProjectName), shellQuote(composePath))

			logArgs := strings.Join(filteredArgs, " ")
			if !strings.Contains(logArgs, "-f") && !strings.Contains(logArgs, "--follow") {
				logArgs += " -f --tail=100"
			}

			filterCommand := fmt.Sprintf("%s %s | grep -iE 'error|critical|fail|exception'", composeCmd, logArgs)
			c := exec.Command("sh", "-c", filterCommand)
			c.Stdout, c.Stderr = os.Stdout, os.Stderr
			return c.Run()
		}

	case "pull":
		fmt.Println()
		pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack, pterm.Bold).Println(" Pulling Project Images ")
		fmt.Println()
		err := engine.RunCompose(cmd.Context(), engine.ComposeOptions{
			ProjectDir:  cwd,
			ProjectName: config.ProjectName,
			ComposeFile: composePath,
			Args:        args,
			Stdout:      cmd.OutOrStdout(),
			Stderr:      cmd.ErrOrStderr(),
			Stdin:       os.Stdin,
		})
		if err != nil {
			pterm.Warning.Printf("docker compose pull failed: %v\n", err)
			pterm.Info.Println("Attempting local Govard image build fallback...")

			built, fallbackErr := engine.FallbackBuildMissingGovardImagesFromCompose(composePath, cmd.OutOrStdout(), cmd.ErrOrStderr())
			if fallbackErr != nil {
				return fmt.Errorf("pull project images: %w (local fallback failed: %v)", err, fallbackErr)
			}

			if len(built) > 0 {
				pterm.Success.Printf("Local fallback built %d image(s): %v\n", len(built), built)
			} else {
				pterm.Info.Println("No missing Govard-managed images required local build.")
			}
		}
		pterm.Success.Println("✅ Images pulled.")
		return nil
	case "cleanup":
		fmt.Println()
		pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack, pterm.Bold).Println(" Cleaning Up Compose Cache ")
		fmt.Println()
		count, err := engine.CleanupStaleComposeFiles(7 * 24 * time.Hour)
		if err != nil {
			return fmt.Errorf("cleanup compose cache: %w", err)
		}
		pterm.Success.Printf("✅ Removed %d stale compose file(s).\n", count)
		return nil
	}

	// Default proxy for everything else
	return engine.RunCompose(cmd.Context(), engine.ComposeOptions{
		ProjectDir:  cwd,
		ProjectName: config.ProjectName,
		ComposeFile: composePath,
		Args:        args,
		Stdout:      cmd.OutOrStdout(),
		Stderr:      cmd.ErrOrStderr(),
		Stdin:       os.Stdin,
	})
}

func init() {
	envCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		rebrandComposeHelp(cmd, "env")
	})

	// Non-standard shortcuts
	envCmd.AddCommand(upCmd)
	envCmd.AddCommand(redisCmd)
	envCmd.AddCommand(valkeyCmd)
	envCmd.AddCommand(elasticsearchCmd)
	envCmd.AddCommand(opensearchCmd)
	envCmd.AddCommand(varnishCmd)

	envCmd.AddCommand(envCleanupCmd)
}

var envCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Prune stale docker-compose files and manifests from Govard home",
	RunE: func(cmd *cobra.Command, args []string) error {
		return proxyEnvToCompose(cmd, []string{"cleanup"})
	},
}
