package tests

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"govard/internal/engine"

	"gopkg.in/yaml.v3"
)

func TestRenderMagento2Blueprint(t *testing.T) {
	testBlueprintRender(t, "magento2", []string{
		"image: ddtcorex/govard-nginx:1.28",
		"image: ddtcorex/govard-php-magento2:",
		"image: ddtcorex/govard-mariadb:",
		"MYSQL_DATABASE: magento",
	})

	// Regression: web service must be connected to govard-proxy so Caddy can route to it.
	// Without this, `govard env up` succeeds but the site returns 502 Bad Gateway.
	tempDir := t.TempDir()
	setTestGovardHome(t, tempDir)
	t.Setenv("GOVARD_BLUEPRINTS_DIR", func() string {
		_, filename, _, _ := runtime.Caller(0)
		return filepath.Join(filepath.Dir(filename), "..", "internal", "blueprints", "files")
	}())

	config := engine.Config{
		ProjectName: "sample-project",
		Framework:   "magento2",
		Domain:      "sample-project.test",
		Stack:       engine.Stack{PHPVersion: "8.3"},
	}
	if err := engine.RenderBlueprint(tempDir, config); err != nil {
		t.Fatalf("render failed: %v", err)
	}
	content, err := os.ReadFile(engine.ComposeFilePath(tempDir, config.ProjectName))
	if err != nil {
		t.Fatalf("read compose file: %v", err)
	}

	var composeStruct struct {
		Services map[string]struct {
			Networks interface{} `yaml:"networks"`
		} `yaml:"services"`
	}
	if err := yaml.Unmarshal(content, &composeStruct); err != nil {
		t.Fatalf("failed to parse yaml: %v", err)
	}
	webSvc, ok := composeStruct.Services["web"]
	if !ok {
		t.Fatal("web service not found in parsed compose file")
	}

	hasProxy := false
	if networks, ok := webSvc.Networks.([]interface{}); ok {
		for _, n := range networks {
			if n == "govard-proxy" {
				hasProxy = true
				break
			}
		}
	} else if networks, ok := webSvc.Networks.(map[string]interface{}); ok {
		if _, ok := networks["govard-proxy"]; ok {
			hasProxy = true
		}
	}
	if !hasProxy {
		t.Errorf("web service must be connected to govard-proxy network (required for Caddy routing). Networks: %v", webSvc.Networks)
	}
}

func TestRenderLaravelBlueprint(t *testing.T) {
	testBlueprintRender(t, "laravel", []string{
		"image: ddtcorex/govard-nginx:1.28",
		"image: ddtcorex/govard-php:",
		"image: ddtcorex/govard-mariadb:",
		"MYSQL_DATABASE: laravel",
		"queue:",
		"php artisan queue:work",
	})
}

func TestRenderNextjsBlueprint(t *testing.T) {
	testBlueprintRender(t, "nextjs", []string{
		"image: node:24-alpine",
		"working_dir: /app",
		"command: npm run dev -- --hostname 0.0.0.0 --port 80",
	})
}

func TestRenderMagento1Blueprint(t *testing.T) {
	testBlueprintRender(t, "magento1", []string{
		"image: ddtcorex/govard-nginx:1.28",
		"image: ddtcorex/govard-php:",
		"image: ddtcorex/govard-mariadb:",
		"MYSQL_DATABASE: magento",
	})
}

func TestRenderDrupalBlueprint(t *testing.T) {
	testBlueprintRender(t, "drupal", []string{
		"image: ddtcorex/govard-nginx:1.28",
		"image: ddtcorex/govard-php:",
		"image: ddtcorex/govard-mariadb:",
		"MYSQL_DATABASE: drupal",
	})
}

func TestRenderSymfonyBlueprint(t *testing.T) {
	testBlueprintRender(t, "symfony", []string{
		"image: ddtcorex/govard-nginx:1.28",
		"image: ddtcorex/govard-php:",
		"image: ddtcorex/govard-mariadb:",
		"MYSQL_DATABASE: symfony",
	})
}

func TestRenderShopwareBlueprint(t *testing.T) {
	testBlueprintRender(t, "shopware", []string{
		"image: ddtcorex/govard-nginx:1.28",
		"image: ddtcorex/govard-php:",
		"image: ddtcorex/govard-mariadb:",
		"MYSQL_DATABASE: shopware",
	})
}

func TestRenderCakephpBlueprint(t *testing.T) {
	testBlueprintRender(t, "cakephp", []string{
		"image: ddtcorex/govard-nginx:1.28",
		"image: ddtcorex/govard-php:",
		"image: ddtcorex/govard-mariadb:",
		"MYSQL_DATABASE: cakephp",
	})
}

func TestRenderWordpressBlueprint(t *testing.T) {
	testBlueprintRender(t, "wordpress", []string{
		"image: ddtcorex/govard-nginx:1.28",
		"image: ddtcorex/govard-php:",
		"image: ddtcorex/govard-mariadb:",
		"MYSQL_DATABASE: wordpress",
	})
}

func TestRenderCustomBlueprint(t *testing.T) {
	testBlueprintRender(t, "custom", []string{
		"image: ddtcorex/govard-nginx:1.28",
		"image: ddtcorex/govard-php:",
		"image: ddtcorex/govard-mariadb:",
		"MYSQL_DATABASE: app",
	})
}

