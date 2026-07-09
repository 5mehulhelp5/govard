package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestElasticsearchBlueprintJoinsProxyNetwork(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "internal", "blueprints", "files", "includes", "elasticsearch.yml"))
	if err != nil {
		t.Fatalf("read elasticsearch blueprint: %v", err)
	}

	if !strings.Contains(string(content), "govard-proxy") {
		t.Fatal("elasticsearch.yml must join the govard-proxy network so Caddy can reach it from the host")
	}
	if !strings.Contains(string(content), "opensearch") {
		t.Fatal("elasticsearch.yml must keep the opensearch network alias")
	}
}
