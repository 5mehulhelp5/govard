package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"govard/internal/engine"
	"govard/internal/engine/remote"
)

type dbCredentials struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

func defaultDBCredentialsForFramework(framework string) dbCredentials {
	switch strings.TrimSpace(framework) {
	case "symfony":
		return dbCredentials{
			Port:     3306,
			Username: "symfony",
			Password: "symfony",
			Database: "symfony",
		}
	case "laravel":
		return dbCredentials{
			Port:     3306,
			Username: "laravel",
			Password: "laravel",
			Database: "laravel",
		}
	case "wordpress":
		return dbCredentials{
			Port:     3306,
			Username: "wordpress",
			Password: "wordpress",
			Database: "wordpress",
		}
	default:
		return dbCredentials{
			Port:     3306,
			Username: "magento",
			Password: "magento",
			Database: "magento",
		}
	}
}

func (credentials dbCredentials) withDefaults() dbCredentials {
	result := credentials
	if strings.TrimSpace(result.Username) == "" {
		result.Username = "magento"
	}
	if strings.TrimSpace(result.Database) == "" {
		result.Database = "magento"
	}
	if strings.TrimSpace(result.Host) != "" && result.Port <= 0 {
		result.Port = 3306
	}
	if result.Port < 0 {
		result.Port = 0
	}
	return result
}

func resolveRemoteDBCredentials(config engine.Config, remoteName string, remoteCfg engine.RemoteConfig) (dbCredentials, error) {
	fallback := defaultDBCredentialsForFramework(config.Framework)
	switch strings.TrimSpace(config.Framework) {
	case "magento2":
		metadata, err := remote.ProbeMagento2Environment(remoteName, remoteCfg)
		if err != nil {
			return fallback, err
		}

		return dbCredentials{
			Host:     metadata.DB.Host,
			Port:     metadata.DB.Port,
			Username: metadata.DB.Username,
			Password: metadata.DB.Password,
			Database: metadata.DB.Database,
		}.withDefaults(), nil
	case "magento1", "openmage":
		metadata, err := remote.ProbeMagento1Environment(remoteName, remoteCfg)
		if err != nil {
			return fallback, err
		}
		return dbCredentials{
			Host:     metadata.DB.Host,
			Port:     metadata.DB.Port,
			Username: metadata.DB.Username,
			Password: metadata.DB.Password,
			Database: metadata.DB.Database,
		}.withDefaults(), nil
	case "wordpress":
		metadata, err := remote.ProbeWordPressEnvironment(remoteName, remoteCfg)
		if err != nil {
			// Fallback to Dotenv for Bedrock-style WordPress sites
			metadataDotenv, errDotenv := remote.ProbeDotenvEnvironment(remoteName, remoteCfg)
			if errDotenv == nil {
				return dbCredentials{
					Host:     metadataDotenv.DB.Host,
					Port:     metadataDotenv.DB.Port,
					Username: metadataDotenv.DB.Username,
					Password: metadataDotenv.DB.Password,
					Database: metadataDotenv.DB.Database,
				}.withDefaults(), nil
			}
			return fallback, err
		}
		return dbCredentials{
			Host:     metadata.DB.Host,
			Port:     metadata.DB.Port,
			Username: metadata.DB.Username,
			Password: metadata.DB.Password,
			Database: metadata.DB.Database,
		}.withDefaults(), nil
	case "symfony", "laravel", "drupal", "shopware", "cakephp":
		metadata, err := remote.ProbeDotenvEnvironment(remoteName, remoteCfg)
		if err != nil {
			return fallback, err
		}
		return dbCredentials{
			Host:     metadata.DB.Host,
			Port:     metadata.DB.Port,
			Username: metadata.DB.Username,
			Password: metadata.DB.Password,
			Database: metadata.DB.Database,
		}.withDefaults(), nil
	default:
		return fallback, nil
	}
}

func resolveLocalDBCredentials(config engine.Config, containerName string) dbCredentials {
	credentials := defaultDBCredentialsForFramework(config.Framework)
	inspectCommand := exec.Command("docker", "inspect", "-f", "{{range .Config.Env}}{{println .}}{{end}}", containerName)
	output, err := inspectCommand.Output()
	if err != nil {
		return credentials
	}

	envMap := parseEnvMap(string(output))
	if user := strings.TrimSpace(envMap["MYSQL_USER"]); user != "" {
		credentials.Username = user
	}
	if password := envMap["MYSQL_PASSWORD"]; password != "" {
		credentials.Password = password
	}
	if database := strings.TrimSpace(envMap["MYSQL_DATABASE"]); database != "" {
		credentials.Database = database
	}

	return credentials.withDefaults()
}

