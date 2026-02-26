package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
)

type magentoCommand struct {
	Desc     string
	Args     []string
	Optional bool
}

// MagentoConfigCommandsForTest exposes command planning for tests.
func MagentoConfigCommandsForTest(projectName string, config Config) []magentoCommand {
	return buildMagentoCommands(projectName, config)
}

func ConfigureMagento(projectName string, config Config) error {
	pterm.Info.Println("Configuring Magento 2 environment...")

	if patched, err := patchMagentoElasticsearchSchemaForLibxml(); err != nil {
		pterm.Warning.Printf("Could not apply Magento XML schema compatibility patch (continuing): %v\n", err)
	} else if patched {
		pterm.Info.Println("Applied Magento XML schema compatibility patch for newer libxml2.")
	}

	containerName := fmt.Sprintf("%s-php-1", projectName)
	if err := ensureMagentoLocalWritableDirs(containerName, config); err != nil {
		pterm.Warning.Printf("Could not prepare Magento writable dirs (continuing): %v\n", err)
	}

	commands := buildMagentoCommands(projectName, config)

	for _, cmd := range commands {
		pterm.Info.Printf("→ %s...\n", cmd.Desc)
		output, err := exec.Command("docker", cmd.Args...).CombinedOutput()
		if err != nil {
			outText := string(output)
			// After importing a database snapshot, Magento may require config import/upgrade before
			// certain CLI commands (config:set, cache, etc) are allowed.
			if needsConfigImport(outText) {
				pterm.Warning.Println("Magento requires app:config:import/setup:upgrade before continuing; attempting auto-repair...")
				_ = ensureMagentoLocalWritableDirs(containerName, config)
				if repairErr := runMagentoConfigImport(containerName, config); repairErr != nil {
					// Fallback to setup:upgrade when import isn't enough.
					pterm.Warning.Printf("app:config:import failed (%v). Trying setup:upgrade...\n", repairErr)
					_ = ensureMagentoLocalWritableDirs(containerName, config)
					if upgradeErr := runMagentoSetupUpgrade(containerName, config); upgradeErr != nil {
						return fmt.Errorf("command failed: %s %v\nOutput: %s", cmd.Desc, err, outText)
					}
				}

				// Retry once after repair.
				output, err = exec.Command("docker", cmd.Args...).CombinedOutput()
				if err == nil {
					continue
				}
				outText = string(output)
			}

			// Some optional settings are unavailable when the related Magento module
			// is disabled (for example TwoFactorAuth on projects that explicitly turn it off).
			// In that case skip the step without warning noise.
			if cmd.Optional && isMagentoConfigPathUnavailable(outText) {
				pterm.Info.Printf("Skipping optional Magento configure step (%s): setting path is unavailable in this project.\n", cmd.Desc)
				continue
			}

			if strings.Contains(string(output), "not found") || strings.Contains(string(output), "No such container") {
				return fmt.Errorf("container %s is not running. Run 'govard env up' first", fmt.Sprintf("%s-php-1", projectName))
			}
			if cmd.Optional {
				pterm.Warning.Printf("Non-fatal Magento configure step failed (%s): %v\n", cmd.Desc, err)
				continue
			}
			return fmt.Errorf("command failed: %s %v\nOutput: %s", cmd.Desc, err, outText)
		}
	}

	pterm.Success.Println("Magento 2 environment configured successfully!")
	return nil
}

func needsConfigImport(output string) bool {
	output = strings.ToLower(output)
	return strings.Contains(output, "app:config:import") && strings.Contains(output, "setup:upgrade")
}

func isMagentoConfigPathUnavailable(output string) bool {
	normalized := strings.ToLower(strings.TrimSpace(output))
	if normalized == "" {
		return false
	}

	// English CLI message: The "..." path doesn't exist...
	if strings.Contains(normalized, "path") && strings.Contains(normalized, "doesn't exist") {
		return true
	}

	// French CLI message: Le chemin "..." n'existe pas...
	if strings.Contains(normalized, "chemin") && strings.Contains(normalized, "n'existe pas") {
		return true
	}

	return false
}

