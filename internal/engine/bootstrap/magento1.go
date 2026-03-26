package bootstrap

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
)

type Magento1Bootstrap struct {
	Options Options
}

func NewMagento1Bootstrap(opts Options) *Magento1Bootstrap {
	return &Magento1Bootstrap{Options: opts}
}

func (m *Magento1Bootstrap) Name() string {
	return "magento1"
}

func (m *Magento1Bootstrap) SupportsFreshInstall() bool {
	return false
}

func (m *Magento1Bootstrap) SupportsClone() bool {
	return true
}

func (m *Magento1Bootstrap) FreshCommands() []string {
	return []string{}
}

func (m *Magento1Bootstrap) CreateProject(projectDir string) error {
	return fmt.Errorf("fresh install not supported for Magento 1, use --clone instead")
}

func (m *Magento1Bootstrap) Install(projectDir string) error {
	pterm.Info.Println("Setting up Magento 1...")
	pterm.Success.Println("Magento 1 setup completed")
	return nil
}

func (m *Magento1Bootstrap) Configure(projectDir string) error {
	pterm.Info.Println("Configuring Magento 1 environment...")

	localXmlPath := filepath.Join(projectDir, "app", "etc", "local.xml")
	if _, err := os.Stat(localXmlPath); err == nil {
		pterm.Info.Println("Found local.xml configuration")
	}

	pterm.Success.Println("Magento 1 configured successfully")
	return nil
}

func (m *Magento1Bootstrap) PostClone(projectDir string) error {
	pterm.Info.Println("Setting up cloned Magento 1 project...")

	varPath := filepath.Join(projectDir, "var")
	_ = os.MkdirAll(varPath, 0777)
	_ = os.MkdirAll(filepath.Join(varPath, "cache"), 0777)
	_ = os.MkdirAll(filepath.Join(varPath, "session"), 0777)

	mediaPath := filepath.Join(projectDir, "media")
	_ = os.MkdirAll(mediaPath, 0777)

	localXmlPath := filepath.Join(projectDir, "app", "etc", "local.xml")
	if _, err := os.Stat(localXmlPath); os.IsNotExist(err) {
		if err := m.createLocalXml(projectDir); err != nil {
			pterm.Warning.Printf("Failed to create local.xml: %v\n", err)
		}
	}

	if err := m.SetConfig(projectDir); err != nil {
		pterm.Warning.Printf("Failed to configure base URLs: %v\n", err)
	}

	if err := m.CreateAdmin(projectDir); err != nil {
		pterm.Warning.Printf("Failed to create admin user: %v\n", err)
	}

	pterm.Success.Println("Post-clone setup completed")
	return nil
}

func (m *Magento1Bootstrap) SetConfig(projectDir string) error {
	baseURL := fmt.Sprintf("https://%s/", m.Options.Domain)
	containerName := fmt.Sprintf("%s-db-1", m.Options.ProjectName)

	pterm.Info.Println("Configuring Magento 1 base URLs...")
	return RunMagento1SetConfigSQL(containerName, baseURL, m.Options.DBUser, m.Options.DBPass, m.Options.DBName, "")
}

func (m *Magento1Bootstrap) CreateAdmin(projectDir string) error {
	adminEmail := fmt.Sprintf("admin@%s", m.Options.Domain)
	containerName := fmt.Sprintf("%s-db-1", m.Options.ProjectName)

	pterm.Info.Println("Creating Magento 1 admin user...")
	return RunMagento1AdminUserSQL(containerName, m.Options.DBUser, m.Options.DBPass, m.Options.DBName, "", adminEmail)
}

