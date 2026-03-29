package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"
)

func upgradeMagento2(ctx context.Context, config Config, opts UpgradeOptions) error {
	containerName := fmt.Sprintf("%s-php-1", opts.ProjectName)

	if opts.TargetVersion == "" {
		return fmt.Errorf("target version is required. Example: govard upgrade --version=2.4.8-p4")
	}

	pterm.Info.Printf("Target version: %s\n", opts.TargetVersion)

	currentVersion, _ := getMagentoCurrentVersion(containerName)
	if currentVersion == "" {
		currentVersion = "unknown"
	}
	pterm.Info.Printf("Current version: %s\n", currentVersion)

	if opts.DryRun {
		pterm.Info.Println("[DRY RUN] Would perform the following steps:")
		pterm.Info.Println("  1. Update .govard.yml configuration for the target framework version")
		pterm.Info.Println("  2. Restart environment (govard env down && govard env up)")
		pterm.Info.Printf("  3. Create temporary magento project %s to fetch composer.json\n", opts.TargetVersion)
		pterm.Info.Println("  4. Merge composer.json preserving 3rd-party dependencies")
		pterm.Info.Println("  5. Run composer update magento/* phpunit/* --with-all-dependencies")
		pterm.Info.Println("  6. bin/magento setup:upgrade, setup:di:compile, cache:flush")
		return nil
	}

	if !opts.NoInteraction {
		pterm.Warning.Println("This will update framework profile dependencies, restart the environment, and modify composer.json.")
		fmt.Printf("Proceed with upgrade to %s? [y/N] ", opts.TargetVersion)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			pterm.Info.Println("Upgrade cancelled.")
			return nil
		}
	}

	// Step 1: Env update
	if !opts.NoEnvUpdate {
		pterm.Info.Println("Step 1/6: Applying runtime profile for target version...")
		targetProfile, err := ResolveRuntimeProfile("magento2", opts.TargetVersion)
		if err != nil {
			pterm.Warning.Printf("Could not resolve specific profile for %s (continuing): %v\n", opts.TargetVersion, err)
		} else {
			ApplyRuntimeProfileToConfig(&config, targetProfile.Profile)
			NormalizeConfig(&config, opts.ProjectDir)
			// Ensure it writes to file
			cleanConfig := PrepareConfigForWrite(config)
			yamlOut, err := yaml.Marshal(&cleanConfig)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			if err := os.WriteFile(filepath.Join(opts.ProjectDir, ".govard.yml"), yamlOut, 0644); err != nil {
				return fmt.Errorf("failed to write .govard.yml: %w", err)
			}
		}

		pterm.Info.Println("Step 2/6: Restarting environment (PHP, DB, Cache, Search)...")
		composePath := ComposeFilePath(opts.ProjectDir, opts.ProjectName)
		if err := RunCompose(ctx, ComposeOptions{
			ProjectDir:  opts.ProjectDir,
			ProjectName: opts.ProjectName,
			ComposeFile: composePath,
			Args:        []string{"down"},
			Stdout:      io.Discard,
			Stderr:      io.Discard,
		}); err != nil {
			pterm.Warning.Printf("Failed to stop environment: %v\n", err)
		}
		if err := RunCompose(ctx, ComposeOptions{
			ProjectDir:  opts.ProjectDir,
			ProjectName: opts.ProjectName,
			ComposeFile: composePath,
			Args:        []string{"up", "-d"},
			Stdout:      io.Discard,
			Stderr:      io.Discard,
		}); err != nil {
			return fmt.Errorf("failed to start environment: %w", err)
		}

		pterm.Info.Println("Waiting for database to be ready...")
		checkDatabaseReady(ctx, config, containerName)
	}

	pterm.Info.Println("Step 3/6: Fetching and merging composer.json...")

	if err := updateMagentoComposerJson(opts, containerName); err != nil {
		return fmt.Errorf("failed to update composer.json: %w", err)
	}

	pterm.Info.Println("Step 4/6: Running composer update...")
	// Relax some packages
	relaxPackages(containerName)

	// Composer update
	cmdArgs := []string{"exec", "-w", "/var/www/html", containerName, "composer", "update", "magento/*", "phpunit/*", "--with-all-dependencies"}
	updateCmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	updateCmd.Stdout = opts.Stdout
	updateCmd.Stderr = opts.Stderr
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("composer update failed: %w", err)
	}

	pterm.Info.Println("Step 5/6: Running setup:upgrade...")
	if !opts.NoDBUpgrade {
		suArgs := []string{"exec", "-w", "/var/www/html", containerName, "bin/magento", "setup:upgrade", "--no-interaction"}
		su := exec.CommandContext(ctx, "docker", suArgs...)
		su.Stdout = opts.Stdout
		su.Stderr = opts.Stderr
		if err := su.Run(); err != nil {
			return fmt.Errorf("setup:upgrade failed: %w", err)
		}
	} else {
		pterm.Info.Println("Skipped (--no-db-upgrade)")
	}

	pterm.Info.Println("Step 6/6: Compiling and flushing cache...")
	diArgs := []string{"exec", "-w", "/var/www/html", containerName, "bin/magento", "setup:di:compile"}
	diCmd := exec.CommandContext(ctx, "docker", diArgs...)
	if err := diCmd.Run(); err != nil {
		pterm.Warning.Printf("setup:di:compile failed: %v\n", err)
	}

	cacheArgs := []string{"exec", "-w", "/var/www/html", containerName, "bin/magento", "cache:flush"}
	cacheCmd := exec.CommandContext(ctx, "docker", cacheArgs...)
	if err := cacheCmd.Run(); err != nil {
		pterm.Warning.Printf("cache:flush failed: %v\n", err)
	}

	// Clean backup
	_ = os.Remove(filepath.Join(opts.ProjectDir, "composer.json.bak"))

	pterm.Success.Printf("✅ Magento upgrade to %s completed!\n", opts.TargetVersion)
	return nil
}

