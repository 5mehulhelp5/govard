package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProxyBlueprintContainsCaddyResumeFlag(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "internal", "blueprints", "files", "proxy.yml"))
	if err != nil {
		t.Fatalf("read proxy blueprint: %v", err)
	}

	if !strings.Contains(string(content), "--resume") {
		t.Fatal("proxy.yml must contain --resume flag for Caddy to persist config across restarts")
	}
}

func TestProxyBlueprintPublishesSearchPort(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "internal", "blueprints", "files", "proxy.yml"))
	if err != nil {
		t.Fatalf("read proxy blueprint: %v", err)
	}

	if !strings.Contains(string(content), `"9200:9200"`) {
		t.Fatal("proxy.yml must publish port 9200 so project.test:9200 can reach a project's search engine")
	}
}
