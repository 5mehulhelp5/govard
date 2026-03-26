package engine

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func prepareWebServerConfigAssets(blueprintsFS fs.FS, data RenderData) (string, string, error) {
	if !frameworkUsesManagedWebServer(data.Config.Framework) {
		return "", "", nil
	}

	webServer := strings.ToLower(strings.TrimSpace(data.Config.Stack.Services.WebServer))
	if webServer == "" {
		webServer = strings.ToLower(strings.TrimSpace(data.Config.Stack.WebServer))
	}

	var nginxPath string
	var apachePath string

	switch webServer {
	case "nginx", "hybrid":
		path, err := prepareNginxConfigAsset(blueprintsFS, data, webServer == "hybrid")
		if err != nil {
			return "", "", err
		}
		nginxPath = path
	}

	switch webServer {
	case "apache", "hybrid":
		path, err := prepareApacheHTTPDConfigAsset(blueprintsFS, data)
		if err != nil {
			return "", "", err
		}
		apachePath = path
	}

	return nginxPath, apachePath, nil
}

func prepareNginxConfigAsset(blueprintsFS fs.FS, data RenderData, hybrid bool) (string, error) {
	if strings.TrimSpace(data.Config.ProjectName) == "" {
		return "", nil
	}

	templateName := data.NGINXTemplate
	if hybrid {
		templateName = "hybrid.conf"
	}
	if strings.TrimSpace(templateName) == "" {
		templateName = "default.conf"
	}

	templatePath := path.Join("support", "nginx", "templates", templateName)
	content, err := fs.ReadFile(blueprintsFS, templatePath)
	if err != nil {
		return "", fmt.Errorf("read nginx support template %s: %w", templatePath, err)
	}

	rendered := strings.NewReplacer(
		"${NGINX_PUBLIC}", data.NGINXPublic,
		"${XDEBUG_SESSION_PATTERN}", data.XdebugSessionPattern,
	).Replace(string(content))

	destPath := filepath.Join(GovardHomeDir(), "nginx", data.Config.ProjectName, "default.conf")
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(destPath, []byte(rendered), 0o644); err != nil {
		return "", err
	}
	return destPath, nil
}

func prepareApacheHTTPDConfigAsset(blueprintsFS fs.FS, data RenderData) (string, error) {
	if strings.TrimSpace(data.Config.ProjectName) == "" {
		return "", nil
	}

	templatePath := path.Join("support", "apache", "httpd.conf")
	content, err := fs.ReadFile(blueprintsFS, templatePath)
	if err != nil {
		return "", fmt.Errorf("read apache support template %s: %w", templatePath, err)
	}

	rendered := strings.NewReplacer(
		"@DOCROOT@", data.ApacheDocumentRoot,
		"@XDEBUG_SESSION@", data.XdebugSessionPattern,
	).Replace(string(content))

	destPath := filepath.Join(GovardHomeDir(), "apache", data.Config.ProjectName, "httpd.conf")
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(destPath, []byte(rendered), 0o644); err != nil {
		return "", err
	}
	if err := copyApacheSupportFile(blueprintsFS, "mime.types", filepath.Join(filepath.Dir(destPath), "mime.types")); err != nil {
		return "", err
	}
	if err := mirrorApacheRunMapIntoConfDir(destPath, data.ApacheMageRunMapPath); err != nil {
		return "", err
	}
	return destPath, nil
}

func buildContainerDocumentRoot(webRoot string) string {
	webRoot = strings.TrimSpace(webRoot)
	switch webRoot {
	case "", "/":
		return "/var/www/html/"
	default:
		if !strings.HasPrefix(webRoot, "/") {
			webRoot = "/" + webRoot
		}
		return "/var/www/html" + webRoot
	}
}

func frameworkUsesManagedWebServer(framework string) bool {
	fwConfig, ok := GetFrameworkConfig(framework)
	if !ok {
		return false
	}

	for _, include := range fwConfig.Includes {
		if include == "includes/base.yml" {
			return true
		}
	}

	return false
}

func mirrorApacheRunMapIntoConfDir(httpdConfigPath string, apacheRunMapPath string) error {
	if strings.TrimSpace(httpdConfigPath) == "" || strings.TrimSpace(apacheRunMapPath) == "" {
		return nil
	}

	content, err := os.ReadFile(apacheRunMapPath)
	if err != nil {
		return fmt.Errorf("read apache mage-run map %s: %w", apacheRunMapPath, err)
	}

	destPath := filepath.Join(filepath.Dir(httpdConfigPath), "extra", "mage-run-map.conf")
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(destPath, content, 0o644); err != nil {
		return err
	}

	return nil
}

func copyApacheSupportFile(blueprintsFS fs.FS, filename string, destPath string) error {
	content, err := fs.ReadFile(blueprintsFS, path.Join("support", "apache", filename))
	if err != nil {
		return fmt.Errorf("read apache support file %s: %w", filename, err)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(destPath, content, 0o644); err != nil {
		return err
	}
	return nil
}
