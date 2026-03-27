package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"govard/internal/engine"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func bootstrapComposerDumpAutoload(cmd *cobra.Command, cwd string) error {
	if !fileExists(filepath.Join(cwd, "composer.json")) {
		return nil
	}
	if err := runGovardSubcommand(cmd, govardComposerSubcommandArgs("dump-autoload", "-n")...); err != nil {
		autoloadPath := filepath.Join(cwd, "vendor", "autoload.php")
		if !fileExists(autoloadPath) {
			return fmt.Errorf("composer autoload generation failed: %w", err)
		}
		pterm.Warning.Printf("composer dump-autoload skipped (%v).\n", err)
	}
	return nil
}

func runBootstrapComposerPrepare(config engine.Config) error {
	if err := runPHPContainerShellCommand(config, "rm -rf vendor"); err != nil {
		return fmt.Errorf("failed to clean vendor directory: %w", err)
	}
	return nil
}

func FixComposerCompatibility(config engine.Config) error {
	version := config.Stack.PHPVersion
	targetVer := config.Stack.ComposerVersion
	if targetVer == "" {
		targetVer = "latest"
	}

	// If user explicitly wants a version, we should try to ensure it
	if targetVer != "latest" && targetVer != "" {
		pterm.Info.Printf("Ensuring Composer version %s as requested in config...\n", targetVer)
		return ensureSpecificComposerVersion(config, targetVer)
	}

	// Automatic compatibility check for old PHP
	if engine.IsNumericDotVersionAtLeast(version, "7.2.5") {
		return nil
	}

	pterm.Info.Printf("PHP version is %s (< 7.2.5). Ensuring Composer 2.2 LTS compatibility...\n", version)

	// Check if composer current runs or fails with support error code
	err := runPHPContainerShellCommand(config, "composer --version")
	if err == nil {
		return nil // Already works
	}

	return ensureSpecificComposerVersion(config, "2.2.24")
}

func ensureSpecificComposerVersion(config engine.Config, version string) error {
	pterm.Info.Printf("Ensuring Composer version %s is installed in container...\n", version)

	containerName := fmt.Sprintf("%s-php-1", config.ProjectName)

	downloadUrl := "https://getcomposer.org/composer-stable.phar"
	if version != "latest" {
		// Try to resolve exactly or use the lts versions
		if version == "2" {
			downloadUrl = "https://getcomposer.org/composer-2.phar"
		} else if version == "1" {
			downloadUrl = "https://getcomposer.org/composer-1.phar"
		} else if version == "2.2" {
			downloadUrl = "https://getcomposer.org/download/2.2.24/composer.phar"
		} else if strings.Contains(version, ".") {
			downloadUrl = fmt.Sprintf("https://getcomposer.org/download/%s/composer.phar", version)
		}
	}

	script := fmt.Sprintf("curl -sS %s -o /tmp/composer.phar && chmod +x /tmp/composer.phar && mv /tmp/composer.phar $(which composer)", downloadUrl)
	cmd := exec.Command("docker", "exec", "-u", "root", containerName, "sh", "-c", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("composer setup failed (%s): %w: %s", version, err, string(out))
	}

	pterm.Success.Printf("Composer %s is now active.\n", version)

	// Fix: Composer 2.2+ blocks plugins by default. Enable them globally in the container for bootstrap.
	_ = runPHPContainerShellCommand(config, "composer config -g allow-plugins true")
	return nil
}

func ensureBootstrapAuthJSON(config engine.Config, opts bootstrapRuntimeOptions) error {
	cwd, _ := os.Getwd()
	authPath := filepath.Join(cwd, "auth.json")
	if _, err := os.Stat(authPath); err == nil {
		return nil
	}

	globalAuthPath := filepath.Join(os.Getenv("HOME"), ".composer", "auth.json")
	if _, err := os.Stat(globalAuthPath); err == nil {
		useGlobal := opts.AssumeYes || shouldUseGlobalAuthByDefault()
		if !useGlobal {
			useGlobal, _ = pterm.DefaultInteractiveConfirm.
				WithDefaultValue(true).
				Show(fmt.Sprintf("Found global auth.json at %s. Use it for this project?", globalAuthPath))
		}

		if useGlobal {
			// Instead of copying, we just acknowledge it's there.
			// If we need to write project-specific logic later, we can,
			// but for now we rely on the mount handled in Render.
			pterm.Success.Printf("Using global auth.json from %s\n", globalAuthPath)
			return nil
		}
	}

	if opts.MageUsername != "" && opts.MagePassword != "" {
		return createAuthJSONFromCredentials(globalAuthPath, opts.MageUsername, opts.MagePassword, cwd)
	}

	if config.Framework == "magento2" && !shouldUseGlobalAuthByDefault() && !opts.AssumeYes {
		pterm.Info.Println("Magento 2 requires authentication for repo.magento.com.")
		pterm.Info.Println("You can find your keys at: https://marketplace.magento.com/customer/accessKeys/")

		username, _ := pterm.DefaultInteractiveTextInput.Show("Magento Public Key")
		password, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Magento Private Key")

		if username != "" && password != "" {
			return createAuthJSONFromCredentials(globalAuthPath, username, password, cwd)
		}
	}

	pterm.Warning.Println("auth.json not found. Provide --mage-username/--mage-password or create auth.json before composer-related steps.")
	return nil
}

func createAuthJSONFromCredentials(path, username, password, cwd string) error {
	payload := fmt.Sprintf("{\n    \"http-basic\": {\n        \"repo.magento.com\": {\n            \"username\": %q,\n            \"password\": %q\n        }\n    }\n}\n", username, password)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to ensure directory for auth.json: %w", err)
	}
	if err := os.WriteFile(path, []byte(payload), 0600); err != nil {
		return fmt.Errorf("failed writing auth.json to %s: %w", path, err)
	}
	pterm.Success.Printf("✅ Created auth.json in %s with provided credentials.\n", path)
	return nil
}

func shouldUseGlobalAuthByDefault() bool {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return true
	}
	return (stdinInfo.Mode() & os.ModeCharDevice) == 0
}

// BootstrapComposerDumpAutoloadForTest exposes bootstrapComposerDumpAutoload for tests in /tests.
func BootstrapComposerDumpAutoloadForTest(cmd *cobra.Command, cwd string) error {
	return bootstrapComposerDumpAutoload(cmd, cwd)
}

// RunBootstrapComposerPrepareForTest exposes runBootstrapComposerPrepare for tests in /tests.
func RunBootstrapComposerPrepareForTest(config engine.Config) error {
	return runBootstrapComposerPrepare(config)
}

// EnsureBootstrapAuthJSONForTest exposes ensureBootstrapAuthJSON for tests in /tests.
func EnsureBootstrapAuthJSONForTest(config engine.Config, mageUsername, magePassword string, assumeYes bool) error {
	return ensureBootstrapAuthJSON(config, bootstrapRuntimeOptions{
		MageUsername: strings.TrimSpace(mageUsername),
		MagePassword: strings.TrimSpace(magePassword),
		AssumeYes:    assumeYes,
	})
}
