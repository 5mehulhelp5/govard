package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
)

type WordPressBootstrap struct {
	Options Options
}

func NewWordPressBootstrap(opts Options) *WordPressBootstrap {
	return &WordPressBootstrap{Options: opts}
}

func (w *WordPressBootstrap) Name() string {
	return "wordpress"
}

func (w *WordPressBootstrap) SupportsFreshInstall() bool {
	return true
}

func (w *WordPressBootstrap) SupportsClone() bool {
	return true
}

func (w *WordPressBootstrap) FreshCommands() []string {
	return []string{
		"wp core download",
	}
}

func (w *WordPressBootstrap) CreateProject(projectDir string) error {
	pterm.Info.Println("Creating fresh WordPress project...")

	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return fmt.Errorf("failed to read project directory: %w", err)
	}

	if len(entries) > 0 {
		pterm.Warning.Println("Project directory is not empty. Cleaning up...")
		for _, entry := range entries {
			if entry.Name() == ".govard" || entry.Name() == ".govard.yml" {
				continue
			}
			path := filepath.Join(projectDir, entry.Name())
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
			}
		}
	}

	command := "wp core download --path=. --allow-root"
	if w.Options.Runner != nil {
		if err := w.Options.Runner(command); err != nil {
			return fmt.Errorf("failed to download WordPress: %w", err)
		}
	} else {
		cmd := exec.Command("wp", "core", "download", "--path=.", "--allow-root")
		cmd.Dir = projectDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to download WordPress: %w", err)
		}
	}

	pterm.Success.Println("WordPress downloaded successfully")
	return nil
}

func (w *WordPressBootstrap) Install(projectDir string) error {
	pterm.Info.Println("Running WordPress installation steps...")

	configPath := filepath.Join(projectDir, "wp-config.php")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		dbHost := w.Options.DBHost
		if dbHost == "" {
			dbHost = "db"
		}
		dbUser := w.Options.DBUser
		if dbUser == "" {
			dbUser = "wordpress"
		}
		dbPass := w.Options.DBPass
		if dbPass == "" {
			dbPass = "wordpress"
		}
		dbName := w.Options.DBName
		if dbName == "" {
			dbName = "wordpress"
		}

		command := fmt.Sprintf("wp config create --dbname=%s --dbuser=%s --dbpass=%s --dbhost=%s --dbprefix=wp_ --allow-root",
			dbName, dbUser, dbPass, dbHost)

		if w.Options.Runner != nil {
			if err := w.Options.Runner(command); err != nil {
				return fmt.Errorf("failed to create wp-config.php: %w", err)
			}
		} else {
			cmd := exec.Command("wp", "config", "create",
				"--dbname="+dbName,
				"--dbuser="+dbUser,
				"--dbpass="+dbPass,
				"--dbhost="+dbHost,
				"--dbprefix=wp_",
				"--allow-root",
			)
			cmd.Dir = projectDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to create wp-config.php: %w", err)
			}
		}
		pterm.Success.Println("Created wp-config.php")
	}

	installCommand := "wp core install --url=http://localhost --title='WordPress Site' --admin_user=admin --admin_password=admin --admin_email=admin@local.test --allow-root"
	if w.Options.Runner != nil {
		if err := w.Options.Runner(installCommand); err != nil {
			pterm.Warning.Printf("WordPress install warning: %v\n", err)
		}
	} else {
		cmd := exec.Command("wp", "core", "install",
			"--url=http://localhost",
			"--title=WordPress Site",
			"--admin_user=admin",
			"--admin_password=admin",
			"--admin_email=admin@local.test",
			"--allow-root",
		)
		cmd.Dir = projectDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			pterm.Warning.Printf("WordPress install warning: %v\n", err)
		}
	}

	pterm.Success.Println("WordPress installation completed")
	return nil
}

func (w *WordPressBootstrap) Configure(projectDir string) error {
	pterm.Info.Println("Configuring WordPress environment...")

	configPath := filepath.Join(projectDir, "wp-config.php")
	if _, err := os.Stat(configPath); err == nil {
		if data, err := os.ReadFile(configPath); err == nil {
			content := string(data)
			content = os.Expand(content, func(key string) string {
				switch key {
				case "DB_HOST":
					return "db"
				case "DB_NAME":
					return "wordpress"
				case "DB_USER":
					return "wordpress"
				case "DB_PASSWORD":
					return "wordpress"
				default:
					return os.Getenv(key)
				}
			})
			_ = os.WriteFile(configPath, []byte(content), 0644)
		}
	}

	pterm.Success.Println("WordPress configured successfully")
	return nil
}

func (w *WordPressBootstrap) PostClone(projectDir string) error {
	pterm.Info.Println("Setting up cloned WordPress project...")

	configPath := filepath.Join(projectDir, "wp-config.php")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		examplePath := filepath.Join(projectDir, "wp-config-sample.php")
		if _, err := os.Stat(examplePath); err == nil {
			data, _ := os.ReadFile(examplePath)
			content := string(data)
			content = os.Expand(content, func(key string) string {
				switch key {
				case "DB_HOST":
					return "db"
				case "DB_NAME":
					return "wordpress"
				case "DB_USER":
					return "wordpress"
				case "DB_PASSWORD":
					return "wordpress"
				default:
					return os.Getenv(key)
				}
			})
			_ = os.WriteFile(configPath, []byte(content), 0644)
		}
	}

	if _, err := os.Stat(configPath); err == nil {
		if data, err := os.ReadFile(configPath); err == nil {
			content := string(data)
			if !strings.Contains(content, "HTTP_X_FORWARDED_PROTO") {
				pterm.Info.Println("Injecting HTTPS proxy support into wp-config.php...")
				fix := "\n// Govard: Trust HTTPS proxy\nif (isset($_SERVER['HTTP_X_FORWARDED_PROTO']) && $_SERVER['HTTP_X_FORWARDED_PROTO'] === 'https') {\n    $_SERVER['HTTPS'] = 'on';\n}\n\n"
				if strings.HasPrefix(content, "<?php") {
					content = strings.Replace(content, "<?php", "<?php"+fix, 1)
				} else {
					content = fix + content
				}
				if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
					pterm.Warning.Printf("Could not inject HTTPS proxy support: %v\n", err)
				}
			}
		}
	}

	siteURL := fmt.Sprintf("https://%s", w.Options.Domain)
	updateCommand := fmt.Sprintf(`php -r "require 'wp-load.php'; update_option('siteurl', '%s'); update_option('home', '%s');"`, siteURL, siteURL)
	if w.Options.Runner != nil {
		if err := w.Options.Runner(updateCommand); err != nil {
			pterm.Warning.Printf("Note: Could not automatically update site URLs: %v\n", err)
		} else {
			pterm.Success.Printf("WordPress site URLs updated to %s\n", siteURL)
		}
	}

	pterm.Success.Println("Post-clone setup completed")
	return nil
}
