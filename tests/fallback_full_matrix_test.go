package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"govard/internal/engine"
)

func TestLocalImageFallbackFullMatrix(t *testing.T) {
	// 1. Create a realistic "Full Matrix" docker-compose.yml
	composeContent := `
services:
  apache:
    image: ddtcorex/govard-apache:2.4
  nginx:
    image: ddtcorex/govard-nginx:1.26
  php:
    image: ddtcorex/govard-php:8.2
  php-debug:
    image: ddtcorex/govard-php:8.2-debug
  magento2:
    image: ddtcorex/govard-php-magento2:8.2
  magento2-debug:
    image: ddtcorex/govard-php-magento2:8.2-debug
  db-mariadb:
    image: ddtcorex/govard-mariadb:10.11
  db-mysql:
    image: ddtcorex/govard-mysql:8.0
  cache-redis:
    image: ddtcorex/govard-redis:7.0
  cache-valkey:
    image: ddtcorex/govard-valkey:7.2
  mq:
    image: ddtcorex/govard-rabbitmq:3.11
  search-opensearch:
    image: ddtcorex/govard-opensearch:2.5
  search-elasticsearch:
    image: ddtcorex/govard-elasticsearch:7.17
  proxy-varnish:
    image: ddtcorex/govard-varnish:7.0
  dns:
    image: ddtcorex/govard-dnsmasq:latest
`
	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(composeContent), 0644); err != nil {
		t.Fatalf("failed to write test compose file: %v", err)
	}

	// 2. We want to test the DISCOVERY and SPEC resolution for ALL services.
	// Since engine.FallbackBuildMissingGovardImagesFromCompose actually TRYS to run 'docker build',
	// we will verify each service using ResolveLocalBuildSpecForTest which is exported for tests.

	services := []struct {
		name     string
		tag      string
		repo     string
		expected ContextSpec
	}{
		{"apache", "2.4", "ddtcorex/govard-", ContextSpec{Context: "apache", Arg: "APACHE_VERSION", Val: "2.4.66"}},
		{"nginx", "1.26", "ddtcorex/govard-", ContextSpec{Context: "nginx", Arg: "NGINX_VERSION", Val: "1.26.0"}},
		{"php", "8.2", "ddtcorex/govard-", ContextSpec{Context: "php", Arg: "PHP_VERSION", Val: "8.2"}},
		{"php", "8.2-debug", "ddtcorex/govard-", ContextSpec{Context: "php", Dockerfile: "php/debug/Dockerfile"}},
		{"php-magento2", "8.2", "ddtcorex/govard-", ContextSpec{Context: "php", Dockerfile: "php/magento2/Dockerfile"}},
		{"php-magento2", "8.2-debug", "ddtcorex/govard-", ContextSpec{Context: "php", Dockerfile: "php/debug/Dockerfile"}},
		{"mariadb", "10.11", "ddtcorex/govard-", ContextSpec{Context: "mariadb", Arg: "MARIADB_VERSION", Val: "10.11"}},
		{"mysql", "8.0", "ddtcorex/govard-", ContextSpec{Context: "mysql", Arg: "MYSQL_VERSION", Val: "8.0"}},
		{"redis", "7.0", "ddtcorex/govard-", ContextSpec{Context: "redis", Arg: "REDIS_VERSION", Val: "7.0"}},
		{"valkey", "7.2", "ddtcorex/govard-", ContextSpec{Context: "valkey", Arg: "VALKEY_VERSION", Val: "7.2"}},
		{"rabbitmq", "3.11", "ddtcorex/govard-", ContextSpec{Context: "rabbitmq", Arg: "RABBITMQ_VERSION", Val: "3.11"}},
		{"opensearch", "2.5", "ddtcorex/govard-", ContextSpec{Context: "opensearch", Arg: "OPENSEARCH_VERSION", Val: "2.5"}},
		{"elasticsearch", "7.17", "ddtcorex/govard-", ContextSpec{Context: "elasticsearch", Arg: "ELASTICSEARCH_VERSION", Val: "7.17"}},
		{"varnish", "7.0", "ddtcorex/govard-", ContextSpec{Context: "varnish", Arg: "VARNISH_VERSION", Val: "7.0"}},
		{"dnsmasq", "latest", "ddtcorex/govard-", ContextSpec{Context: "dnsmasq"}},
	}

	for _, s := range services {
		t.Run(s.name+":"+s.tag, func(t *testing.T) {
			spec, err := engine.ResolveLocalBuildSpecForTest(s.name, s.tag, s.repo)
			if err != nil {
				t.Fatalf("failed to resolve build spec: %v", err)
			}

			if spec.ContextRel != s.expected.Context {
				t.Errorf("expected context %q, got %q", s.expected.Context, spec.ContextRel)
			}
			if s.expected.Dockerfile != "" && spec.DockerfileRel != s.expected.Dockerfile {
				t.Errorf("expected dockerfile %q, got %q", s.expected.Dockerfile, spec.DockerfileRel)
			}
			if s.expected.Arg != "" {
				if val, ok := spec.BuildArgs[s.expected.Arg]; !ok || val != s.expected.Val {
					t.Errorf("expected build arg %s=%s, got %q", s.expected.Arg, s.expected.Val, val)
				}
			}
		})
	}

	// 3. Verify images are extracted from the compose file correctly.
	images, err := engine.ReadServiceImagesFromCompose(composePath)
	if err != nil {
		t.Fatalf("failed to read images from compose: %v", err)
	}

	expectedImageCount := 15 // Number of services in the string above
	if len(images) != expectedImageCount {
		t.Errorf("expected %d images, got %d: %v", expectedImageCount, len(images), images)
	}

	// Double check some key image strings
	foundMagento := false
	for _, img := range images {
		if strings.Contains(img, "php-magento2:8.2-debug") {
			foundMagento = true
			break
		}
	}
	if !foundMagento {
		t.Error("failed to find magento2-debug image in extracted images")
	}
}

type ContextSpec struct {
	Context    string
	Dockerfile string
	Arg        string
	Val        string
}
