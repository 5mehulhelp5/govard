package cmd

import (
	"fmt"
	"govard/internal/engine"
	"os"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func loadConfig() engine.Config {
	wd, _ := os.Getwd()
	config, _, err := engine.LoadConfigFromDir(wd, false)
	if err != nil {
		pterm.Warning.Printf("Failed to load layered config: %v\n", err)
		return engine.Config{}
	}
	return config
}

func loadFullConfig() (engine.Config, error) {
	return loadFullConfigWithProfile("")
}

func loadFullConfigWithProfile(profile string) (engine.Config, error) {
	wd, _ := os.Getwd()
	config, _, err := engine.LoadConfigFromDirWithProfile(wd, true, profile)
	if err != nil {
		return engine.Config{}, fmt.Errorf("could not load config: %w", err)
	}
	return config, nil
}

func loadWritableConfig() (engine.Config, error) {
	wd, _ := os.Getwd()
	config, err := engine.LoadBaseConfigFromDir(wd, true)
	if err != nil {
		return engine.Config{}, fmt.Errorf("could not load %s: %w", engine.BaseConfigFile, err)
	}
	return config, nil
}

func saveConfig(config engine.Config) {
	if err := engine.ValidateConfig(config); err != nil {
		pterm.Error.Printf("Config validation failed: %v\n", err)
		return
	}
	writableConfig := engine.PrepareConfigForWrite(config)

	data, err := yaml.Marshal(&writableConfig)
	if err != nil {
		pterm.Error.Printf("Failed to marshal config: %v\n", err)
		return
	}
	err = os.WriteFile(engine.BaseConfigFile, data, 0644)
	if err != nil {
		pterm.Error.Printf("Failed to write %s: %v\n", engine.BaseConfigFile, err)
	}
}

func runUp() {
	// Call up command logic
	_ = upCmd.RunE(upCmd, []string{})
}

var govardSubcommandRunner = func(cmd *cobra.Command, args ...string) error {
	executablePath, err := os.Executable()
	commandPath := "govard"
	if err == nil && strings.TrimSpace(executablePath) != "" {
		commandPath = executablePath
	}

	command := exec.Command(commandPath, args...)
	command.Dir, _ = os.Getwd()
	command.Stdin = os.Stdin
	command.Stdout = cmd.OutOrStdout()
	command.Stderr = cmd.ErrOrStderr()
	return command.Run()
}

func runGovardSubcommand(cmd *cobra.Command, args ...string) error {
	return govardSubcommandRunner(cmd, args...)
}

// rebrandComposeHelp runs `docker compose --help`, rebrands the output to use govard command names,
// and prints it to the command's stdout.
func rebrandComposeHelp(cmd *cobra.Command, govardCmdName string) {
	// First, print our own Govard-specific help header if available
	printGovardHelpHeader(cmd)

	// Determine the subcommand by looking at os.Args
	dockerArgs := []string{"compose"}
	for i, arg := range os.Args {
		if arg == govardCmdName {
			// Append subsequent args to get subcommand-specific help
			for j := i + 1; j < len(os.Args); j++ {
				candidate := os.Args[j]
				if candidate != "--help" && candidate != "-h" && candidate != "help" && !strings.HasPrefix(candidate, "-") {
					dockerArgs = append(dockerArgs, candidate)
				}
			}
			break
		}
	}
	dockerArgs = append(dockerArgs, "--help")

	c := exec.Command("docker", dockerArgs...)
	out, err := c.CombinedOutput()
	if err != nil {
		pterm.Error.Printf("Failed to get Docker Compose help: %v\n", err)
		return
	}

	helpText := string(out)

	// Rebrand: replace `docker compose` with `govard [govardCmdName]`
	helpText = strings.ReplaceAll(helpText, "docker compose", "govard "+govardCmdName)
	
	// Remove noise (flags that user shouldn't use directly because Govard manages them)
	noiseLines := []string{
		"--file",
		"-f",
		"--project-name",
		"-p",
		"--project-directory",
	}
	
	lines := strings.Split(helpText, "\n")
	var filteredLines []string
	for _, line := range lines {
		skip := false
		for _, noise := range noiseLines {
			if strings.Contains(line, noise) {
				skip = true
				break
			}
		}
		if !skip {
			filteredLines = append(filteredLines, line)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), strings.Join(filteredLines, "\n"))
}

func printGovardHelpHeader(cmd *cobra.Command) {
	if cmd.Long != "" {
		fmt.Fprintln(cmd.OutOrStdout(), cmd.Long)
		fmt.Fprintln(cmd.OutOrStdout())
	} else if cmd.Short != "" {
		fmt.Fprintln(cmd.OutOrStdout(), cmd.Short)
		fmt.Fprintln(cmd.OutOrStdout())
	}

	if cmd.Example != "" {
		pterm.DefaultSection.WithLevel(2).Println("Examples:")
		fmt.Fprintln(cmd.OutOrStdout(), cmd.Example)
		fmt.Fprintln(cmd.OutOrStdout())
	}
}

func boolFlagOrDefault(cmd *cobra.Command, name string, fallback bool) bool {
	if cmd == nil {
		return fallback
	}
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return fallback
	}
	value, err := cmd.Flags().GetBool(name)
	if err != nil {
		return fallback
	}
	return value
}

// isComposeMaintenanceCommand returns true if the command is a common Docker Compose maintenance command
// that accepts a service name at the end of its arguments.
func isComposeMaintenanceCommand(cmd string) bool {
	commands := map[string]bool{
		"ps":       true,
		"logs":     true,
		"top":      true,
		"stop":     true,
		"start":    true,
		"restart":  true,
		"pause":    true,
		"unpause":  true,
		"pull":     true,
		"build":    true,
		"port":     true,
		"images":   true,
		"rm":       true,
		"kill":     true,
	}
	return commands[cmd]
}

// proxyServiceToCompose forwards a command to Docker Compose for a specific service.
func proxyServiceToCompose(cmd *cobra.Command, service string, args []string) error {
	config := loadConfig()
	cwd, _ := os.Getwd()
	composePath := engine.ComposeFilePath(cwd, config.ProjectName)

	subcommand := args[0]
	remainingArgs := args[1:]

	composeArgs := append([]string{subcommand}, remainingArgs...)
	composeArgs = append(composeArgs, service)

	return engine.RunCompose(cmd.Context(), engine.ComposeOptions{
		ProjectDir:  cwd,
		ProjectName: config.ProjectName,
		ComposeFile: composePath,
		Args:        composeArgs,
		Stdout:      cmd.OutOrStdout(),
		Stderr:      cmd.ErrOrStderr(),
		Stdin:       os.Stdin,
	})
}
