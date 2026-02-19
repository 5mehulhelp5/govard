package cmd

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"

	"govard/internal/proxy"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

const proxyContainerName = "proxy-caddy-1"

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Manage the Govard Caddy proxy",
}

var proxyStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Govard Caddy proxy container",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("Starting Govard Proxy")
		if err := runDocker("start", proxyContainerName); err != nil {
			handleProxyError("start", err)
			return
		}
		pterm.Success.Println("✅ Proxy started.")
	},
}

var proxyStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Govard Caddy proxy container",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("Stopping Govard Proxy")
		if err := runDocker("stop", proxyContainerName); err != nil {
			handleProxyError("stop", err)
			return
		}
		pterm.Success.Println("✅ Proxy stopped.")
	},
}

var proxyRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Govard Caddy proxy container",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("Restarting Govard Proxy")
		if err := runDocker("stop", proxyContainerName); err != nil {
			handleProxyError("stop", err)
			return
		}
		if err := runDocker("start", proxyContainerName); err != nil {
			handleProxyError("start", err)
			return
		}
		pterm.Success.Println("✅ Proxy restarted.")
	},
}

var proxyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the Govard Caddy proxy status",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("Govard Proxy Status")
		running, err := isProxyRunning()
		if err != nil {
			pterm.Error.Printf("Failed to check proxy status: %v\n", err)
			return
		}
		if running {
			pterm.Success.Printf("✅ %s is running.\n", proxyContainerName)
			return
		}
		pterm.Warning.Printf("⚠️ %s is not running.\n", proxyContainerName)
	},
}

var proxyRoutesCmd = &cobra.Command{
	Use:   "routes",
	Short: "Register global Govard proxy routes",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("Registering Govard Proxy Routes")
		if err := proxy.RegisterDomain("mail.govard.test", "proxy-mail-1:8025"); err != nil {
			pterm.Error.Printf("Failed to register mail route: %v\n", err)
			return
		}
		if err := proxy.RegisterDomain("pma.govard.test", "proxy-pma-1:80"); err != nil {
			pterm.Error.Printf("Failed to register pma route: %v\n", err)
			return
		}
		pterm.Success.Println("✅ Proxy routes registered.")
	},
}

func init() {
	proxyCmd.AddCommand(proxyStartCmd)
	proxyCmd.AddCommand(proxyStopCmd)
	proxyCmd.AddCommand(proxyRestartCmd)
	proxyCmd.AddCommand(proxyStatusCmd)
	proxyCmd.AddCommand(proxyRoutesCmd)
}

// ProxyCommand exposes the proxy command for testing.
func ProxyCommand() *cobra.Command {
	return proxyCmd
}

func runDocker(args ...string) error {
	c := exec.Command("docker", args...)
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	return c.Run()
}

func isProxyRunning() (bool, error) {
	c := exec.Command("docker", "ps", "--filter", "name="+proxyContainerName, "--format", "{{.Names}}")
	out, err := c.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == proxyContainerName, nil
}

func handleProxyError(action string, err error) {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		pterm.Error.Printf("Failed to %s proxy: %v\n", action, err)
		return
	}

	msg := strings.ToLower(string(bytes.TrimSpace(exitErr.Stderr)))
	if strings.Contains(msg, "no such container") {
		pterm.Warning.Printf("Proxy container not found. Run `govard up` to create it.\n")
		return
	}

	pterm.Error.Printf("Failed to %s proxy: %v\n", action, err)
}