func TestRenderBlueprintWithFeatures(t *testing.T) {
	tempDir := t.TempDir()
	setTestGovardHome(t, tempDir)

	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..")
	blueprintsDir := filepath.Join(projectRoot, "internal", "blueprints", "files")

	destBlueprintsDir := filepath.Join(tempDir, "blueprints")
	if err := copyDir(blueprintsDir, destBlueprintsDir); err != nil {
		t.Fatalf("Failed to copy blueprints: %v", err)
	}

	config := engine.Config{
		ProjectName: "test-magento-full",
		Framework:   "magento2",
		Domain:      "magento.test",
		Stack: engine.Stack{
			PHPVersion: "8.3",
			DBType:     "mariadb",
			DBVersion:  "10.6",
			WebServer:  "nginx",
			Features: engine.Features{
				Xdebug:        true,
				Varnish:       true,
				Redis:         true,
				Elasticsearch: true,
			},
		},
	}

	err := engine.RenderBlueprint(tempDir, config)
	if err != nil {
		t.Fatalf("Failed to render blueprint with features: %v", err)
	}

	content, err := os.ReadFile(engine.ComposeFilePath(tempDir, config.ProjectName))
	if err != nil {
		t.Fatalf("Failed to read generated compose file: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "XDEBUG_MODE: debug") {
		t.Error("Expected Xdebug to be enabled")
	}
	if !strings.Contains(contentStr, "varnish:") {
		t.Error("Expected varnish service")
	}
	if !strings.Contains(contentStr, "redis:") {
		t.Error("Expected redis service")
	}
	if !strings.Contains(contentStr, "elasticsearch:") {
		t.Error("Expected elasticsearch service")
	}
	if !strings.Contains(contentStr, "aliases:") || !strings.Contains(contentStr, "- opensearch") {
		t.Error("Expected opensearch alias for elasticsearch service")
	}
}

func TestRenderMagento2BlueprintHybridWebServer(t *testing.T) {
	tempDir := t.TempDir()
	setTestGovardHome(t, tempDir)

	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..")
	blueprintsDir := filepath.Join(projectRoot, "internal", "blueprints", "files")

	destBlueprintsDir := filepath.Join(tempDir, "blueprints")
	if err := copyDir(blueprintsDir, destBlueprintsDir); err != nil {
		t.Fatalf("Failed to copy blueprints: %v", err)
	}

	config := engine.Config{
		ProjectName: "test-magento-hybrid",
		Framework:   "magento2",
		Domain:      "magento-hybrid.test",
		Stack: engine.Stack{
			PHPVersion: "8.4",
			DBType:     "mariadb",
			DBVersion:  "11.4",
			Services: engine.Services{
				WebServer: "hybrid",
				Search:    "none",
				Cache:     "none",
				Queue:     "none",
			},
		},
	}

	err := engine.RenderBlueprint(tempDir, config)
	if err != nil {
		t.Fatalf("Failed to render blueprint: %v", err)
	}

	content, err := os.ReadFile(engine.ComposeFilePath(tempDir, config.ProjectName))
	if err != nil {
		t.Fatalf("Failed to read generated compose file: %v", err)
	}
	contentStr := string(content)

	if !strings.Contains(contentStr, "web:") || !strings.Contains(contentStr, "- apache") {
		t.Fatalf("expected web service to depend on apache in hybrid mode, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "NGINX_TEMPLATE=hybrid.conf") {
		t.Fatalf("expected hybrid nginx template, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "apache:") || !strings.Contains(contentStr, "APACHE_DOCUMENT_ROOT=/var/www/html/pub") {
		t.Fatalf("expected apache sidecar service in hybrid mode, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "image: ddtcorex/govard-apache:2.4") {
		t.Fatalf("expected apache image in hybrid mode, got:\n%s", contentStr)
	}
}

func testBlueprintRender(t *testing.T, framework string, expectedStrings []string) {
	t.Helper()

	tempDir := t.TempDir()
	setTestGovardHome(t, tempDir)

	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..")
	blueprintsDir := filepath.Join(projectRoot, "internal", "blueprints", "files")

	destBlueprintsDir := filepath.Join(tempDir, "blueprints")
	if err := copyDir(blueprintsDir, destBlueprintsDir); err != nil {
		t.Fatalf("Failed to copy blueprints: %v", err)
	}

	config := engine.Config{
		ProjectName: "test-" + framework,
		Framework:   framework,
		Domain:      framework + ".test",
		Stack: engine.Stack{
			PHPVersion:  "8.3",
			NodeVersion: "24",
			DBType:      "mariadb",
			DBVersion:   "10.6",
			WebServer:   "nginx",
			Features: engine.Features{
				Xdebug:        false,
				Varnish:       false,
				Redis:         false,
				Elasticsearch: false,
			},
		},
	}

	err := engine.RenderBlueprint(tempDir, config)
	if err != nil {
		t.Fatalf("Failed to render %s blueprint: %v", framework, err)
	}

	outputPath := engine.ComposeFilePath(tempDir, config.ProjectName)
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated compose file: %v", err)
	}

	contentStr := string(content)

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected '%s' to be in generated compose file for %s", expected, framework)
		}
	}

	if !strings.Contains(contentStr, "govard-net:") {
		t.Errorf("Expected govard-net network in %s compose file", framework)
	}
	if !strings.Contains(contentStr, "govard-proxy:") {
		t.Errorf("Expected govard-proxy network in %s compose file", framework)
	}
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func setTestGovardHome(t *testing.T, root string) string {
	t.Helper()

	homeDir := filepath.Join(root, ".govard-home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("create test govard home: %v", err)
	}
	t.Setenv("GOVARD_HOME_DIR", homeDir)
	return homeDir
}