func parseEnvMap(raw string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}
		result[strings.TrimSpace(parts[0])] = parts[1]
	}
	return result
}

func buildRemoteMySQLDumpCommandString(credentials dbCredentials, noNoise bool, noPII bool, framework string, compress bool) string {
	credentials = credentials.withDefaults()

	dbCliDetect := `if command -v mariadb-dump >/dev/null 2>&1; then DUMP_BIN=mariadb-dump; else DUMP_BIN=mysqldump; fi`

	// Common options
	commonArgs := []string{"\"$DUMP_BIN\"", "--max-allowed-packet=512M", "--force", "--single-transaction", "--no-tablespaces"}
	if host := strings.TrimSpace(credentials.Host); host != "" {
		commonArgs = append(commonArgs, "-h"+engine.ShellQuote(host))
	}
	if credentials.Port > 0 {
		commonArgs = append(commonArgs, "-P"+strconv.Itoa(credentials.Port))
	}
	commonArgs = append(commonArgs, "-u"+engine.ShellQuote(credentials.Username))

	// Pass 1: Metadata (no data, routines, triggers)
	metadataArgs := append([]string{}, commonArgs...)
	metadataArgs = append(metadataArgs, "--no-data", "--routines", "--triggers")
	metadataArgs = append(metadataArgs, engine.ShellQuote(credentials.Database))

	// Pass 2: Data (no create info, skip triggers, exclude noise/PII)
	dataArgs := append([]string{}, commonArgs...)
	dataArgs = append(dataArgs, "--no-create-info", "--skip-triggers")
	ignoreArgs := buildIgnoredTableArgs(credentials.Database, "", noNoise, noPII, framework)
	dataArgs = append(dataArgs, ignoreArgs...)
	dataArgs = append(dataArgs, engine.ShellQuote(credentials.Database))

	// Combine passes
	dumpCmd := fmt.Sprintf("{ %s; %s; }", strings.Join(metadataArgs, " "), strings.Join(dataArgs, " "))
	if compress {
		dumpCmd += " | gzip -c"
	}

	return dbCliDetect + " && " + mysqlPasswordExportPrefix(credentials.Password) + dumpCmd
}

func buildRemoteMySQLConnectCommandString(credentials dbCredentials) string {
	credentials = credentials.withDefaults()

	args := []string{"mysql"}
	if host := strings.TrimSpace(credentials.Host); host != "" {
		args = append(args, "-h"+engine.ShellQuote(host))
	}
	if credentials.Port > 0 {
		args = append(args, "-P"+strconv.Itoa(credentials.Port))
	}
	args = append(args, "-u"+engine.ShellQuote(credentials.Username), engine.ShellQuote(credentials.Database))

	return mysqlPasswordExportPrefix(credentials.Password) + strings.Join(args, " ")
}

func buildRemoteMySQLImportCommandString(credentials dbCredentials) string {
	credentials = credentials.withDefaults()

	args := []string{"mysql", "--max-allowed-packet=512M"}
	if host := strings.TrimSpace(credentials.Host); host != "" {
		args = append(args, "-h"+engine.ShellQuote(host))
	}
	if credentials.Port > 0 {
		args = append(args, "-P"+strconv.Itoa(credentials.Port))
	}
	args = append(args, "-u"+engine.ShellQuote(credentials.Username), engine.ShellQuote(credentials.Database), "-f")

	return mysqlPasswordExportPrefix(credentials.Password) + strings.Join(args, " ")
}

func buildLocalDBConnectCommand(containerName string, credentials dbCredentials) *exec.Cmd {
	credentials = credentials.withDefaults()
	args := []string{"exec", "-it"}
	if strings.TrimSpace(credentials.Password) != "" {
		args = append(args, "-e", "MYSQL_PWD="+credentials.Password)
	}
	args = append(args, containerName, "sh", "-lc", buildLocalMySQLClientCommandScript(credentials, false))
	return exec.Command("docker", args...)
}

