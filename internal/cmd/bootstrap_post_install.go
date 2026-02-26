package cmd

import (
	"fmt"
	"os"
	"strings"

	"govard/internal/engine"
	"govard/internal/engine/remote"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func runBootstrapHyvaInstall(cmd *cobra.Command, opts bootstrapRuntimeOptions) error {
	if err := runGovardSubcommand(
		cmd,
		govardComposerSubcommandArgs(
			"config",
			"http-basic.hyva-themes.repo.packagist.com",
			"token",
			opts.HyvaToken,
		)...,
	); err != nil {
		return fmt.Errorf("failed to set Hyva token: %w", err)
	}
	if err := runGovardSubcommand(
		cmd,
		govardComposerSubcommandArgs(
			"config",
			"repositories.hyva-themes",
			"composer",
			"https://hyva-themes.repo.packagist.com/app-hyva-test-dv1dgx/",
		)...,
	); err != nil {
		return fmt.Errorf("failed to add Hyva repository: %w", err)
	}
	if err := runGovardSubcommand(cmd, govardComposerSubcommandArgs("require", "-n", "hyva-themes/magento2-default-theme")...); err != nil {
		return fmt.Errorf("failed to install Hyva package: %w", err)
	}
	return nil
}

func runBootstrapMagentoSetupInstall(cmd *cobra.Command, config engine.Config, opts bootstrapRuntimeOptions) error {
	emailDomain := config.Domain
	if emailDomain == "" {
		emailDomain = "local.test"
	}

	setupArgs := []string{
		"setup:install",
		"--backend-frontname=admin",
		"--db-host=db",
		"--db-name=magento",
		"--db-user=magento",
		"--db-password=magento",
		"--db-prefix=" + strings.TrimSpace(os.Getenv("DB_PREFIX")),
		"--search-engine=opensearch",
		"--opensearch-host=elasticsearch",
		"--opensearch-port=9200",
		"--opensearch-index-prefix=magento2",
		"--opensearch-enable-auth=0",
		"--opensearch-timeout=15",
		"--admin-user=admin",
		"--admin-password=Admin123$",
		"--admin-firstname=Admin",
		"--admin-lastname=User",
		"--admin-email=admin@" + emailDomain,
	}

	if opts.MetaVersion != "" {
		if comparison, comparable := compareNumericDotVersions(opts.MetaVersion, "2.4.8"); comparable && comparison < 0 {
			setupArgs = []string{
				"setup:install",
				"--backend-frontname=admin",
				"--db-host=db",
				"--db-name=magento",
				"--db-user=magento",
				"--db-password=magento",
				"--db-prefix=" + strings.TrimSpace(os.Getenv("DB_PREFIX")),
				"--search-engine=elasticsearch7",
				"--elasticsearch-host=elasticsearch",
				"--elasticsearch-port=9200",
				"--elasticsearch-index-prefix=magento2",
				"--elasticsearch-enable-auth=0",
				"--elasticsearch-timeout=15",
				"--admin-user=admin",
				"--admin-password=Admin123$",
				"--admin-firstname=Admin",
				"--admin-lastname=User",
				"--admin-email=admin@" + emailDomain,
			}
		}
	}

	if err := runGovardSubcommand(cmd, govardMagentoSubcommandArgs(setupArgs...)...); err != nil {
		return fmt.Errorf("magento setup:install failed: %w", err)
	}
	return nil
}

func runBootstrapSampleData(cmd *cobra.Command) error {
	commands := [][]string{
		{"sample:deploy"},
		{"setup:upgrade"},
		{"indexer:reindex"},
		{"cache:flush"},
	}
	for _, args := range commands {
		if err := runGovardSubcommand(cmd, govardMagentoSubcommandArgs(args...)...); err != nil {
			return fmt.Errorf("sample data step failed (%s): %w", strings.Join(args, " "), err)
		}
	}
	return nil
}

func runBootstrapMagentoReindex(cmd *cobra.Command) error {
	pterm.Info.Println("Reindexing data...")
	if err := runGovardSubcommand(cmd, govardMagentoSubcommandArgs("indexer:reindex")...); err != nil {
		return fmt.Errorf("reindex failed: %w", err)
	}
	return nil
}

func runBootstrapAdminCreate(cmd *cobra.Command, config engine.Config) {
	emailDomain := config.Domain
	if emailDomain == "" {
		emailDomain = "local.test"
	}

	err := runGovardSubcommand(
		cmd,
		govardMagentoSubcommandArgs(
			"admin:user:create",
			"--admin-user=admin",
			"--admin-password=Admin123$",
			"--admin-firstname=Admin",
			"--admin-lastname=User",
			"--admin-email=admin@"+emailDomain,
		)...,
	)
	if err != nil {
		pterm.Warning.Printf("Admin user creation skipped: %v\n", err)
	}
}

func runBootstrapFixDeps(cmd *cobra.Command, opts bootstrapRuntimeOptions) {
	args := []string{"custom", "fix-deps"}
	if opts.MetaVersion != "" {
		args = append(args, "--", "--framework-version="+opts.MetaVersion)
	}
	if err := runGovardSubcommand(cmd, args...); err != nil {
		pterm.Warning.Printf("fix-deps step skipped: %v\n", err)
	}
}

func maybeAutoDetectBootstrapVersion(config engine.Config, opts *bootstrapRuntimeOptions) {
	if opts == nil {
		return
	}
	if !opts.Clone || !opts.FixDeps || strings.TrimSpace(opts.MetaVersion) != "" {
		return
	}

	remoteCfg, ok := config.Remotes[opts.Source]
	if !ok {
		return
	}

	detectedVersion, err := remote.DetectMagento2Version(opts.Source, remoteCfg)
	if err != nil {
		pterm.Warning.Printf("Could not auto-detect Magento version from remote '%s' for fix-deps (%v).\n", opts.Source, err)
		return
	}
	opts.MetaVersion = detectedVersion
	pterm.Info.Printf("Detected remote Magento version %s from '%s' for fix-deps.\n", detectedVersion, opts.Source)
}
