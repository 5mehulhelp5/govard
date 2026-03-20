package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug [on|off|status|shell]",
	Short: "Manage Xdebug for the current environment",
	Long:  `Toggle Xdebug on or off, check its status, or open a debug shell. Changes to on/off will trigger an environment update.`,
}

var debugOnCmd = &cobra.Command{
	Use:   "on",
	Short: "Enable Xdebug",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadWritableConfig()
		if err != nil {
			return err
		}
		if config.Stack.Features.Xdebug {
			pterm.Info.Println("Xdebug is already enabled")
			return nil
		}
		config.Stack.Features.Xdebug = true
		saveConfig(config)
		pterm.Success.Println("Xdebug enabled in .govard.yml. Running 'govard env up' to apply...")
		runUp()
		return nil
	},
}

var debugOffCmd = &cobra.Command{
	Use:   "off",
	Short: "Disable Xdebug",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadWritableConfig()
		if err != nil {
			return err
		}
		if !config.Stack.Features.Xdebug {
			pterm.Info.Println("Xdebug is already disabled")
			return nil
		}
		config.Stack.Features.Xdebug = false
		saveConfig(config)
		pterm.Success.Println("Xdebug disabled in .govard.yml. Running 'govard env up' to apply...")
		runUp()
		return nil
	},
}

var debugStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Xdebug status",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadFullConfig()
		if err != nil {
			return err
		}
		status := "disabled"
		if config.Stack.Features.Xdebug {
			status = "enabled"
		}
		pterm.Info.Printf("Xdebug is currently %s\n", status)
		return nil
	},
}

var debugShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open a debug shell",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadFullConfig()
		if err != nil {
			return err
		}
		if !config.Stack.Features.Xdebug {
			return fmt.Errorf("xdebug is disabled. Enable it with 'govard debug on'")
		}
		containerName := fmt.Sprintf("%s-php-debug-1", config.ProjectName)
		// Set XDEBUG_SESSION and PHP_IDE_CONFIG to trigger debugger in CLI scripts.
		// We use -c to export variables and then exec bash to keep the shell interactive.
		bashCmd := "export XDEBUG_SESSION=PHPSTORM; export PHP_IDE_CONFIG=\"serverName=govard\"; exec bash"
		return RunInContainer(containerName, ResolveProjectExecUser(config, "www-data"), "bash", []string{"-c", bashCmd})
	},
}

func init() {
	debugCmd.AddCommand(debugOnCmd)
	debugCmd.AddCommand(debugOffCmd)
	debugCmd.AddCommand(debugStatusCmd)
	debugCmd.AddCommand(debugShellCmd)
}
