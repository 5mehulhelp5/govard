package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"govard/internal/engine"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	profileFrameworkOverride string
	profileVersionOverride   string
	profileJSONOutput        bool
	profileSkipTuningPrompt  bool // Prevents re-entry during nested up calls
)

type profileDetectedPayload struct {
	Framework string `json:"framework"`
	Version   string `json:"version"`
}

type profileSelectedPayload struct {
	Framework        string `json:"framework"`
	FrameworkVersion string `json:"framework_version"`
	PHPVersion       string `json:"php_version"`
	NodeVersion      string `json:"node_version"`
	DB               string `json:"db"`
	DBVersion        string `json:"db_version"`
	WebRoot          string `json:"web_root"`
	WebServer        string `json:"web_server"`
	Cache            string `json:"cache"`
	CacheVersion     string `json:"cache_version"`
	Search           string `json:"search"`
	SearchVersion    string `json:"search_version"`
	Queue            string `json:"queue"`
	QueueVersion     string `json:"queue_version"`
	XdebugSession    string `json:"xdebug_session"`
}

type profileOutputPayload struct {
	Detected profileDetectedPayload `json:"detected"`
	Selected profileSelectedPayload `json:"selected"`
	Source   string                 `json:"source"`
	Notes    []string               `json:"notes"`
	Warnings []string               `json:"warnings"`
}

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage environment profiles (show, switch, apply, clear)",
	RunE: func(cmd *cobra.Command, args []string) error {
		metadata, result, err := resolveProfileForCurrentProject()
		if err != nil {
			return err
		}

		if profileJSONOutput {
			payload := buildProfileOutputPayload(metadata, result)
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(payload)
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Detected framework: %s\n", metadata.Framework)
		if metadata.Version != "" {
			fmt.Fprintf(out, "Detected framework version: %s\n", metadata.Version)
		} else {
			fmt.Fprintln(out, "Detected framework version: (not detected)")
		}
		fmt.Fprintf(out, "Profile source: %s\n", result.Source)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Recommended Govard profile:")
		fmt.Fprintf(out, "  framework: %s\n", result.Profile.Framework)
		if result.Profile.FrameworkVersion != "" {
			fmt.Fprintf(out, "  framework_version: %s\n", result.Profile.FrameworkVersion)
		}
		fmt.Fprintf(out, "  stack.php_version: %s\n", result.Profile.PHPVersion)
		if result.Profile.NodeVersion != "" {
			fmt.Fprintf(out, "  stack.node_version: %s\n", result.Profile.NodeVersion)
		}
		fmt.Fprintf(out, "  stack.services.db: %s\n", result.Profile.DB)
		if result.Profile.DBVersion != "" {
			fmt.Fprintf(out, "  stack.db_version: %s\n", result.Profile.DBVersion)
		}
		if result.Profile.WebRoot != "" {
			fmt.Fprintf(out, "  stack.web_root: %s\n", result.Profile.WebRoot)
		}
		fmt.Fprintf(out, "  stack.services.web_server: %s\n", result.Profile.WebServer)
		fmt.Fprintf(out, "  stack.services.cache: %s\n", result.Profile.Cache)
		if result.Profile.CacheVersion != "" {
			fmt.Fprintf(out, "  stack.cache_version: %s\n", result.Profile.CacheVersion)
		}
		fmt.Fprintf(out, "  stack.services.search: %s\n", result.Profile.Search)
		if result.Profile.SearchVersion != "" {
			fmt.Fprintf(out, "  stack.search_version: %s\n", result.Profile.SearchVersion)
		}
		fmt.Fprintf(out, "  stack.services.queue: %s\n", result.Profile.Queue)
		if result.Profile.QueueVersion != "" {
			fmt.Fprintf(out, "  stack.queue_version: %s\n", result.Profile.QueueVersion)
		}

		if len(result.Notes) > 0 {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Notes:")
			for _, note := range result.Notes {
				fmt.Fprintf(out, "  - %s\n", note)
			}
		}
		if len(result.Warnings) > 0 {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Warnings:")
			for _, warning := range result.Warnings {
				fmt.Fprintf(out, "  - %s\n", warning)
			}
		}

		return nil
	},
}

var profileApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the recommended runtime profile to .govard.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, result, err := resolveProfileForCurrentProject()
		if err != nil {
			return err
		}

		wd, _ := os.Getwd()
		config, err := engine.LoadBaseConfigFromDir(wd, false)
		if err != nil {
			return err
		}

		engine.ApplyRuntimeProfileToConfig(&config, result.Profile)
		engine.NormalizeConfig(&config, wd)
		saveConfig(config)
		pterm.Success.Println("Applied profile to .govard.yml")
		return nil
	},
}

