package cmd

import (
	"fmt"
	"govard/internal/engine"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [get|set] [key] [value]",
	Short: "Manage govard.yml configuration from CLI",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		subcommand := args[0]

		switch subcommand {
		case "get":
			config := loadFullConfig()
			if len(args) < 2 {
				pterm.Error.Println("Usage: govard config get <key>")
				return
			}
			key := args[1]
			val := getConfigValue(config, key)
			fmt.Println(val)
		case "set":
			config := loadWritableConfig()
			if len(args) < 3 {
				pterm.Error.Println("Usage: govard config set <key> <value>")
				return
			}
			key, value := args[1], args[2]
			if setConfigValue(&config, key, value) {
				engine.NormalizeConfig(&config)
				saveConfig(config)
				pterm.Success.Printf("Config updated: %s = %s\n", key, value)
			} else {
				pterm.Error.Printf("Unknown config key: %s\n", key)
			}
		default:
			pterm.Error.Printf("Unknown config subcommand: %s\n", subcommand)
		}
	},
}

func getConfigValue(config engine.Config, key string) string {
	// Simple key mapping for common fields
	switch strings.ToLower(key) {
	case "project_name":
		return config.ProjectName
	case "domain":
		return config.Domain
	case "framework_version":
		return config.FrameworkVersion
	case "php_version":
		return config.Stack.PHPVersion
	case "db_type":
		return config.Stack.DBType
	case "services.web_server", "web_server":
		return config.Stack.Services.WebServer
	case "services.search", "search":
		return config.Stack.Services.Search
	case "services.cache", "cache":
		return config.Stack.Services.Cache
	}
	return "N/A"
}

func setConfigValue(config *engine.Config, key string, value string) bool {
	switch strings.ToLower(key) {
	case "project_name":
		config.ProjectName = value
	case "domain":
		config.Domain = value
	case "framework_version":
		config.FrameworkVersion = value
	case "php_version":
		config.Stack.PHPVersion = value
	case "services.web_server", "web_server":
		config.Stack.Services.WebServer = value
	case "services.search", "search":
		config.Stack.Services.Search = value
	case "services.cache", "cache":
		config.Stack.Services.Cache = value
	default:
		return false
	}
	return true
}
