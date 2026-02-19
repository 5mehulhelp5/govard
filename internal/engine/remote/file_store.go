package remote

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type fileStore struct {
	path string
}

func NewFileStore() AuthStore {
	if override := os.Getenv("GOVARD_AUTH_STORE_PATH"); override != "" {
		return &fileStore{path: override}
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	return &fileStore{
		path: filepath.Join(home, ".govard", "auth.json"),
	}
}

func (f *fileStore) Get(key string) (string, error) {
	data, err := f.load()
	if err != nil {
		return "", err
	}
	return data[key], nil
}

func (f *fileStore) Set(key, value string) error {
	data, err := f.load()
	if err != nil {
		return err
	}
	data[key] = value
	return f.save(data)
}

func (f *fileStore) load() (map[string]string, error) {
	payload, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("read auth store: %w", err)
	}

	var entries map[string]string
	if err := json.Unmarshal(payload, &entries); err != nil {
		return nil, fmt.Errorf("parse auth store: %w", err)
	}
	if entries == nil {
		entries = map[string]string{}
	}
	return entries, nil
}

func (f *fileStore) save(entries map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(f.path), 0700); err != nil {
		return fmt.Errorf("create auth store dir: %w", err)
	}

	payload, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("encode auth store: %w", err)
	}

	tempPath := f.path + ".tmp"
	if err := os.WriteFile(tempPath, payload, 0600); err != nil {
		return fmt.Errorf("write auth store temp file: %w", err)
	}
	if err := os.Rename(tempPath, f.path); err != nil {
		return fmt.Errorf("replace auth store: %w", err)
	}
	return nil
}