var profileSwitchCmd = &cobra.Command{
	Use:   "switch [profile_name]",
	Short: "Switch to a different environment profile",
	Long: `Switches the active environment profile for the current project.
The profile name is persisted per-project in ~/.govard/projects.json.

When called without an argument, shows an interactive profile selector.
When called with a profile name, switches directly.

Use "govard config profile clear" to reset to default profile.`,
	Example: `  # Interactive profile selection
  govard config profile switch

  # Switch directly to a profile
  govard config profile switch upgrade`,
	RunE: runProfileSwitch,
}

var profileClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Reset to default profile (clears saved profile)",
	Long:  `Clears the saved profile, reverting to the default (no profile) behavior.`,
	RunE:  runProfileClear,
}

func init() {
	registerProfileFlags(profileCmd)
	profileCmd.AddCommand(profileApplyCmd)
	profileCmd.AddCommand(profileSwitchCmd)
	profileCmd.AddCommand(profileClearCmd)
	configCmd.AddCommand(profileCmd)
}

func registerProfileFlags(command *cobra.Command) {
	command.PersistentFlags().StringVar(&profileFrameworkOverride, "framework", "", "Override detected framework")
	command.PersistentFlags().StringVar(&profileVersionOverride, "framework-version", "", "Override detected framework version")
	command.PersistentFlags().BoolVar(&profileJSONOutput, "json", false, "Output selected profile as JSON")
}

func resolveProfileForCurrentProject() (engine.ProjectMetadata, engine.RuntimeProfileResult, error) {
	cwd, _ := os.Getwd()
	metadata := engine.DetectFramework(cwd)

	if v := strings.TrimSpace(profileFrameworkOverride); v != "" {
		metadata.Framework = strings.ToLower(v)
		if metadata.Framework == "magento" {
			metadata.Framework = "magento2"
		}
	}
	if v := strings.TrimSpace(profileVersionOverride); v != "" {
		metadata.Version = v
	}

	if metadata.Framework == "" || metadata.Framework == "generic" {
		return metadata, engine.RuntimeProfileResult{}, fmt.Errorf("could not detect framework, please use --framework")
	}

	result, err := engine.ResolveRuntimeProfile(metadata.Framework, metadata.Version)
	if err != nil {
		return metadata, engine.RuntimeProfileResult{}, err
	}

	return metadata, result, nil
}

func buildProfileOutputPayload(metadata engine.ProjectMetadata, result engine.RuntimeProfileResult) profileOutputPayload {
	notes := result.Notes
	if notes == nil {
		notes = []string{}
	}
	warnings := result.Warnings
	if warnings == nil {
		warnings = []string{}
	}

	return profileOutputPayload{
		Detected: profileDetectedPayload{
			Framework: metadata.Framework,
			Version:   metadata.Version,
		},
		Selected: profileSelectedPayload{
			Framework:        result.Profile.Framework,
			FrameworkVersion: result.Profile.FrameworkVersion,
			PHPVersion:       result.Profile.PHPVersion,
			NodeVersion:      result.Profile.NodeVersion,
			DB:               result.Profile.DB,
			DBVersion:        result.Profile.DBVersion,
			WebRoot:          result.Profile.WebRoot,
			WebServer:        result.Profile.WebServer,
			Cache:            result.Profile.Cache,
			CacheVersion:     result.Profile.CacheVersion,
			Search:           result.Profile.Search,
			SearchVersion:    result.Profile.SearchVersion,
			Queue:            result.Profile.Queue,
			QueueVersion:     result.Profile.QueueVersion,
			XdebugSession:    result.Profile.XdebugSession,
		},
		Source:   result.Source,
		Notes:    notes,
		Warnings: warnings,
	}
}

// detectAvailableProfiles finds all .govard.<name>.yml files in the project root.
func detectAvailableProfiles(projectRoot string) []string {
	entries, err := os.ReadDir(projectRoot)
	if err != nil {
		return nil
	}

	var profiles []string
	prefix := ".govard."
	suffix := ".yml"
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix) && len(name) > len(prefix+suffix) {
			profile := strings.TrimSuffix(strings.TrimPrefix(name, prefix), suffix)
			// Skip special files
			if profile == "local" || profile == "project.local" || profile == "compose" {
				continue
			}
			profiles = append(profiles, profile)
		}
	}
	return profiles
}

