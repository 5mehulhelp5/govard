package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"govard/internal/conventions"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	targetVer := config.Stack.ComposerVersion
	if targetVer == "" || targetVer == "latest" {
		return nil // Image composer is suitable for PHP >= 7.2.5
	}

	pterm.Info.Printf("Ensuring Composer version %s as configured...\n", targetVer)
	return ensureSpecificComposerVersion(config, targetVer)
}

func ensureSpecificComposerVersion(config engine.Config, version string) error {
	pterm.Info.Printf("Ensuring Composer version %s is installed in container...\n", version)

	containerName := fmt.Sprintf("%s%s", config.ProjectName, conventions.PHPSuffix)

	// Dynamically resolve minor versions like "2.7" to the latest patch release (e.g. "2.7.9")
	if isMinorComposerVersion(version) && version != "2.2" {
		resolved, err := resolveComposerPatchVersion(version)
		if err == nil && resolved != "" {
			pterm.Info.Printf("Resolved Composer minor version %s to patch release %s\n", version, resolved)
			version = resolved
		} else {
			pterm.Warning.Printf("Could not dynamically resolve Composer version %s: %v\n", version, err)
		}
	}

	downloadUrl := "https://getcomposer.org/composer-stable.phar"
	if version != "latest" {
		// Try to resolve exactly or use the lts versions
		if version == "2" {
			downloadUrl = "https://getcomposer.org/composer-2.phar"
		} else if version == "1" {
			downloadUrl = "https://getcomposer.org/composer-1.phar"
		} else if version == "2.2" {
			downloadUrl = "https://getcomposer.org/download/latest-2.2.x/composer.phar"
		} else if strings.Contains(version, ".") {
			downloadUrl = fmt.Sprintf("https://getcomposer.org/download/%s/composer.phar", version)
		}
	}

	// Build a shell script that checks for pre-baked local binaries first
	script := fmt.Sprintf(`
		case "%[1]s" in
			1)   [ -f /usr/local/bin/composer1 ]   && echo "Using pre-baked Composer version 1..." && ln -sf /usr/local/bin/composer1 $(which composer) && exit 0 ;;
			2)   [ -f /usr/local/bin/composer2 ]   && echo "Using pre-baked Composer version 2..." && ln -sf /usr/local/bin/composer2 $(which composer) && exit 0 ;;
			2.2) [ -f /usr/local/bin/composer2lts ] && echo "Using pre-baked Composer version 2.2..." && ln -sf /usr/local/bin/composer2lts $(which composer) && exit 0 ;;
		esac
		echo "Ensuring Composer version %[1]s (downloading from %[2]s)..."
		curl -sSfL %[2]s -o /tmp/composer.phar && chmod +x /tmp/composer.phar && mv /tmp/composer.phar $(which composer)
	`, version, downloadUrl)

	cmd := exec.Command("docker", "exec", "-u", "root", containerName, "sh", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("composer setup failed (%s): %w: %s", version, err, string(out))
	}

	trimmedOut := strings.TrimSpace(string(out))
	if trimmedOut != "" {
		pterm.Info.Println(trimmedOut)
	}

	pterm.Success.Printf("Composer %s is now active.\n", version)

	// Fix: Composer 2.2+ blocks plugins by default. Enable them globally in the container for bootstrap.
	// We check if the directory is writable first and redirect stderr to silence noise when it's a read-only mount (common in govard).
	_ = runPHPContainerShellCommand(config, "[ -w ~/.composer ] && composer config -g allow-plugins true 2>/dev/null || true")
	return nil
}

func ensureBootstrapAuthJSON(config engine.Config, opts BootstrapRuntimeOptions) error {
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
	if err := os.MkdirAll(filepath.Dir(path), conventions.SecretDirPerm); err != nil {
		return fmt.Errorf("failed to ensure directory for auth.json: %w", err)
	}
	if err := os.WriteFile(path, []byte(payload), conventions.SecretFilePerm); err != nil {
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
	return ensureBootstrapAuthJSON(config, BootstrapRuntimeOptions{
		MageUsername: strings.TrimSpace(mageUsername),
		MagePassword: strings.TrimSpace(magePassword),
		AssumeYes:    assumeYes,
	})
}

func isMinorComposerVersion(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) == 2 {
		return true
	}
	return false
}

func resolveComposerPatchVersion(minorVersion string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://repo.packagist.org/p2/composer/composer.json", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Packages struct {
			Composer []struct {
				Version string `json:"version"`
			} `json:"composer/composer"`
		} `json:"packages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	for _, c := range result.Packages.Composer {
		// We want the most recent matching version that isn't a pre-release like RC/alpha/beta
		if strings.HasPrefix(c.Version, minorVersion+".") && !strings.Contains(strings.ToLower(c.Version), "-") {
			return c.Version, nil
		}
	}
	return "", fmt.Errorf("no stable patch versions found for %s", minorVersion)
}
