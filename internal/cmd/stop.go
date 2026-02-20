package cmd

import (
	"fmt"
	"govard/internal/engine"
	"govard/internal/proxy"
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop project containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		pterm.DefaultHeader.Println("Stopping Govard Environment")

		config := loadConfig()
		cwd, _ := os.Getwd()
		composePath := engine.ComposeFilePath(cwd, config.ProjectName)
		if err := engine.RunHooks(config, engine.HookPreStop, cmd.OutOrStdout(), cmd.ErrOrStderr()); err != nil {
			return fmt.Errorf("pre-stop hooks failed: %w", err)
		}

		c := exec.Command("docker", "compose", "--project-directory", cwd, "-f", composePath, "stop")
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("failed to stop containers: %w", err)
		}

		if config.Domain != "" {
			if err := proxy.UnregisterDomain(config.Domain); err != nil {
				pterm.Warning.Printf("Could not remove proxy route for %s: %v\n", config.Domain, err)
			}
			if err := engine.RemoveHostsEntry(config.Domain); err != nil {
				pterm.Warning.Printf("Could not remove hosts entry for %s: %v\n", config.Domain, err)
			}
		}

		if err := engine.RunHooks(config, engine.HookPostStop, cmd.OutOrStdout(), cmd.ErrOrStderr()); err != nil {
			return fmt.Errorf("post-stop hooks failed: %w", err)
		}

		pterm.Success.Println("✅ Environment stopped.")
		return nil
	},
}