// runProfileSwitch handles the profile switch command.
func runProfileSwitch(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()

	// 1. Determine profile name
	var profileName string
	if len(args) > 0 {
		profileName = args[0]
	} else {
		// Interactive mode
		profiles := detectAvailableProfiles(cwd)
		if len(profiles) == 0 {
			pterm.Warning.Println("No profile files (.govard.<name>.yml) found in this project.")
			return nil
		}

		selected, err := pterm.DefaultInteractiveSelect.
			WithOptions(profiles).
			Show("Select a profile to switch to")
		if err != nil || selected == "" {
			return fmt.Errorf("profile selection cancelled")
		}
		profileName = selected
	}

	// Validate profile exists (or is empty for default)
	if profileName != "" {
		profilePath := fmt.Sprintf("%s/.govard.%s.yml", cwd, profileName)
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			return fmt.Errorf("profile %q not found (expected file: %s)", profileName, profilePath)
		}
	}

	// 2. Update project registry with new profile
	// Empty profileName means "use default" - clear the saved profile
	entry := engine.ProjectRegistryEntry{
		Path:    cwd,
		Profile: profileName,
	}
	if err := engine.UpsertProjectRegistryEntry(entry); err != nil {
		return fmt.Errorf("save profile to registry: %w", err)
	}

	// 3. Update .govard.lock with new profile (preserve other fields)
	lockPath := engine.LockFilePath(cwd)
	if existingLock, err := engine.ReadLockFile(lockPath); err == nil {
		existingLock.Project.Profile = profileName
		if err := engine.WriteLockFile(lockPath, existingLock); err != nil {
			pterm.Warning.Printf("Could not update .govard.lock: %v\n", err)
		}
	}

	// 4. Optionally run profile tuning
	tuningTriggered := askProfileTuning(cwd, profileName, cmd.OutOrStdout(), cmd.ErrOrStderr())
	if tuningTriggered {
		return nil // Success message printed by askProfileTuning
	}

	// Print success if tuning wasn't triggered
	pterm.Success.Printf("Switched to profile %q\n", profileName)
	return nil
}

// askProfileTuning prompts user to run framework profile tuning if needed.
// Returns true if environment was started, false otherwise.
func askProfileTuning(cwd, profileName string, out, errOut io.Writer) bool {
	// Skip if already in a nested call (e.g., from up command)
	if profileSkipTuningPrompt {
		return false
	}

	config, _, err := engine.LoadConfigFromDirWithProfile(cwd, false, "")
	if err != nil || config.Framework == "" || config.Framework == "generic" {
		return false
	}

	metadata := engine.DetectFramework(cwd)
	version := strings.TrimSpace(metadata.Version)
	if version == "" {
		version = strings.TrimSpace(config.FrameworkVersion)
	}

	profileResult, err := engine.ResolveRuntimeProfile(config.Framework, version)
	if err != nil {
		return false
	}

	// Check if tuning is needed
	normalized := config
	engine.NormalizeConfig(&normalized, cwd)
	needsTuning := (profileResult.Profile.PHPVersion != "" && normalized.Stack.PHPVersion != profileResult.Profile.PHPVersion) ||
		(profileResult.Profile.DBVersion != "" && normalized.Stack.DBVersion != profileResult.Profile.DBVersion)

	if needsTuning {
		fmt.Println()
		pterm.Warning.Println("Detected profile mismatch with framework recommendations.")
		fmt.Println("Profile tuning will run automatically when the environment starts.")
		proceed, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultValue(false).
			WithDefaultText("Do you want to start the environment now?").
			Show("Start environment with the new profile?")
		if proceed {
			// Set flag to prevent re-entry during up command
			profileSkipTuningPrompt = true
			// Skip up command's success message since we print our own
			upCmdSkipSuccessMessage = true

			// Run env up to start the environment (tuning happens via ConfigureMagento)
			// Use subprocess to get full output like running "govard env up"
			executablePath, _ := os.Executable()
			command := exec.Command(executablePath, "env", "up")
			command.Dir = cwd
			command.Stdin = os.Stdin
			command.Stdout = out
			command.Stderr = errOut
			if err := command.Run(); err != nil {
				fmt.Fprintf(errOut, "Warning: env up returned error: %v\n", err)
			}

			// Reset flags after up completes
			profileSkipTuningPrompt = false
			upCmdSkipSuccessMessage = false

			return true
		}
	}
	return false
}

func runProfileClear(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()

	// Clear profile from project registry
	entry := engine.ProjectRegistryEntry{
		Path:    cwd,
		Profile: "",
	}
	if err := engine.UpsertProjectRegistryEntry(entry); err != nil {
		return fmt.Errorf("clear profile from registry: %w", err)
	}

	// Clear profile from lock file
	lockPath := engine.LockFilePath(cwd)
	if existingLock, err := engine.ReadLockFile(lockPath); err == nil {
		existingLock.Project.Profile = ""
		if err := engine.WriteLockFile(lockPath, existingLock); err != nil {
			pterm.Warning.Printf("Could not update .govard.lock: %v\n", err)
		}
	}

	pterm.Success.Println("Profile reset to default")
	return nil
}
