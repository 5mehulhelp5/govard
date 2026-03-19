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

type OpenMageBootstrap struct {
	Options Options
}

func NewOpenMageBootstrap(opts Options) *OpenMageBootstrap {
	return &OpenMageBootstrap{Options: opts}
}

func (o *OpenMageBootstrap) Name() string {
	return "openmage"
}

func (o *OpenMageBootstrap) SupportsFreshInstall() bool {
	return true
}

func (o *OpenMageBootstrap) SupportsClone() bool {
	return true
}

func (o *OpenMageBootstrap) FreshCommands() []string {
	return []string{
		"composer create-project openmage/magento-lts .",
	}
}

func (o *OpenMageBootstrap) CreateProject(projectDir string) error {
	pterm.Info.Println("Creating fresh OpenMage project...")

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

	version := o.Options.Version
	packageName := "openmage/magento-lts"
	if version != "" {
		packageName = fmt.Sprintf("openmage/magento-lts:%s", version)
	}

	cmd := exec.Command("composer", "create-project", packageName, ".", "--no-interaction")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create OpenMage project: %w", err)
	}

	pterm.Success.Println("OpenMage project created successfully")
	return nil
}

func (o *OpenMageBootstrap) Install(projectDir string) error {
	pterm.Info.Println("Running OpenMage installation steps...")

	if err := o.runComposerCommand(projectDir, "install", "--no-interaction"); err != nil {
		pterm.Warning.Printf("Composer install warning: %v\n", err)
	}

	localXmlPath := filepath.Join(projectDir, "app", "etc", "local.xml")
	if _, err := os.Stat(localXmlPath); os.IsNotExist(err) {
		pterm.Info.Println("Creating local.xml configuration...")
		if err := o.createLocalXml(projectDir); err != nil {
			pterm.Warning.Printf("Failed to create local.xml: %v\n", err)
		}
	}

	varPath := filepath.Join(projectDir, "var")
	_ = os.MkdirAll(varPath, 0777)
	_ = os.MkdirAll(filepath.Join(varPath, "cache"), 0777)
	_ = os.MkdirAll(filepath.Join(varPath, "session"), 0777)
	_ = os.MkdirAll(filepath.Join(varPath, "log"), 0777)

	mediaPath := filepath.Join(projectDir, "media")
	_ = os.MkdirAll(mediaPath, 0777)

	pterm.Success.Println("OpenMage installation completed")
	return nil
}

func (o *OpenMageBootstrap) Configure(projectDir string) error {
	pterm.Info.Println("Configuring OpenMage environment...")

	localXmlPath := filepath.Join(projectDir, "app", "etc", "local.xml")
	if _, err := os.Stat(localXmlPath); os.IsNotExist(err) {
		if err := o.createLocalXml(projectDir); err != nil {
			return err
		}
	}

	pterm.Success.Println("OpenMage configured successfully")
	return nil
}

