package tests

import (
	"path/filepath"
	"testing"

	"govard/internal/engine/remote"
)

func TestAuthStoreRoundTrip(t *testing.T) {
	store := remote.NewInMemoryStore()
	if err := store.Set("staging", "secret"); err != nil {
		t.Fatal(err)
	}
	val, err := store.Get("staging")
	if err != nil {
		t.Fatal(err)
	}
	if val != "secret" {
		t.Fatalf("expected secret, got %s", val)
	}
}

func TestFileStoreRoundTrip(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "auth.json")
	t.Setenv("GOVARD_AUTH_STORE_PATH", storePath)

	store := remote.NewFileStore()
	if err := store.Set("staging", "token-123"); err != nil {
		t.Fatal(err)
	}

	val, err := store.Get("staging")
	if err != nil {
		t.Fatal(err)
	}
	if val != "token-123" {
		t.Fatalf("expected token-123, got %s", val)
	}
}

func TestKeychainStoreFallbackRoundTrip(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "auth.json")
	t.Setenv("GOVARD_AUTH_STORE_PATH", storePath)

	store := remote.NewKeychainStore()
	if err := store.Set("prod", "token-prod"); err != nil {
		t.Fatal(err)
	}

	val, err := store.Get("prod")
	if err != nil {
		t.Fatal(err)
	}
	if val != "token-prod" {
		t.Fatalf("expected token-prod, got %s", val)
	}
}
