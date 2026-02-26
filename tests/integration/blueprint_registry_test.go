//go:build integration
// +build integration

package integration

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"govard/internal/engine"
)

func TestRenderBlueprintWithRemoteRegistryArchive(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	archive := mustBuildTarGzArchive(t, map[string]string{
		"blueprints/legacytest.tmpl": "services:\n  app:\n    image: alpine:3.20\n",
	})
	checksum := sha256HexArchive(archive)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	projectDir := t.TempDir()
	config := engine.Config{
		ProjectName: "registry-demo",
		Framework:   "legacytest",
		Domain:      "registry-demo.test",
		BlueprintRegistry: engine.BlueprintRegistryConfig{
			Provider: "http",
			URL:      server.URL + "/blueprints.tar.gz",
			Checksum: checksum,
			Trusted:  true,
		},
	}

	if err := engine.RenderBlueprint(projectDir, config); err != nil {
		t.Fatalf("render blueprint with registry: %v", err)
	}

	composePath := engine.ComposeFilePath(projectDir, config.ProjectName)
	data, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("read rendered compose file: %v", err)
	}

	if !strings.Contains(string(data), "alpine:3.20") {
		t.Fatalf("expected rendered compose to include archive blueprint content")
	}
}

func mustBuildTarGzArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)

	for path, content := range files {
		header := &tar.Header{
			Name:    path,
			Mode:    0o644,
			Size:    int64(len(content)),
			ModTime: time.Unix(0, 0),
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("write tar header %s: %v", path, err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("write tar content %s: %v", path, err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}

	return archive.Bytes()
}

func sha256HexArchive(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