func buildLocalDBImportCommand(containerName string, credentials dbCredentials) *exec.Cmd {
	credentials = credentials.withDefaults()
	args := []string{"exec", "-i"}
	if strings.TrimSpace(credentials.Password) != "" {
		args = append(args, "-e", "MYSQL_PWD="+credentials.Password)
	}
	args = append(args, containerName, "sh", "-lc", buildLocalMySQLClientCommandScript(credentials, true))
	return exec.Command("docker", args...)
}

func buildLocalDBDumpCommand(containerName string, credentials dbCredentials, noNoise bool, noPII bool, framework string) *exec.Cmd {
	credentials = credentials.withDefaults()
	args := []string{"exec", "-i"}
	if strings.TrimSpace(credentials.Password) != "" {
		args = append(args, "-e", "MYSQL_PWD="+credentials.Password)
	}
	args = append(args, containerName, "sh", "-lc", buildLocalMySQLDumpCommandScript(credentials, noNoise, noPII, framework))
	return exec.Command("docker", args...)
}

func buildLocalMySQLDumpCommandScript(credentials dbCredentials, noNoise bool, noPII bool, framework string) string {
	credentials = credentials.withDefaults()

	dbCliDetect := `if command -v mariadb-dump >/dev/null 2>&1; then DUMP_BIN=mariadb-dump; else DUMP_BIN=mysqldump; fi`

	// Common options
	commonArgs := []string{"\"$DUMP_BIN\"", "--max-allowed-packet=512M", "--force", "--single-transaction", "--no-tablespaces", "-hdb", "-u" + engine.ShellQuote(credentials.Username)}

	// Pass 1: Metadata
	metadataArgs := append([]string{}, commonArgs...)
	metadataArgs = append(metadataArgs, "--no-data", "--routines", "--triggers")
	metadataArgs = append(metadataArgs, engine.ShellQuote(credentials.Database))

	// Pass 2: Data
	dataArgs := append([]string{}, commonArgs...)
	dataArgs = append(dataArgs, "--no-create-info", "--skip-triggers")
	ignoreArgs := buildIgnoredTableArgs(credentials.Database, "", noNoise, noPII, framework)
	dataArgs = append(dataArgs, ignoreArgs...)
	dataArgs = append(dataArgs, engine.ShellQuote(credentials.Database))

	dumpCmd := fmt.Sprintf("{ %s; %s; }", strings.Join(metadataArgs, " "), strings.Join(dataArgs, " "))

	return dbCliDetect + " && " + dumpCmd
}

// magento1IgnoredTables is the list of ephemeral/noise tables excluded for Magento 1 when --no-noise is specified.
// Ported from warden-custom-commands (env-adapters/magento1/utils.sh).
var magento1IgnoredTables = []string{
	"catalogsearch_fulltext",
	"catalogsearch_query",
	"catalogsearch_result",
	"core_session",
	"cron_schedule",
	"enterprise_logging_event",
	"enterprise_logging_event_changes",
	"index_event",
	"log_customer",
	"log_quote",
	"log_summary",
	"log_summary_type",
	"log_url",
	"log_url_info",
	"log_visitor",
	"log_visitor_info",
	"log_visitor_online",
	"mkp_api_session_vendor",
	"report_compared_product_index",
	"report_viewed_product_index",
	"smtppro_email_log",
	"udprod_images",
}

// magento1SensitiveTables is the list of PII/sensitive tables excluded for Magento 1 when --no-pii is specified.
// Ported from warden-custom-commands (env-adapters/magento1/utils.sh).
var magento1SensitiveTables = []string{
	"admin_user",
	"api_user",
	"customer_address_entity",
	"customer_address_entity_datetime",
	"customer_address_entity_decimal",
	"customer_address_entity_int",
	"customer_address_entity_text",
	"customer_address_entity_varchar",
	"customer_entity",
	"customer_entity_datetime",
	"customer_entity_decimal",
	"customer_entity_int",
	"customer_entity_text",
	"customer_entity_varchar",
	"newsletter_subscriber",
	"sales_flat_order",
	"sales_flat_order_address",
	"sales_flat_order_grid",
	"sales_flat_order_item",
	"sales_flat_order_payment",
	"sales_flat_order_status_history",
	"sales_flat_quote",
	"sales_flat_quote_address",
	"sales_flat_quote_item",
	"sales_flat_quote_payment",
	"sales_flat_shipment",
	"sales_flat_shipment_grid",
	"sales_flat_shipment_item",
	"sales_flat_shipment_track",
	"sales_flat_invoice",
	"sales_flat_invoice_grid",
	"sales_flat_invoice_item",
	"sales_flat_creditmemo",
	"sales_flat_creditmemo_grid",
	"sales_flat_creditmemo_item",
	"wishlist",
	"wishlist_item",
}