func runMagentoConfigImport(containerName string, config Config) error {
	args := magentoDockerExecArgs(containerName, config, "bin/magento", "app:config:import", "--no-interaction")
	output, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("app:config:import failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func runMagentoSetupUpgrade(containerName string, config Config) error {
	args := magentoDockerExecArgs(containerName, config, "bin/magento", "setup:upgrade", "--no-interaction")
	output, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("setup:upgrade failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func ensureMagentoLocalWritableDirs(containerName string, config Config) error {
	script := strings.Join([]string{
		"set -e",
		`fix_dir() { p="$1"; if [ -L "$p" ]; then rm -f "$p"; fi; mkdir -p "$p"; }`,
		"mkdir -p generated pub/static pub/media var",
		"rm -rf generated/code/*",
		"fix_dir var/session",
		"fix_dir var/tmp",
		"fix_dir var/report",
		"fix_dir var/import",
		"fix_dir var/export",
		"fix_dir var/import_history",
		"fix_dir var/importexport",
		"fix_dir pub/static/_cache",
		"fix_dir pub/.well-known",
	}, " && ")

	args := magentoDockerExecArgs(containerName, config, "sh", "-lc", script)
	output, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ensure writable dirs failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func buildMagentoCommands(projectName string, config Config) []magentoCommand {
	containerName := fmt.Sprintf("%s-php-1", projectName)
	searchEngine := resolveMagentoSearchEngine(config)

	configSetArgs := []string{
		"bin/magento",
		"setup:config:set",
		"--db-host=db",
		"--db-name=magento",
		"--db-user=magento",
		"--db-password=magento",
	}
	configSetArgs = append(configSetArgs, "--no-interaction")

	commands := []magentoCommand{{
		Desc: "Setting Database connection",
		Args: magentoDockerExecArgs(containerName, config, configSetArgs...),
	}}

	if searchEngine != "" {
		commands = append(commands, magentoCommand{
			Desc: "Setting Search Engine",
			Args: magentoDockerExecArgs(containerName, config, "bin/magento", "config:set",
				"catalog/search/engine", searchEngine, "--no-interaction"),
			Optional: true,
		})
		commands = append(commands, buildMagentoSearchConfigSetCommands(containerName, config, searchEngine)...)
	}

	if config.Stack.Services.Cache == "redis" || config.Stack.Services.Cache == "valkey" {
		commands = append(commands, magentoCommand{
			Desc: "Configuring Redis Cache",
			Args: magentoDockerExecArgs(containerName, config, "bin/magento", "setup:config:set",
				"--cache-backend=redis", "--cache-backend-redis-server=redis", "--cache-backend-redis-db=0", "--no-interaction"),
			Optional: true,
		})
		commands = append(commands, magentoCommand{
			Desc: "Configuring Redis Sessions",
			Args: magentoDockerExecArgs(containerName, config, "bin/magento", "setup:config:set",
				"--session-save=redis", "--session-save-redis-host=redis", "--session-save-redis-db=2", "--no-interaction"),
			Optional: true,
		})
	}

	if config.Stack.Features.Varnish {
		commands = append(commands, magentoCommand{
			Desc: "Configuring Varnish Page Cache",
			Args: magentoDockerExecArgs(containerName, config, "bin/magento", "config:set",
				"system/full_page_cache/caching_application", "2", "--no-interaction"),
		})
	}

	if config.Domain != "" {
		baseUrl := fmt.Sprintf("https://%s/", config.Domain)
		commands = append(commands, magentoCommand{
			Desc: "Setting Base URLs",
			Args: magentoDockerExecArgs(containerName, config, "bin/magento", "setup:store-config:set",
				"--base-url="+baseUrl, "--base-url-secure="+baseUrl, "--no-interaction"),
		})
	}

	// Per-store base URLs
	for domain, storeCode := range config.StoreDomains {
		baseURL := fmt.Sprintf("https://%s/", domain)
		commands = append(commands, magentoCommand{
			Desc: fmt.Sprintf("Setting Base URL for store %s", storeCode),
			Args: magentoDockerExecArgs(containerName, config, "bin/magento", "config:set",
				"--scope=stores", "--scope-code="+storeCode, "web/unsecure/base_url", baseURL, "--no-interaction"),
			Optional: true,
		}, magentoCommand{
			Desc: fmt.Sprintf("Setting Secure Base URL for store %s", storeCode),
			Args: magentoDockerExecArgs(containerName, config, "bin/magento", "config:set",
				"--scope=stores", "--scope-code="+storeCode, "web/secure/base_url", baseURL, "--no-interaction"),
			Optional: true,
		})
	}

	commands = append(commands, magentoCommand{
		Desc: "Enable Web Server Rewrites",
		Args: magentoDockerExecArgs(containerName, config, "bin/magento", "config:set",
			"web/seo/use_rewrites", "1", "--no-interaction"),
		Optional: true,
	})

	commands = append(commands, magentoCommand{
		Desc: "Disable reCAPTCHA",
		Args: magentoDockerExecArgs(containerName, config, "bin/magento", "config:set",
			"recaptcha_frontend/type_for/customer_login", "invisible", "--no-interaction"),
		Optional: true,
	})
	commands = append(commands, magentoCommand{
		Desc: "Disable 2FA",
		Args: magentoDockerExecArgs(containerName, config, "bin/magento", "config:set",
			"twofactorauth/general/enable", "0", "--no-interaction"),
		Optional: true,
	})
	commands = append(commands, magentoCommand{
		Desc: "Enable Developer Mode",
		Args: magentoDockerExecArgs(containerName, config, "bin/magento", "deploy:mode:set", "developer", "--no-interaction"),
	})

	return commands
}

func buildMagentoSearchConfigSetCommands(containerName string, config Config, engineName string) []magentoCommand {
	prefix := resolveMagentoSearchConfigPrefix(engineName)
	if prefix == "" {
		return nil
	}

	settings := []struct {
		desc  string
		path  string
		value string
	}{
		{desc: "Setting Search Host", path: "catalog/search/" + prefix + "_server_hostname", value: "elasticsearch"},
		{desc: "Setting Search Port", path: "catalog/search/" + prefix + "_server_port", value: "9200"},
		{desc: "Setting Search Index Prefix", path: "catalog/search/" + prefix + "_index_prefix", value: "magento2"},
		{desc: "Setting Search Auth", path: "catalog/search/" + prefix + "_enable_auth", value: "0"},
		{desc: "Setting Search Timeout", path: "catalog/search/" + prefix + "_server_timeout", value: "15"},
	}

	commands := make([]magentoCommand, 0, len(settings))
	for _, setting := range settings {
		commands = append(commands, magentoCommand{
			Desc: setting.desc,
			Args: magentoDockerExecArgs(
				containerName,
				config,
				"bin/magento",
				"config:set",
				setting.path,
				setting.value,
				"--no-interaction",
			),
			Optional: true,
		})
	}
	return commands
}

func resolveMagentoSearchConfigPrefix(engineName string) string {
	switch engineName {
	case "opensearch":
		return "opensearch"
	case "elasticsearch7":
		return "elasticsearch7"
	default:
		return ""
	}
}

func resolveMagentoSearchEngine(config Config) string {
	// ElasticSuite must remain the selected engine when the module is present.
	// Forcing elasticsearch7/opensearch breaks Smile query objects on Magento 2.4.7 stacks.
	if isMagentoElasticsuiteProject() {
		return "elasticsuite"
	}

	search := strings.ToLower(strings.TrimSpace(config.Stack.Services.Search))
	if search == "" || search == "none" {
		return ""
	}

	// Magento < 2.4.8 uses the elasticsearch7 engine name/flags even when running OpenSearch.
	if isMagentoVersionAtLeast(config.FrameworkVersion, "2.4.8") && search == "opensearch" {
		return "opensearch"
	}
	return "elasticsearch7"
}

func isMagentoElasticsuiteProject() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	elasticSuiteCore := filepath.Join(cwd, "vendor", "smile", "elasticsuite", "src", "module-elasticsuite-core")
	if info, statErr := os.Stat(elasticSuiteCore); statErr == nil && info.IsDir() {
		return true
	}

	configPHP := filepath.Join(cwd, "app", "etc", "config.php")
	data, readErr := os.ReadFile(configPHP)
	if readErr != nil {
		return false
	}
	return strings.Contains(string(data), "'Smile_ElasticsuiteCore' => 1")
}

func patchMagentoElasticsearchSchemaForLibxml() (bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, err
	}

	schemaPath := filepath.Join(cwd, "vendor", "magento", "module-elasticsearch", "etc", "esconfig.xsd")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	content := string(data)
	legacyBlock := `<xs:complexType name="mixedDataType">
    <xs:choice maxOccurs="unbounded" minOccurs="1">
        <xs:element type="xs:string" name="default" minOccurs="1" maxOccurs="1" />
        <xs:any processContents="lax" minOccurs="0" maxOccurs="unbounded" />
    </xs:choice>
</xs:complexType>`

	replacementBlock := `<xs:complexType name="mixedDataType">
    <xs:sequence>
        <xs:element type="xs:string" name="type" minOccurs="0" maxOccurs="1" />
        <xs:element type="xs:string" name="default" minOccurs="1" maxOccurs="1" />
        <xs:any processContents="lax" minOccurs="0" maxOccurs="unbounded" />
    </xs:sequence>
</xs:complexType>`

	if !strings.Contains(content, legacyBlock) {
		return false, nil
	}

	updated := strings.Replace(content, legacyBlock, replacementBlock, 1)
	if updated == content {
		return false, nil
	}

	if err := os.WriteFile(schemaPath, []byte(updated), 0644); err != nil {
		return false, err
	}
	return true, nil
}

func isMagentoVersionAtLeast(raw string, minimum string) bool {
	return isNumericDotVersionAtLeast(raw, minimum)
}

func magentoDockerExecArgs(containerName string, config Config, args ...string) []string {
	result := []string{"exec"}
	if user := resolveMagentoExecUser(config); strings.TrimSpace(user) != "" {
		result = append(result, "-u", user)
	}
	result = append(result, "-w", "/var/www/html", containerName)
	result = append(result, args...)
	return result
}

func resolveMagentoExecUser(config Config) string {
	if config.Stack.UserID > 0 && config.Stack.GroupID > 0 {
		return fmt.Sprintf("%d:%d", config.Stack.UserID, config.Stack.GroupID)
	}
	return "www-data"
}
