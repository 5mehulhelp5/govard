package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"govard/internal/engine"
	"govard/internal/proxy"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

const globalProxyProjectName = "proxy"

var errGlobalServicesNotInitialized = errors.New("global services are not initialized")

var svcCmd = &cobra.Command{
	Use:   "svc",
	Short: "Manage global services and workspace sleep state",
}

var svcUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Start global services (proxy, mailpit, pma)",
	Args:  cobra.NoArgs,
	RunE:  runSvcUp,
}

var svcDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop global services (proxy, mailpit, pma)",
	Args:  cobra.NoArgs,
	RunE:  runSvcDown,
}

var svcRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart global services (proxy, mailpit, pma)",
	Args:  cobra.NoArgs,
	RunE:  runSvcRestart,
}

var svcPsCmd = &cobra.Command{
	Use:   "ps",
	Short: "List running global service containers",
	Args:  cobra.NoArgs,
	RunE:  runSvcPs,
}

var svcLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Tail logs for global services",
	Args:  cobra.NoArgs,
	RunE:  runSvcLogs,
}

var svcSleepCmd = &cobra.Command{
	Use:   "sleep",
	Short: "Stop all running Govard projects and persist wake state",
	Args:  cobra.NoArgs,
	RunE:  runSvcSleep,
}

var svcWakeCmd = &cobra.Command{
	Use:   "wake",
	Short: "Start all projects recorded in sleep state",
	Args:  cobra.NoArgs,
	RunE:  runSvcWake,
}

func runSvcUp(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.Println("Starting Govard Global Services")

	if err := engine.EnsureGlobalProxy(); err != nil {
		return fmt.Errorf("ensure global proxy: %w", err)
	}

	if err := runGlobalProxyCompose(cmd, "up", "-d"); err != nil {
		return fmt.Errorf("start global services: %w", err)
	}

	if err := registerGlobalServiceRoutes(); err != nil {
		pterm.Warning.Printf("Could not refresh global proxy routes: %v\n", err)
	}

	pterm.Success.Println("✅ Global services are running.")
	return nil
}

func runSvcDown(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.Println("Stopping Govard Global Services")

	if err := runGlobalProxyCompose(cmd, "down"); err != nil {
		if errors.Is(err, errGlobalServicesNotInitialized) {
			pterm.Warning.Println("Global services are not initialized yet. Run `govard svc up` first.")
			return nil
		}
		return fmt.Errorf("stop global services: %w", err)
	}

	pterm.Success.Println("✅ Global services stopped.")
	return nil
}

func runSvcRestart(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.Println("Restarting Govard Global Services")

	if err := runSvcDown(cmd, args); err != nil {
		return err
	}

	return runSvcUp(cmd, args)
}

func runSvcPs(cmd *cobra.Command, args []string) error {
	if err := runGlobalProxyCompose(cmd, "ps"); err != nil {
		if errors.Is(err, errGlobalServicesNotInitialized) {
			pterm.Warning.Println("Global services are not initialized yet. Run `govard svc up` first.")
			return nil
		}
		return fmt.Errorf("list global services: %w", err)
	}
	return nil
}

func runSvcLogs(cmd *cobra.Command, args []string) error {
	if err := runGlobalProxyCompose(cmd, "logs", "-f", "--tail=100"); err != nil {
		if errors.Is(err, errGlobalServicesNotInitialized) {
			return fmt.Errorf("global services are not initialized yet, run `govard svc up` first")
		}
		return fmt.Errorf("stream global service logs: %w", err)
	}
	return nil
}

func runSvcSleep(cmd *cobra.Command, args []string) error {
	return runSleep()
}

func runSvcWake(cmd *cobra.Command, args []string) error {
	return runWake()
}

func runGlobalProxyCompose(cmd *cobra.Command, args ...string) error {
	composeFile := globalProxyComposeFilePath()
	composeDir := globalProxyComposeDirPath()

	if _, err := os.Stat(composeFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s", errGlobalServicesNotInitialized, composeFile)
		}
		return fmt.Errorf("stat global compose file: %w", err)
	}

	dockerArgs := []string{
		"compose",
		"--project-directory",
		composeDir,
		"-p",
		globalProxyProjectName,
		"-f",
		composeFile,
	}
	dockerArgs = append(dockerArgs, args...)

	command := exec.Command("docker", dockerArgs...)
	command.Stdout = cmd.OutOrStdout()
	command.Stderr = cmd.ErrOrStderr()
	return command.Run()
}

func registerGlobalServiceRoutes() error {
	if err := proxy.RegisterDomain("mail.govard.test", "proxy-mail-1:8025"); err != nil {
		return fmt.Errorf("register mail route: %w", err)
	}
	if err := proxy.RegisterDomain("pma.govard.test", "proxy-pma-1:80"); err != nil {
		return fmt.Errorf("register pma route: %w", err)
	}
	return nil
}

func globalProxyComposeDirPath() string {
	return filepath.Join(os.Getenv("HOME"), ".govard", "proxy")
}

func globalProxyComposeFilePath() string {
	return filepath.Join(globalProxyComposeDirPath(), "docker-compose.yml")
}

func init() {
	svcCmd.AddCommand(svcUpCmd)
	svcCmd.AddCommand(svcDownCmd)
	svcCmd.AddCommand(svcRestartCmd)
	svcCmd.AddCommand(svcPsCmd)
	svcCmd.AddCommand(svcLogsCmd)
	svcCmd.AddCommand(svcSleepCmd)
	svcCmd.AddCommand(svcWakeCmd)
}
