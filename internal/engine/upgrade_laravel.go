package engine

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/pterm/pterm"
)

func upgradeLaravel(ctx context.Context, config Config, opts UpgradeOptions) error {
	pterm.Info.Println("Laravel Upgrade Pipeline")
	containerName := fmt.Sprintf("%s-php-1", opts.ProjectName)

	if opts.TargetVersion == "" {
		return fmt.Errorf("target version is required. Example: govard upgrade --version=11")
	}

	if opts.DryRun {
		pterm.Info.Println("[DRY RUN] Would perform the following steps:")
		pterm.Info.Printf("  1. composer require laravel/framework:^%s --no-update\n", opts.TargetVersion)
		pterm.Info.Println("  2. composer update")
		pterm.Info.Println("  3. php artisan migrate --force")
		return nil
	}

	// Step 1: Composer require
	pterm.Info.Println("Step 1/3: Updating composer.json...")
	requireCmd := exec.CommandContext(ctx, "docker", "exec", "-w", "/var/www/html", containerName, "composer", "require", fmt.Sprintf("laravel/framework:^%s", opts.TargetVersion), "--no-update")
	requireCmd.Stdout = opts.Stdout
	requireCmd.Stderr = opts.Stderr
	if err := requireCmd.Run(); err != nil {
		return fmt.Errorf("failed to update composer.json: %w", err)
	}

	// Step 2: Composer update
	pterm.Info.Println("Step 2/3: Running composer update...")
	updateCmd := exec.CommandContext(ctx, "docker", "exec", "-w", "/var/www/html", containerName, "composer", "update")
	updateCmd.Stdout = opts.Stdout
	updateCmd.Stderr = opts.Stderr
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("composer update failed: %w", err)
	}

	// Step 3: Migration
	pterm.Info.Println("Step 3/3: Running migrations...")
	migrateCmd := exec.CommandContext(ctx, "docker", "exec", "-w", "/var/www/html", containerName, "php", "artisan", "migrate", "--force")
	migrateCmd.Stdout = opts.Stdout
	migrateCmd.Stderr = opts.Stderr
	if err := migrateCmd.Run(); err != nil {
		pterm.Warning.Printf("Migrations failed: %v\n", err)
	}

	pterm.Success.Printf("✅ Laravel upgrade to ^%s completed!\n", opts.TargetVersion)
	return nil
}
