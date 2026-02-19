package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestApacheHTTPDConfigIncludesRequiredModules(t *testing.T) {
	content := readApacheHTTPDConfig(t)

	if !strings.Contains(content, "LoadModule unixd_module") {
		t.Fatalf("expected mod_unixd to be loaded in apache httpd.conf")
	}
	if !strings.Contains(content, "LoadModule version_module") {
		t.Fatalf("expected mod_version to be loaded in apache httpd.conf")
	}
}

func TestApacheHTTPDConfigAvoidsUnsupportedProxyPassEnvParameter(t *testing.T) {
	content := readApacheHTTPDConfig(t)
	if strings.Contains(content, "env=use_debug") {
		t.Fatalf("expected apache config to avoid unsupported env=use_debug ProxyPassMatch parameter")
	}
}

func TestApacheHTTPDConfigAllowsMagentoRootHtaccessOverrides(t *testing.T) {
	content := readApacheHTTPDConfig(t)
	if !strings.Contains(content, "<Directory \"/var/www/html\">") {
		t.Fatalf("expected apache config to allow overrides from Magento root directory")
	}
}

func readApacheHTTPDConfig(t *testing.T) string {
	t.Helper()

	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..")
	configPath := filepath.Join(projectRoot, "docker", "apache", "etc", "httpd.conf")

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read apache httpd.conf: %v", err)
	}
	return string(content)
}