func getMagentoCurrentVersion(containerName string) (string, error) {
	cmdArgs := []string{"exec", "-w", "/var/www/html", containerName, "php", "bin/magento", "--version"}
	out, err := exec.Command("docker", cmdArgs...).CombinedOutput()
	if err == nil {
		re := regexp.MustCompile(`\d+\.\d+\.\d+(-p\d+)?`)
		match := re.FindString(string(out))
		if match != "" {
			return match, nil
		}
	}
	return "", fmt.Errorf("could not detect")
}

func updateMagentoComposerJson(opts UpgradeOptions, containerName string) error {
	composerPath := filepath.Join(opts.ProjectDir, "composer.json")
	backupPath := filepath.Join(opts.ProjectDir, "composer.json.bak")

	// Backup
	b, err := os.ReadFile(composerPath)
	if err == nil {
		if err := os.WriteFile(backupPath, b, 0644); err != nil {
			pterm.Warning.Printf("Failed to create backup: %v\n", err)
		}
	}

	projCmd := exec.Command("docker", "exec", "-w", "/var/www/html", containerName, "composer", "create-project", "--no-install", "--ignore-platform-reqs", "--repository=https://repo.magento.com/", "magento/project-community-edition", "temp_upgrade_source", opts.TargetVersion)
	projCmd.Stdout = opts.Stdout
	projCmd.Stderr = opts.Stderr
	if err := projCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch composer.json for %s", opts.TargetVersion)
	}

	newComposerBytes, err := exec.Command("docker", "exec", "-w", "/var/www/html", containerName, "cat", "temp_upgrade_source/composer.json").Output()
	_ = exec.Command("docker", "exec", "-w", "/var/www/html", containerName, "rm", "-rf", "temp_upgrade_source").Run()

	if err != nil {
		return fmt.Errorf("failed to read fetched composer.json")
	}

	var currentMap, newMap map[string]interface{}
	if err := json.Unmarshal(b, &currentMap); err != nil {
		return err
	}
	if err := json.Unmarshal(newComposerBytes, &newMap); err != nil {
		return err
	}

	// Merge require, require-dev, conflict, extra
	mergeComposerMapKeys(currentMap, newMap, "require")
	mergeComposerMapKeys(currentMap, newMap, "require-dev")
	mergeComposerMapKeys(currentMap, newMap, "conflict")
	mergeComposerMapKeys(currentMap, newMap, "autoload")
	mergeComposerMapKeys(currentMap, newMap, "minimum-stability")
	mergeComposerMapKeys(currentMap, newMap, "prefer-stable")
	mergeComposerMapKeys(currentMap, newMap, "extra")

	mergedBytes, err := json.MarshalIndent(currentMap, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(composerPath, mergedBytes, 0644)
}

func mergeComposerMapKeys(current map[string]interface{}, target map[string]interface{}, key string) {
	if _, ok := target[key]; !ok {
		return
	}
	targetVal := target[key]

	// If it's not a map but a direct value (like string for minimum-stability)
	targetMap, isMap := targetVal.(map[string]interface{})
	if !isMap {
		current[key] = targetVal
		return
	}

	if _, currOk := current[key]; !currOk {
		current[key] = map[string]interface{}{}
	}

	currentMap, currIsMap := current[key].(map[string]interface{})
	if !currIsMap {
		current[key] = targetVal
		return
	}

	for k, v := range targetMap {
		currentMap[k] = v
	}
}

func MergeComposerMapKeysForTest(current map[string]interface{}, target map[string]interface{}, key string) {
	mergeComposerMapKeys(current, target, key)
}

func relaxPackages(containerName string) {
	relax := []string{
		"phpunit/phpunit:*",
		"pdepend/pdepend:*",
		"phpmd/phpmd:*",
		"friendsofphp/php-cs-fixer:*",
		"magento/magento-coding-standard:*",
		"symfony/finder:*",
		"allure-framework/allure-phpunit:*",
		"sebastian/phpcpd:*",
		"laminas/laminas-dom:*",
	}

	// Check existing in composer.json
	cmdGet := exec.Command("docker", "exec", "-w", "/var/www/html", containerName, "cat", "composer.json")
	out, err := cmdGet.Output()
	if err != nil {
		return
	}
	content := string(out)

	var toRelax []string
	for _, pkgStr := range relax {
		pkgName := strings.Split(pkgStr, ":")[0]
		if strings.Contains(content, `"`+pkgName+`"`) {
			toRelax = append(toRelax, pkgStr)
		}
	}

	if len(toRelax) > 0 {
		args := append([]string{"exec", "-w", "/var/www/html", containerName, "composer", "require"}, toRelax...)
		args = append(args, "--no-update")
		if err := exec.Command("docker", args...).Run(); err != nil {
			pterm.Warning.Printf("Failed to relax packages: %v\n", err)
		}
	}
}

func checkDatabaseReady(ctx context.Context, config Config, containerName string) {
	for i := 0; i < 30; i++ {
		cmdArgs := []string{"exec", "-w", "/var/www/html", containerName, "php", "-r", "$m=new mysqli('db', 'magento', 'magento'); if($m->connect_error) exit(1); exit(0);"}
		out := exec.CommandContext(ctx, "docker", cmdArgs...)
		if err := out.Run(); err == nil {
			return
		}
		time.Sleep(2 * time.Second)
	}
	pterm.Warning.Println("Database may not be ready, continuing anyway...")
}
