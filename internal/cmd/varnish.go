package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var varnishCmd = &cobra.Command{
	Use:   "varnish [log|ban|stats]",
	Short: "Varnish utility commands",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config := loadConfig()
		containerName := fmt.Sprintf("%s-varnish-1", config.ProjectName)

		subcommand := args[0]
		switch subcommand {
		case "log":
			pterm.Info.Println("Streaming Varnish logs...")
			runVarnishCmd(containerName, []string{"varnishlog"})
		case "stats":
			runVarnishCmd(containerName, []string{"varnishstat"})
		case "ban":
			if len(args) < 2 {
				pterm.Error.Println("Usage: govard varnish ban <pattern>")
				pterm.Description.Println("Example: govard varnish ban /.*")
				return
			}
			pattern := args[1]
			pterm.Info.Printf("Banning pattern: %s\n", pattern)
			// varnishadm ban "req.url ~ /.*"
			banCmd := fmt.Sprintf("req.url ~ %s", pattern)
			runVarnishCmd(containerName, []string{"varnishadm", "ban", banCmd})
			pterm.Success.Println("Ban command sent to Varnish")
		default:
			pterm.Error.Printf("Unknown varnish subcommand: %s\n", subcommand)
		}
	},
}

func runVarnishCmd(containerName string, args []string) {
	dockerArgs := []string{"exec", "-it", containerName}
	dockerArgs = append(dockerArgs, args...)

	c := exec.Command("docker", dockerArgs...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	c.Run()
}