// magento2IgnoredTables is the list of ephemeral/noise tables excluded when --no-noise is specified.
// Ported from warden-custom-commands v2.7.0 IGNORED_TABLES (env-adapters/magento2/utils.sh).
var magento2IgnoredTables = []string{
	"admin_system_messages",
	"admin_user_expiration",
	"admin_user_session",
	"adminnotification_inbox",
	"amasty_fpc_activity",
	"amasty_fpc_context_debug",
	"amasty_fpc_flushes_log",
	"amasty_fpc_job_queue",
	"amasty_fpc_log",
	"amasty_fpc_pages_to_flush",
	"amasty_fpc_queue_page",
	"amasty_fpc_reports",
	"amasty_mostviewed_product_index",
	"amasty_mostviewed_product_viewed_index",
	"amasty_mostviewed_product_viewed_index_replica",
	"amasty_reports_abandoned_cart",
	"amasty_xsearch_users_search",
	"cache_tag",
	"catalog_category_product_cl",
	"catalog_category_product_index_replica",
	"catalog_category_product_index_tmp",
	"catalog_product_attribute_cl",
	"catalog_product_category_cl",
	"catalog_product_index_eav_decimal_idx",
	"catalog_product_index_eav_decimal_tmp",
	"catalog_product_index_eav_idx",
	"catalog_product_index_eav_replica",
	"catalog_product_index_eav_tmp",
	"catalog_product_index_price_bundle_idx",
	"catalog_product_index_price_bundle_opt_idx",
	"catalog_product_index_price_bundle_opt_tmp",
	"catalog_product_index_price_bundle_sel_idx",
	"catalog_product_index_price_bundle_sel_tmp",
	"catalog_product_index_price_bundle_tmp",
	"catalog_product_index_price_cfg_opt_agr_idx",
	"catalog_product_index_price_cfg_opt_agr_tmp",
	"catalog_product_index_price_cfg_opt_idx",
	"catalog_product_index_price_cfg_opt_tmp",
	"catalog_product_index_price_downlod_idx",
	"catalog_product_index_price_downlod_tmp",
	"catalog_product_index_price_final_idx",
	"catalog_product_index_price_final_tmp",
	"catalog_product_index_price_idx",
	"catalog_product_index_price_opt_agr_idx",
	"catalog_product_index_price_opt_agr_tmp",
	"catalog_product_index_price_opt_idx",
	"catalog_product_index_price_opt_tmp",
	"catalog_product_index_price_replica",
	"catalog_product_index_price_tmp",
	"catalog_product_price_cl",
	"cataloginventory_stock_cl",
	"cataloginventory_stock_status_idx",
	"cataloginventory_stock_status_tmp",
	"catalogsearch_fulltext_cl",
	"catalogsearch_fulltext_scope1",
	"catalogsearch_fulltext_scope2",
	"cron_schedule",
	"customer_grid_flat_cl",
	"customer_log",
	"customer_visitor",
	"design_config_grid_flat_cl",
	"elasticsuite_tracker_log_customer_link",
	"elasticsuite_tracker_log_event",
	"import_history",
	"inventory_cl",
	"inventory_stock_sales_channel_cl",
	"kiwicommerce_activity",
	"kiwicommerce_activity_detail",
	"kiwicommerce_activity_log",
	"klaviyo_sync_queue",
	"login_as_customer",
	"magento_bulk",
	"magento_logging_event",
	"magento_logging_event_changes",
	"magento_login_as_customer_log",
	"mageplaza_smtp_log",
	"mailchimp_errors",
	"mailchimp_sync_batches",
	"mailchimp_sync_ecommerce",
	"mailchimp_webhook_request",
	"mst_cache_warmer_job",
	"mst_cache_warmer_page",
	"mst_cache_warmer_trace",
	"mst_search_index_store",
	"mst_seo_audit_check_result_aggregated",
	"oauth_nonce",
	"oauth_token_request_log",
	"password_reset_request_event",
	"persistent_session",
	"queue_message",
	"queue_message_status",
	"report_compared_product_index",
	"report_event",
	"report_viewed_product_aggregated_daily",
	"report_viewed_product_aggregated_monthly",
	"report_viewed_product_aggregated_yearly",
	"report_viewed_product_index",
	"reporting_module_status",
	"reporting_system_updates",
	"reporting_users",
	"sales_bestsellers_aggregated_daily",
	"sales_bestsellers_aggregated_monthly",
	"sales_bestsellers_aggregated_yearly",
	"sales_invoiced_aggregated",
	"sales_invoiced_aggregated_order",
	"sales_order_aggregated_created",
	"sales_order_aggregated_updated",
	"sales_refunded_aggregated",
	"sales_refunded_aggregated_order",
	"sales_shipping_aggregated",
	"sales_shipping_aggregated_order",
	"search_query",
	"session",
	"sutunam_activity",
	"sutunam_activity_detail",
	"sutunam_activity_log",
	"ui_bookmark",
	"yotpo_order_sync",
	"yotpo_sync_queue",
}

