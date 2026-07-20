package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"govard/internal/engine"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

const xdebugLaunchConfigName = "Listen for Xdebug (Govard)"

var vscodeSetupGlobal bool

var vscodeSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Write or update VSCode settings to use this project's container instead of the host",
	Long: `Write (or merge into) the VSCode settings needed to run PHP tooling inside the
project's container instead of the host machine.

Without --global, run this from inside a project (or any subdirectory of one)
to write .vscode/settings.json (Intelephense PHP version, PHPStan path mapping)
and .vscode/launch.json (a "Listen for Xdebug" configuration).

With --global, updates VSCode's user settings.json once for every project:
creates the govard-php / govard-php-cs-fixer wrapper scripts under
~/.govard/bin and points php.validate.executablePath, phpstan.binCommand, and
php-cs-fixer.executablePath at them.

Existing keys in either file are preserved; only the keys this command manages
are added or overwritten. Note: settings.json is parsed as plain JSON, so any
comments in it will be dropped when rewritten.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if vscodeSetupGlobal {
			return runVSCodeSetupGlobal()
		}
		return runVSCodeSetupProject()
	},
}

func init() {
	vscodeSetupCmd.Flags().BoolVar(&vscodeSetupGlobal, "global", false, "Update VSCode's user settings.json instead of the current project")
}

func runVSCodeSetupProject() error {
	root, err := findProjectRootUpward()
	if err != nil {
		return err
	}
	if err := os.Chdir(root); err != nil {
		return fmt.Errorf("switch to project root %q: %w", root, err)
	}

	config := loadConfig()
	workdir := engine.ResolveFrameworkAppWorkdir(config.Framework)

	settingsPath := filepath.Join(root, ".vscode", "settings.json")
	set := map[string]interface{}{
		"phpstan.paths": map[string]interface{}{
			root: workdir,
		},
	}
	if version := normalizeIntelephensePHPVersion(config.Stack.PHPVersion); version != "" {
		set["intelephense.environment.phpVersion"] = version
	}
	if err := mergeJSONObjectFile(settingsPath, set, nil); err != nil {
		return fmt.Errorf("write %s: %w", settingsPath, err)
	}
	pterm.Success.Printf("Updated %s\n", settingsPath)

	launchPath := filepath.Join(root, ".vscode", "launch.json")
	if err := mergeLaunchConfig(launchPath, workdir); err != nil {
		return fmt.Errorf("write %s: %w", launchPath, err)
	}
	pterm.Success.Printf("Updated %s\n", launchPath)

	return nil
}

func runVSCodeSetupGlobal() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}

	binDir := filepath.Join(home, ".govard", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", binDir, err)
	}

	phpWrapper := filepath.Join(binDir, "govard-php")
	if err := ensureWrapperScript(phpWrapper, "php"); err != nil {
		return err
	}
	csFixerWrapper := filepath.Join(binDir, "govard-php-cs-fixer")
	if err := ensureWrapperScript(csFixerWrapper, "php-cs-fixer"); err != nil {
		return err
	}
	pterm.Success.Printf("Wrote %s and %s\n", phpWrapper, csFixerWrapper)

	settingsPath, err := vscodeGlobalSettingsPath()
	if err != nil {
		return err
	}

	set := map[string]interface{}{
		"php.validate.executablePath": phpWrapper,
		"phpstan.binCommand":          []string{"govard", "vscode", "phpstan"},
		"php-cs-fixer.executablePath": csFixerWrapper,
	}
	unset := []string{"phpstan.binPath"}
	if err := mergeJSONObjectFile(settingsPath, set, unset); err != nil {
		return fmt.Errorf("write %s: %w", settingsPath, err)
	}
	pterm.Success.Printf("Updated %s\n", settingsPath)

	return nil
}

// vscodeGlobalSettingsPath returns the path to VSCode's user settings.json,
// preferring an existing install (Code, Code - Insiders, VSCodium, Code -
// OSS) and falling back to the standard Code location.
func vscodeGlobalSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}

	var base string
	switch runtime.GOOS {
	case "darwin":
		base = filepath.Join(home, "Library", "Application Support")
	case "windows":
		base = os.Getenv("APPDATA")
		if base == "" {
			base = filepath.Join(home, "AppData", "Roaming")
		}
	default:
		base = filepath.Join(home, ".config")
	}

	candidates := []string{"Code", "Code - Insiders", "VSCodium", "Code - OSS"}
	for _, name := range candidates {
		path := filepath.Join(base, name, "User", "settings.json")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// None found yet (fresh install) — default to stable Code.
	return filepath.Join(base, candidates[0], "User", "settings.json"), nil
}

// normalizeIntelephensePHPVersion converts a "8.2"-style php_version into the
// "8.2.0"-style value intelephense.environment.phpVersion expects. Returns ""
// if no usable version is configured (e.g. php_version: none).
func normalizeIntelephensePHPVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" || version == "none" {
		return ""
	}
	if strings.Count(version, ".") == 1 {
		return version + ".0"
	}
	return version
}

// NormalizeIntelephensePHPVersionForTest exposes normalizeIntelephensePHPVersion to the tests package.
func NormalizeIntelephensePHPVersionForTest(version string) string {
	return normalizeIntelephensePHPVersion(version)
}

// MergeJSONObjectFileForTest exposes mergeJSONObjectFile to the tests package.
func MergeJSONObjectFileForTest(path string, set map[string]interface{}, unset []string) error {
	return mergeJSONObjectFile(path, set, unset)
}

// MergeLaunchConfigForTest exposes mergeLaunchConfig to the tests package.
func MergeLaunchConfigForTest(path, workdir string) error {
	return mergeLaunchConfig(path, workdir)
}

// ensureWrapperScript writes an executable one-line wrapper that delegates to
// `govard vscode <subcommand>`, used for editor settings that require a
// single executable path rather than a command array.
func ensureWrapperScript(path, subcommand string) error {
	body := fmt.Sprintf("#!/usr/bin/env bash\nexec govard vscode %s \"$@\"\n", subcommand)
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// mergeJSONObjectFile reads path as a JSON object (treating a missing file as
// empty), applies each entry in set, removes each key in unset, and writes
// the result back with indentation. Keys are re-sorted alphabetically by
// encoding/json; unrelated existing keys are preserved.
func mergeJSONObjectFile(path string, set map[string]interface{}, unset []string) error {
	obj := map[string]interface{}{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &obj); err != nil {
			return fmt.Errorf("parse existing %s (comments/trailing commas aren't supported): %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}

	for _, key := range unset {
		delete(obj, key)
	}
	for key, value := range set {
		obj[key] = value
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}

	data, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// mergeLaunchConfig ensures launch.json at path contains a "Listen for
// Xdebug (Govard)" configuration with the given container workdir mapped to
// ${workspaceFolder}. Any other configurations are left untouched.
func mergeLaunchConfig(path, workdir string) error {
	launch := struct {
		Version        string                   `json:"version"`
		Configurations []map[string]interface{} `json:"configurations"`
	}{Version: "0.2.0"}

	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &launch); err != nil {
			return fmt.Errorf("parse existing %s (comments/trailing commas aren't supported): %w", path, err)
		}
		if launch.Version == "" {
			launch.Version = "0.2.0"
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}

	xdebugConfig := map[string]interface{}{
		"name":    xdebugLaunchConfigName,
		"type":    "php",
		"request": "launch",
		"port":    9003,
		"pathMappings": map[string]interface{}{
			workdir: "${workspaceFolder}",
		},
	}

	replaced := false
	for i, c := range launch.Configurations {
		if name, _ := c["name"].(string); name == xdebugLaunchConfigName {
			launch.Configurations[i] = xdebugConfig
			replaced = true
			break
		}
	}
	if !replaced {
		launch.Configurations = append(launch.Configurations, xdebugConfig)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}

	data, err := json.MarshalIndent(launch, "", "    ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
