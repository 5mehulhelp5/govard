package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApacheTemplateHonorsForwardedHTTPS(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "docker", "apache", "etc", "httpd.conf"))
	if err != nil {
		t.Fatalf("read apache template: %v", err)
	}

	template := string(content)

	if !strings.Contains(template, "SetEnvIf X-Forwarded-Proto https HTTPS=on") {
		t.Fatal("apache template must map X-Forwarded-Proto=https to HTTPS=on for apps behind Caddy")
	}

	if !strings.Contains(template, `SetHandler "proxy:fcgi://php:9000"`) {
		t.Fatal("apache template must use SetHandler-based PHP-FPM proxying so per-directory .htaccess env is preserved")
	}

	if !strings.Contains(template, "AcceptPathInfo On") {
		t.Fatal("apache template must allow PATH_INFO for legacy index.php/ routes")
	}

	if strings.Contains(template, "[L,DPI]") {
		t.Fatal("apache template must not discard PATH_INFO when rewriting legacy index.php/ routes")
	}
}
