package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"govard/internal/engine"

	"github.com/spf13/cobra"
)

type FrameworkCommand struct {
	Name        string
	Short       string
	Framework   string // empty means all
	Binary      string
	PrependArgs []string
	DefaultUser string
}

type commandExecutionTarget struct {
	ContainerName string
	Workdir       string
	User          string
}

var toolCmd = &cobra.Command{
	Use:   "tool [command]",
	Short: "Run framework/tooling commands inside project containers",
	Long: `Run framework CLIs and common package manager commands directly inside the project containers.
This eliminates the need to install PHP, Composer, or Node.js on your host machine.
Govard automatically routes commands to the correct container (usually PHP) and
executes them as the appropriate user (e.g., www-data).

Case Studies:
- Clean Workspace: Run 'govard tool magento setup:upgrade' without needing PHP/MySQL on your laptop.
- Unified Workflow: Use the same command regardless of which PHP version the project requires.
- Package Management: Use 'govard tool composer install' to ensure dependencies match the container environment.`,
	Example: `  # Run Magento CLI
  govard tool magento cache:flush

  # Install a composer package
  govard tool composer require monolog/monolog

  # Run npm install
  govard tool npm install`,
}

var frameworkCommands = []FrameworkCommand{
	{
		Name:        "magento",
		Short:       "Run Magento CLI commands",
		Framework:   "magento2",
		Binary:      "php",
		PrependArgs: []string{"bin/magento"},
		DefaultUser: "",
	},
	{
		Name:        "artisan",
		Short:       "Run Laravel Artisan commands",
		Framework:   "laravel",
		Binary:      "php",
		PrependArgs: []string{"artisan"},
		DefaultUser: "",
	},
	{
		Name:        "magerun",
		Short:       "Run n98-magerun commands",
		Framework:   "magento1",
		Binary:      "n98-magerun.phar",
		DefaultUser: "",
	},
	{
		Name:        "drush",
		Short:       "Run Drupal Drush commands",
		Framework:   "drupal",
		Binary:      "drush",
		DefaultUser: "",
	},
	{
		Name:        "symfony",
		Short:       "Run Symfony CLI commands",
		Framework:   "symfony",
		Binary:      "php",
		PrependArgs: []string{"bin/console"},
		DefaultUser: "",
	},
	{
		Name:        "shopware",
		Short:       "Run Shopware CLI commands",
		Framework:   "shopware",
		Binary:      "bin/console",
		DefaultUser: "",
	},
	{
		Name:        "cake",
		Short:       "Run CakePHP CLI commands",
		Framework:   "cakephp",
		Binary:      "bin/cake",
		DefaultUser: "",
	},
	{
		Name:        "composer",
		Short:       "Run composer commands",
		Binary:      "composer",
		DefaultUser: "",
	},
	{
		Name:        "wp",
		Short:       "Run WordPress CLI commands",
		Framework:   "wordpress",
		Binary:      "wp",
		DefaultUser: "",
	},
	{
		Name:        "npm",
		Short:       "Run npm commands",
		Binary:      "npm",
		DefaultUser: "",
	},
	{
		Name:        "yarn",
		Short:       "Run yarn commands",
		Binary:      "yarn",
		DefaultUser: "",
	},
	{
		Name:        "npx",
		Short:       "Run npx commands",
		Binary:      "npx",
		DefaultUser: "",
	},
	{
		Name:        "pnpm",
		Short:       "Run pnpm commands",
		Binary:      "pnpm",
		DefaultUser: "",
	},
	{
		Name:        "grunt",
		Short:       "Run grunt commands",
		Binary:      "grunt",
		DefaultUser: "",
	},
}

func initFrameworkCommands() {
	for _, fc := range frameworkCommands {
		usage := fmt.Sprintf("%s [args]", fc.Name)
		longDesc := fc.Short
		if fc.Framework != "" {
			longDesc = fmt.Sprintf("%s (Requires %s project)", fc.Short, fc.Framework)
		}
		cmd := &cobra.Command{
			Use:                usage,
			Short:              fc.Short,
			Long:               longDesc,
			DisableFlagParsing: true,
			RunE: func(c *cobra.Command, args []string) error {
				// Find which command we are running
				name := c.Name()
				var target FrameworkCommand
				for _, tc := range frameworkCommands {
					if tc.Name == name {
						target = tc
						break
					}
				}

				config := loadConfig()

				// Validate framework if required
				if target.Framework != "" && config.Framework != target.Framework {
					return fmt.Errorf("the '%s' command is only available for %s projects (current: %s)", name, target.Framework, config.Framework)
				}

				targetExec := resolveToolExecution(config, target.Binary, target.DefaultUser)
				return RunInContainerAt(targetExec.ContainerName, targetExec.User, targetExec.Workdir, target.Binary, append(target.PrependArgs, args...))
			},
		}
		toolCmd.AddCommand(cmd)
	}
	rootCmd.AddCommand(toolCmd)
}

func RunInContainer(containerName string, user string, binary string, args []string) error {
	return RunInContainerAt(containerName, user, "/var/www/html", binary, args)
}

func RunInContainerAt(containerName string, user string, workdir string, binary string, args []string) error {
	dockerArgs := []string{"exec"}
	if stdinIsTerminal() {
		dockerArgs = append(dockerArgs, "-it")
	} else {
		dockerArgs = append(dockerArgs, "-i")
	}
	if user != "" {
		dockerArgs = append(dockerArgs, "-u", user)
	}
	if strings.TrimSpace(workdir) == "" {
		workdir = "/var/www/html"
	}
	dockerArgs = append(dockerArgs, "-w", workdir, containerName, binary)
	dockerArgs = append(dockerArgs, args...)

	c := exec.Command("docker", dockerArgs...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}

func ResolveProjectExecUser(config engine.Config, fallback string) string {
	return config.ResolveProjectExecUser(fallback)
}

func resolveToolExecution(config engine.Config, binary string, defaultUser string) commandExecutionTarget {
	serviceName := engine.ResolveFrameworkAppService(config.Framework)
	workdir := engine.ResolveFrameworkAppWorkdir(config.Framework)
	user := defaultUser

	if engine.FrameworkUsesNodeRuntime(config.Framework) {
		return commandExecutionTarget{
			ContainerName: fmt.Sprintf("%s-%s-1", config.ProjectName, serviceName),
			Workdir:       workdir,
			User:          user,
		}
	}

	if config.Framework == "magento2" && (binary == "php" || binary == "composer" ||
		binary == "npm" || binary == "yarn" || binary == "npx" ||
		binary == "pnpm" || binary == "grunt") {
		user = config.ResolveProjectExecUser("www-data")
	} else if user == "" {
		user = config.ResolveProjectExecUser("www-data")
	}

	return commandExecutionTarget{
		ContainerName: fmt.Sprintf("%s-%s-1", config.ProjectName, serviceName),
		Workdir:       workdir,
		User:          user,
	}
}

func ResolveToolExecutionForTest(config engine.Config, binary string) (string, string, string) {
	target := resolveToolExecution(config, binary, "")
	return target.ContainerName, target.Workdir, target.User
}