// magento2SensitiveTables is the list of PII/sensitive tables excluded when --no-pii is specified.
// Ported from warden-custom-commands v2.7.0 SENSITIVE_TABLES (env-adapters/magento2/utils.sh).
var magento2SensitiveTables = []string{
	"admin_passwords",
	"admin_user",
	"aw_ca_company",
	"aw_ca_company_domain",
	"aw_ca_company_payments",
	"aw_ca_company_user",
	"company",
	"company_advanced_customer_entity",
	"company_credit",
	"company_credit_history",
	"company_order_entity",
	"company_payment",
	"company_permissions",
	"company_roles",
	"company_shipping",
	"company_structure",
	"company_team",
	"company_user_roles",
	"customer_address_entity",
	"customer_address_entity_datetime",
	"customer_address_entity_decimal",
	"customer_address_entity_int",
	"customer_address_entity_text",
	"customer_address_entity_varchar",
	"customer_entity",
	"customer_entity_datetime",
	"customer_entity_decimal",
	"customer_entity_int",
	"customer_entity_text",
	"customer_entity_varchar",
	"customer_grid_flat",
	"downloadable_link_purchased",
	"downloadable_link_purchased_item",
	"email_automation",
	"email_contact",
	"magento_customerbalance",
	"magento_customerbalance_history",
	"magento_customersegment_customer",
	"magento_giftcardaccount",
	"magento_reward",
	"magento_reward_history",
	"magento_rma",
	"magento_rma_grid",
	"magento_rma_item_entity",
	"magento_rma_shipping_label",
	"magento_rma_status_history",
	"newsletter_subscriber",
	"paypal_billing_agreement",
	"paypal_billing_agreement_order",
	"paypal_payment_transaction",
	"paypal_settlement_report",
	"paypal_settlement_report_row",
	"product_alert_price",
	"product_alert_stock",
	"purchase_order_company_config",
	"quote",
	"quote_address",
	"quote_address_item",
	"quote_id_mask",
	"quote_item",
	"quote_item_option",
	"quote_payment",
	"quote_shipping_rate",
	"sales_creditmemo",
	"sales_creditmemo_comment",
	"sales_creditmemo_grid",
	"sales_creditmemo_item",
	"sales_invoice",
	"sales_invoice_comment",
	"sales_invoice_grid",
	"sales_invoice_item",
	"sales_order",
	"sales_order_address",
	"sales_order_grid",
	"sales_order_item",
	"sales_order_payment",
	"sales_order_status_history",
	"sales_order_tax",
	"sales_order_tax_item",
	"sales_payment_transaction",
	"sales_shipment",
	"sales_shipment_comment",
	"sales_shipment_grid",
	"sales_shipment_item",
	"sales_shipment_track",
	"vault_payment_token",
	"vault_payment_token_order_payment_link",
	"wishlist",
	"wishlist_item",
	"wishlist_item_option",
}

var laravelIgnoredTables = []string{
	"cache",
	"cache_locks",
	"failed_jobs",
	"job_batches",
	"jobs",
	"sessions",
	"telescope_entries",
	"telescope_entries_tags",
	"telescope_monitoring",
}

var laravelSensitiveTables = []string{
	"password_reset_tokens",
	"password_resets",
	"personal_access_tokens",
	"users",
}

var wordpressIgnoredTables = []string{
	"options_bak",
	"options_replica",
	"options_tmp",
	"redirection_404",
	"wflogs",
}

var wordpressSensitiveTables = []string{
	"commentmeta",
	"comments",
	"usermeta",
	"users",
}

