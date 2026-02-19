package cmd

import (
	"fmt"
	"govard/internal/engine"
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var errorFilter bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View project logs",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("Govard Log Stream")
		config := loadConfig()
		cwd, _ := os.Getwd()
		composePath := engine.ComposeFilePath(cwd, config.ProjectName)

		dockerArgs := []string{"compose", "--project-directory", cwd, "-f", composePath, "logs", "-f", "--tail=100"}

		if errorFilter {
			pterm.Info.Println("Filtering for errors...")
			// Simple grep for error-like strings
			filterCommand := fmt.Sprintf(
				"docker compose --project-directory %q -f %q logs -f | grep -iE 'error|critical|fail|exception'",
				cwd,
				composePath,
			)
			c := exec.Command("sh", "-c", filterCommand)
			c.Stdout, c.Stderr = os.Stdout, os.Stderr
			c.Run()
			return
		}

		c := exec.Command("docker", dockerArgs...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		c.Run()
	},
}
