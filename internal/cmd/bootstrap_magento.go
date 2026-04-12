package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"govard/internal/engine/bootstrap"
	"os"
	"path/filepath"
	"strings"

	"govard/internal/engine"
	"govard/internal/engine/remote"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func ensureBootstrapMagentoEnvPHP(config engine.Config, opts BootstrapRuntimeOptions) error {
	if config.Framework != "magento2" {
		return nil
	}

	cwd, _ := os.Getwd()
	envPath := filepath.Join(cwd, "app", "etc", "env.php")

	if info, err := os.Lstat(envPath); err == nil && (info.Mode()&os.ModeSymlink) != 0 {
		if _, err := os.Stat(envPath); err != nil {
			if err := os.Remove(envPath); err != nil {
				return fmt.Errorf("failed to remove env.php symlink: %w", err)
			}
		} else {
			return nil
		}
	} else if err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(envPath), bootstrap.DefaultDirPerm); err != nil {
		return fmt.Errorf("failed to create app/etc: %w", err)
	}

	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Errorf("failed to generate random bytes: %w", err)
	}
	cryptKey := hex.EncodeToString(randomBytes)

	if remoteCfg, ok := config.Remotes[opts.Source]; ok {
		if metadata, err := remote.ProbeMagento2Environment(opts.Source, remoteCfg); err == nil {
			if strings.TrimSpace(metadata.CryptKey) != "" {
				cryptKey = strings.TrimSpace(metadata.CryptKey)
			}
		} else {
			pterm.Warning.Printf("Could not extract crypt/key from remote env.php (%v). Using fallback key.\n", err)
		}
	}

	containerName := fmt.Sprintf("%s-db-1", config.ProjectName)
	localDB := resolveLocalDBCredentials(config, containerName)

	template := fmt.Sprintf(`<?php
return [
    'backend' => [
        'frontName' => 'admin'
    ],
    'crypt' => [
        'key' => %q
    ],
    'db' => [
        'table_prefix' => '',
        'connection' => [
            'default' => [
                'host' => 'db',
                'dbname' => %q,
                'username' => %q,
                'password' => %q,
                'active' => '1'
            ],
            'indexer' => [
                'host' => 'db',
                'dbname' => %q,
                'username' => %q,
                'password' => %q,
                'active' => '1'
            ]
        ]
    ],
    'resource' => [
        'default_setup' => [
            'connection' => 'default'
        ]
    ],
    'x-frame-options' => 'SAMEORIGIN',
    'MAGE_MODE' => 'developer',
    'session' => [
        'save' => 'files'
    ],
    'install' => [
        'date' => 'Mon, 01 May 2023 00:00:00 +0000'
    ]
];
`, cryptKey,
		localDB.Database, localDB.Username, localDB.Password,
		localDB.Database, localDB.Username, localDB.Password,
	)

	if err := os.WriteFile(envPath, []byte(template), bootstrap.DefaultFilePerm); err != nil {
		return fmt.Errorf("failed to write app/etc/env.php: %w", err)
	}

	pterm.Info.Println("Generated local app/etc/env.php for bootstrap.")
	return nil
}

func runMagentoSearchHostFixViaCLI(cmd *cobra.Command, config engine.Config) error {
	host := "elasticsearch"
	if s := strings.ToLower(strings.TrimSpace(config.Stack.Services.Search)); s != "" && s != "none" {
		host = s
	}
	searchEngine := engine.ResolveMagentoSearchEngine(config)
	sql := engine.BuildMagentoSearchHostFixSQL(host, searchEngine)
	// Skip the --environment flag implicitly because we're running it locally
	err := runGovardSubcommand(cmd, "db", "query", sql)
	if err != nil {
		pterm.Warning.Printf("Could not fix search host via 'govard db query' (continuing): %v\n", err)
	}
	return err
}