// createLocalXml generates app/etc/local.xml with a random 32-hex crypt key and
// the default local Warden database credentials.
func (m *Magento1Bootstrap) createLocalXml(projectDir string) error {
	cryptKey, err := generateMagento1CryptKey()
	if err != nil {
		return fmt.Errorf("failed to generate crypt key: %w", err)
	}

	localXmlContent := fmt.Sprintf(`<?xml version="1.0"?>
<config>
    <global>
        <install>
            <date><![CDATA[Wed, 01 Jan 2025 00:00:00 +0000]]></date>
        </install>
        <crypt>
            <key><![CDATA[%s]]></key>
        </crypt>
        <disable_local_modules>false</disable_local_modules>
        <resources>
            <db>
                <table_prefix><![CDATA[]]></table_prefix>
            </db>
            <default_setup>
                <connection>
                    <host><![CDATA[db]]></host>
                    <username><![CDATA[magento]]></username>
                    <password><![CDATA[magento]]></password>
                    <dbname><![CDATA[magento]]></dbname>
                    <initStatements><![CDATA[SET NAMES utf8]]></initStatements>
                    <model><![CDATA[mysql4]]></model>
                    <type><![CDATA[pdo_mysql]]></type>
                    <pdoType></pdoType>
                    <active>1</active>
                </connection>
            </default_setup>
        </resources>
        <session_save><![CDATA[files]]></session_save>
        <session_save_path><![CDATA[var/session]]></session_save_path>
    </global>
    <default>
        <web>
            <secure>
                <offloader_header><![CDATA[HTTP_X_FORWARDED_PROTO]]></offloader_header>
            </secure>
        </web>
    </default>
    <admin>
        <routers>
            <adminhtml>
                <args>
                    <frontName><![CDATA[admin]]></frontName>
                </args>
            </adminhtml>
        </routers>
    </admin>
</config>
`, cryptKey)

	etcPath := filepath.Join(projectDir, "app", "etc")
	if err := os.MkdirAll(etcPath, 0755); err != nil {
		return fmt.Errorf("failed to create app/etc directory: %w", err)
	}

	localXmlPath := filepath.Join(etcPath, "local.xml")
	if err := os.WriteFile(localXmlPath, []byte(localXmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write local.xml: %w", err)
	}

	pterm.Success.Println("Created local.xml with random crypt key")
	return nil
}

// generateMagento1CryptKey returns a random 32-character hex string for use as an encryption key.
func generateMagento1CryptKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// RunMagento1SetConfigSQL executes Magento 1 base URL configuration SQL against the local DB container.
// containerName is the docker container (e.g. "myproject-db-1"), baseURL is https://host.test/.
func RunMagento1SetConfigSQL(containerName string, baseURL string, dbUser string, dbPassword string, dbName string, dbPrefix string) error {
	sqls := []string{
		fmt.Sprintf("UPDATE %score_config_data SET value = '%s' WHERE path IN ('web/secure/base_url', 'web/unsecure/base_url')", dbPrefix, baseURL),
		"UPDATE " + dbPrefix + "core_config_data SET value = '{{secure_base_url}}' WHERE path IN ('web/unsecure/base_link_url', 'web/secure/base_link_url')",
		"UPDATE " + dbPrefix + "core_config_data SET value = '{{secure_base_url}}skin/' WHERE path IN ('web/unsecure/base_skin_url', 'web/secure/base_skin_url')",
		"UPDATE " + dbPrefix + "core_config_data SET value = '{{secure_base_url}}media/' WHERE path IN ('web/unsecure/base_media_url', 'web/secure/base_media_url')",
		"UPDATE " + dbPrefix + "core_config_data SET value = '{{secure_base_url}}js/' WHERE path IN ('web/unsecure/base_js_url', 'web/secure/base_js_url')",
		"UPDATE " + dbPrefix + "core_config_data SET value = 'HTTP_X_FORWARDED_PROTO' WHERE path = 'web/secure/offloader_header'",
		"UPDATE " + dbPrefix + "core_config_data SET value = '1' WHERE path = 'web/secure/use_in_frontend'",
		"UPDATE " + dbPrefix + "core_config_data SET value = '1' WHERE path = 'web/secure/use_in_adminhtml'",
		"UPDATE " + dbPrefix + "core_config_data SET value = '0' WHERE path = 'web/url/redirect_to_base'",
		"UPDATE " + dbPrefix + "core_config_data SET value = NULL WHERE path = 'web/cookie/cookie_domain'",
		"UPDATE " + dbPrefix + "core_config_data SET value = '/' WHERE path = 'web/cookie/cookie_path'",
	}

	for _, sql := range sqls {
		if err := RunMagento1SQL(containerName, dbUser, dbPassword, dbName, sql); err != nil {
			pterm.Warning.Printf("set-config SQL failed (continuing): %v\n", err)
		}
	}
	return nil
}

// RunMagento1AdminUserSQL inserts/updates the admin user in the local DB using a salted MD5 hash.
// This matches the approach in warden-custom-commands bootstrap.cmd for maximum M1 compatibility.
func RunMagento1AdminUserSQL(containerName string, dbUser string, dbPassword string, dbName string, dbPrefix string, adminEmail string) error {
	// Salted MD5: md5("admin" + "Admin123$") + ":admin"
	passHash := Md5SaltedHash("admin", "Admin123$")
	saltedPass := passHash + ":admin"

	insertSQL := fmt.Sprintf(`
INSERT INTO %sadmin_user(username, firstname, lastname, email, password, created, lognum, reload_acl_flag, is_active, extra, rp_token, rp_token_created_at)
VALUES ("admin", "Admin", "User", %q, %q, NOW(), 0, 0, 1, NULL, NULL, NOW())
ON DUPLICATE KEY UPDATE password = %q, is_active = 1;

-- Ensure Administrators group exists
INSERT IGNORE INTO %sadmin_role (parent_id, tree_level, sort_order, role_type, user_id, role_name)
VALUES (0, 1, 1, 'G', 0, 'Administrators');

-- Ensure full permissions
INSERT IGNORE INTO %sadmin_rule (role_id, resource_id, privileges, assert_id, role_type, permission)
SELECT role_id, 'all', NULL, 0, 'G', 'allow' FROM %sadmin_role WHERE role_type = 'G' AND role_name = 'Administrators' LIMIT 1;

-- Assign user to Administrators
INSERT INTO %sadmin_role (parent_id, tree_level, sort_order, role_type, user_id, role_name)
SELECT role_id, 2, 0, 'U', (SELECT user_id FROM %sadmin_user WHERE username = 'admin' LIMIT 1), 'admin'
FROM %sadmin_role WHERE role_type = 'G' AND role_name = 'Administrators' LIMIT 1
ON DUPLICATE KEY UPDATE parent_id = VALUES(parent_id);
`,
		dbPrefix, adminEmail, saltedPass, saltedPass,
		dbPrefix, dbPrefix, dbPrefix, dbPrefix, dbPrefix, dbPrefix)

	return RunMagento1SQL(containerName, dbUser, dbPassword, dbName, insertSQL)
}

// RunMagento1SQL executes a SQL statement via docker exec on the given DB container.
func RunMagento1SQL(containerName string, dbUser string, dbPassword string, dbName string, sql string) error {
	script := fmt.Sprintf(
		`if command -v mysql >/dev/null 2>&1; then DB_CLI=mysql; elif command -v mariadb >/dev/null 2>&1; then DB_CLI=mariadb; else exit 1; fi && echo %s | "$DB_CLI" -u %s %s -f`,
		ShellEscape(sql), ShellEscape(dbUser), ShellEscape(dbName),
	)

	args := []string{"exec", "-i"}
	if dbPassword != "" {
		args = append(args, "-e", "MYSQL_PWD="+dbPassword)
	}
	args = append(args, containerName, "sh", "-lc", script)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("SQL exec failed: %w: %s", err, out)
	}
	return nil
}
