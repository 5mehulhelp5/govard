package tests

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"govard/internal/engine"
)

func TestBlueprintRegistryRequiresTrustOptIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := engine.ResolveBlueprintRegistryForTest(engine.BlueprintRegistryConfig{
		Provider: "http",
		URL:      "https://example.com/blueprints.tar.gz",
		Checksum: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if err == nil {
		t.Fatal("expected trust opt-in error")
	}
}

func TestBlueprintRegistryRequiresChecksum(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := engine.ResolveBlueprintRegistryForTest(engine.BlueprintRegistryConfig{
		Provider: "http",
		URL:      "https://example.com/blueprints.tar.gz",
		Trusted:  true,
	})
	if err == nil {
		t.Fatal("expected missing checksum error")
	}
}

func TestBlueprintRegistryCachesHTTPArchive(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	archive := mustBuildTarGz(t, map[string]string{
		"blueprints/legacytest.tmpl": "services:\n  app:\n    image: alpine:3.20\n",
	})
	checksum := sha256Hex(archive)

	var requests int
	restore := engine.SetBlueprintRegistryHTTPClientForTest(&http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			requests++
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"application/gzip"},
				},
				Body: io.NopCloser(bytes.NewReader(archive)),
			}, nil
		}),
	})
	defer restore()

	cfg := engine.BlueprintRegistryConfig{
		Provider: "http",
		URL:      "https://example.com/blueprints.tar.gz",
		Checksum: checksum,
		Trusted:  true,
	}

	firstPath, err := engine.ResolveBlueprintRegistryForTest(cfg)
	if err != nil {
		t.Fatalf("resolve first fetch: %v", err)
	}
	if _, err := os.Stat(filepath.Join(firstPath, "legacytest.tmpl")); err != nil {
		t.Fatalf("expected cached blueprint template file: %v", err)
	}

	secondPath, err := engine.ResolveBlueprintRegistryForTest(cfg)
	if err != nil {
		t.Fatalf("resolve cached fetch: %v", err)
	}
	if firstPath != secondPath {
		t.Fatalf("expected stable cache path, got %s then %s", firstPath, secondPath)
	}
	if requests != 1 {
		t.Fatalf("expected one HTTP request due to cache reuse, got %d", requests)
	}
}

func TestBlueprintRegistryRejectsChecksumMismatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	archive := mustBuildTarGz(t, map[string]string{
		"blueprints/legacytest.tmpl": "services:\n  app:\n    image: alpine:3.20\n",
	})

	restore := engine.SetBlueprintRegistryHTTPClientForTest(&http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"application/gzip"},
				},
				Body: io.NopCloser(bytes.NewReader(archive)),
			}, nil
		}),
	})
	defer restore()

	_, err := engine.ResolveBlueprintRegistryForTest(engine.BlueprintRegistryConfig{
		Provider: "http",
		URL:      "https://example.com/blueprints.tar.gz",
		Checksum: strings.Repeat("0", 64),
		Trusted:  true,
	})
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch error, got %v", err)
	}
}

func mustBuildTarGz(t *testing.T, files map[string]string) []byte {
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

func sha256Hex(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