// buildIgnoredTableArgs returns docker exec --ignore-table flags for the given credentials and filter flags.
func buildIgnoredTableArgs(dbName string, dbPrefix string, noNoise bool, noPII bool, framework string) []string {
	tables := getIgnoredTableList(noNoise, noPII, framework)
	if len(tables) == 0 {
		return nil
	}

	args := make([]string, 0, len(tables))
	for _, t := range tables {
		args = append(args, "--ignore-table="+dbName+"."+dbPrefix+t)
	}
	return args
}

func getIgnoredTableList(noNoise bool, noPII bool, framework string) []string {
	if !noNoise && !noPII {
		return nil
	}

	var ignored []string
	var sensitive []string

	switch strings.TrimSpace(framework) {
	case "laravel":
		ignored = laravelIgnoredTables
		sensitive = laravelSensitiveTables
	case "wordpress":
		ignored = wordpressIgnoredTables
		sensitive = wordpressSensitiveTables
	case "magento1", "openmage":
		ignored = magento1IgnoredTables
		sensitive = magento1SensitiveTables
	default:
		// Default to magento2 behavior
		ignored = magento2IgnoredTables
		sensitive = magento2SensitiveTables
	}

	tables := make([]string, 0)
	if noNoise {
		tables = append(tables, ignored...)
	}
	if noPII {
		tables = append(tables, sensitive...)
	}
	return tables
}

func mysqlPasswordExportPrefix(password string) string {
	if strings.TrimSpace(password) == "" {
		return ""
	}
	return "export MYSQL_PWD=" + engine.ShellQuote(password) + "; "
}

func buildLocalMySQLClientCommandScript(credentials dbCredentials, force bool) string {
	credentials = credentials.withDefaults()

	query := "exec \"$DB_CLI\" --max-allowed-packet=512M -u " + engine.ShellQuote(credentials.Username) + " " + engine.ShellQuote(credentials.Database)
	if force {
		query += " -f"
		query = "{ echo \"SET FOREIGN_KEY_CHECKS=0; SET UNIQUE_CHECKS=0; SET AUTOCOMMIT=0;\"; cat; echo \"COMMIT; SET FOREIGN_KEY_CHECKS=1; SET UNIQUE_CHECKS=1; SET AUTOCOMMIT=1;\"; } | " + query
	}

	return strings.Join([]string{
		`if command -v mysql >/dev/null 2>&1; then DB_CLI=mysql; elif command -v mariadb >/dev/null 2>&1; then DB_CLI=mariadb; else echo "mysql client not found (mysql/mariadb)" >&2; exit 127; fi`,
		query,
	}, " && ")
}

func formatRemoteDBProbeWarning(remoteName string, err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("Could not auto-detect DB credentials for '%s' from remote metadata (.env/env.php) (%v). Falling back to default credentials.", remoteName, err)
}

func BuildRemoteMySQLDumpCommandForTest(host string, port int, username string, password string, database string, compress bool) string {
	return buildRemoteMySQLDumpCommandString(dbCredentials{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
	}, false, false, "magento2", compress)
}

func BuildLocalDBImportCommandForTest(containerName string, username string, password string, database string) []string {
	command := buildLocalDBImportCommand(containerName, dbCredentials{
		Username: username,
		Password: password,
		Database: database,
	})
	return command.Args
}

func ParseEnvMapForTest(raw string) map[string]string {
	return parseEnvMap(raw)
}

// BuildIgnoredTableArgsForTest exposes buildIgnoredTableArgs for tests.
func BuildIgnoredTableArgsForTest(dbName string, dbPrefix string, noNoise bool, noPII bool, framework string) []string {
	return buildIgnoredTableArgs(dbName, dbPrefix, noNoise, noPII, framework)
}

func buildLocalDBQueryCommand(containerName string, credentials dbCredentials, query string) *exec.Cmd {
	credentials = credentials.withDefaults()
	args := []string{"exec", "-i"}
	if strings.TrimSpace(credentials.Password) != "" {
		args = append(args, "-e", "MYSQL_PWD="+credentials.Password)
	}
	args = append(args, containerName, "sh", "-lc", buildLocalMySQLQueryCommandScript(credentials, query))
	return exec.Command("docker", args...)
}