func (o *OpenMageBootstrap) PostClone(projectDir string) error {
	pterm.Info.Println("Setting up cloned OpenMage project...")

	if err := o.runComposerCommand(projectDir, "install", "--no-interaction"); err != nil {
		return fmt.Errorf("composer install failed: %w", err)
	}

	localXmlPath := filepath.Join(projectDir, "app", "etc", "local.xml")
	if _, err := os.Stat(localXmlPath); os.IsNotExist(err) {
		if err := o.createLocalXml(projectDir); err != nil {
			pterm.Warning.Printf("Failed to create local.xml: %v\n", err)
		}
	}

	varPath := filepath.Join(projectDir, "var")
	_ = os.MkdirAll(varPath, 0777)
	_ = os.MkdirAll(filepath.Join(varPath, "cache"), 0777)
	_ = os.MkdirAll(filepath.Join(varPath, "session"), 0777)

	pterm.Success.Println("Post-clone setup completed")
	return nil
}
func (o *OpenMageBootstrap) createLocalXml(projectDir string) error {
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
                    <username><![CDATA[openmage]]></username>
                    <password><![CDATA[openmage]]></password>
                    <dbname><![CDATA[openmage]]></dbname>
                    <initStatements><![CDATA[SET NAMES utf8]]></initStatements>
                    <model><![CDATA[mysql4]]></model>
                    <type><![CDATA[pdo_mysql]]></type>
                    <pdoType><![CDATA[]]></pdoType>
                    <active>1</active>
                </connection>
            </default_setup>
        </resources>
        <session_save><![CDATA[files]]></session_save>
        <session_save_path><![CDATA[var/session]]></session_save_path>
    </global>
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

// runMagento1SetConfigSQL executes Magento 1 base URL configuration SQL against the local DB container.
// containerName is the docker container (e.g. "myproject-db-1"), baseURL is https://host.test/.
func runMagento1SetConfigSQL(containerName string, baseURL string, dbUser string, dbPassword string, dbName string, dbPrefix string) error {
	sqls := []string{
		fmt.Sprintf("UPDATE %score_config_data SET value = '%s' WHERE path = 'web/secure/base_url'", dbPrefix, baseURL),
		fmt.Sprintf("UPDATE %score_config_data SET value = '%s' WHERE path = 'web/unsecure/base_url'", dbPrefix, baseURL),
		"UPDATE " + dbPrefix + "core_config_data SET value = '{{secure_base_url}}' WHERE path = 'web/unsecure/base_link_url'",
		"UPDATE " + dbPrefix + "core_config_data SET value = '{{secure_base_url}}skin/' WHERE path = 'web/unsecure/base_skin_url'",
		"UPDATE " + dbPrefix + "core_config_data SET value = '{{secure_base_url}}media/' WHERE path = 'web/unsecure/base_media_url'",
		"UPDATE " + dbPrefix + "core_config_data SET value = '{{secure_base_url}}js/' WHERE path = 'web/unsecure/base_js_url'",
		"UPDATE " + dbPrefix + "core_config_data SET value = '1' WHERE path = 'web/secure/use_in_frontend'",
		"UPDATE " + dbPrefix + "core_config_data SET value = '1' WHERE path = 'web/secure/use_in_adminhtml'",
		"UPDATE " + dbPrefix + "core_config_data SET value = NULL WHERE path = 'web/cookie/cookie_domain'",
	}

	for _, sql := range sqls {
		if err := runMagento1SQL(containerName, dbUser, dbPassword, dbName, sql); err != nil {
			pterm.Warning.Printf("set-config SQL failed (continuing): %v\n", err)
		}
	}
	return nil
}

// runMagento1AdminUserSQL inserts/updates the admin user in the local DB using a salted MD5 hash.
// This matches the approach in warden-custom-commands bootstrap.cmd for maximum M1 compatibility.
func runMagento1AdminUserSQL(containerName string, dbUser string, dbPassword string, dbName string, dbPrefix string, adminEmail string) error {
	// Salted MD5: md5("admin" + "Admin123$") + ":admin"
	passHash := md5SaltedHash("admin", "Admin123$")
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

	return runMagento1SQL(containerName, dbUser, dbPassword, dbName, insertSQL)
}

// runMagento1SQL executes a SQL statement via docker exec on the given DB container.
func runMagento1SQL(containerName string, dbUser string, dbPassword string, dbName string, sql string) error {
	script := fmt.Sprintf(
		`if command -v mysql >/dev/null 2>&1; then DB_CLI=mysql; elif command -v mariadb >/dev/null 2>&1; then DB_CLI=mariadb; else exit 1; fi && echo %s | "$DB_CLI" -u %s %s -f`,
		shellEscape(sql), shellEscape(dbUser), shellEscape(dbName),
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

func (o *OpenMageBootstrap) runComposerCommand(projectDir string, args ...string) error {
	cmd := exec.Command("composer", args...)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