func buildLocalMySQLQueryCommandScript(credentials dbCredentials, query string) string {
	credentials = credentials.withDefaults()

	escapedQuery := strings.ReplaceAll(query, "'", "'\"'\"'")
	queryCmd := "exec \"$DB_CLI\" -u " + engine.ShellQuote(credentials.Username) + " -e '" + escapedQuery + "'" + " " + engine.ShellQuote(credentials.Database)

	return strings.Join([]string{
		`if command -v mysql >/dev/null 2>&1; then DB_CLI=mysql; elif command -v mariadb >/dev/null 2>&1; then DB_CLI=mariadb; else echo "mysql client not found (mysql/mariadb)" >&2; exit 127; fi`,
		queryCmd,
	}, " && ")
}

func buildRemoteMySQLQueryCommandString(credentials dbCredentials, query string) string {
	credentials = credentials.withDefaults()

	args := []string{"mysql"}
	if host := strings.TrimSpace(credentials.Host); host != "" {
		args = append(args, "-h"+engine.ShellQuote(host))
	}
	if credentials.Port > 0 {
		args = append(args, "-P"+strconv.Itoa(credentials.Port))
	}
	args = append(args, "-u"+engine.ShellQuote(credentials.Username), "-e", engine.ShellQuote(query))

	return mysqlPasswordExportPrefix(credentials.Password) + strings.Join(args, " ")
}
func GetDatabaseSize(config engine.Config, remoteName string, remoteCfg engine.RemoteConfig, credentials dbCredentials, noNoise bool, noPII bool) (int64, error) {
	ignoredTables := getIgnoredTableList(noNoise, noPII, config.Framework)
	whereClause := fmt.Sprintf("WHERE table_schema = '%s'", strings.ReplaceAll(credentials.Database, "'", "''"))
	if len(ignoredTables) > 0 {
		quotedTables := make([]string, len(ignoredTables))
		for i, t := range ignoredTables {
			quotedTables[i] = "'" + strings.ReplaceAll(t, "'", "''") + "'"
		}
		whereClause += fmt.Sprintf(" AND table_name NOT IN (%s)", strings.Join(quotedTables, ","))
	}

	// query the total logical size (data_length is better for estimating dump size than avg_row_length)
	query := fmt.Sprintf("SELECT SUM(data_length) FROM information_schema.tables %s", whereClause)

	credentials = credentials.withDefaults()
	mysqlArgs := []string{"\"$DB_CLI\"", "-BN"}
	if host := strings.TrimSpace(credentials.Host); host != "" {
		mysqlArgs = append(mysqlArgs, "-h"+engine.ShellQuote(host))
	}
	if credentials.Port > 0 {
		mysqlArgs = append(mysqlArgs, "-P"+strconv.Itoa(credentials.Port))
	}
	mysqlArgs = append(mysqlArgs, "-u"+engine.ShellQuote(credentials.Username), "-e", engine.ShellQuote(query))

	dbCliDetect := `if command -v mysql >/dev/null 2>&1; then DB_CLI=mysql; elif command -v mariadb >/dev/null 2>&1; then DB_CLI=mariadb; else echo "mysql client not found" >&2; exit 127; fi`
	mysqlCmd := mysqlPasswordExportPrefix(credentials.Password) + strings.Join(mysqlArgs, " ")
	cmdStr := fmt.Sprintf("%s && %s", dbCliDetect, mysqlCmd)

	var output []byte
	var err error
	if remoteName == "local" {
		containerName := fmt.Sprintf("%s-db-1", config.ProjectName)
		output, err = exec.Command("docker", "exec", containerName, "sh", "-c", cmdStr).CombinedOutput()
	} else {
		sshCmd := remote.BuildSSHExecCommand(remoteName, remoteCfg, true, cmdStr)
		output, err = sshCmd.CombinedOutput()
	}

	if err != nil {
		return 0, err
	}

	totalSizeStr := strings.TrimSpace(string(output))
	if totalSizeStr == "" || totalSizeStr == "NULL" {
		return 0, nil
	}

	var logicalSize int64
	_, _ = fmt.Sscanf(totalSizeStr, "%d", &logicalSize)

	// Since mysqldump generates a compact SQL text file while InnoDB stores data in 16KB pages
	// (often with significant internal overhead/fragmentation), the logical size is usually
	// an overestimate. We apply a 0.6 heuristic to bring it closer to actual dump results.
	targetSize := int64(float64(logicalSize) * 0.6)

	return targetSize, nil
}
