package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	ProjectRegistryPathEnvVar = "GOVARD_PROJECT_REGISTRY_PATH"
	projectRegistryVersion    = 1
)

type ProjectRegistryEntry struct {
	Path         string    `json:"path"`
	ProjectName  string    `json:"project_name"`
	Domain       string    `json:"domain,omitempty"`
	ExtraDomains []string  `json:"extra_domains,omitempty"`
	Framework    string    `json:"framework,omitempty"`
	LastSeenAt   time.Time `json:"last_seen_at"`
	LastCommand  string    `json:"last_command,omitempty"`
}

type projectRegistryDocument struct {
	Version  int                    `json:"version"`
	Projects []ProjectRegistryEntry `json:"projects"`
}

func ProjectRegistryPath() string {
	if override := strings.TrimSpace(os.Getenv(ProjectRegistryPathEnvVar)); override != "" {
		return filepath.Clean(override)
	}

	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return filepath.Join(".govard", "projects.json")
	}
	return filepath.Join(home, ".govard", "projects.json")
}

func ReadProjectRegistryEntries() ([]ProjectRegistryEntry, error) {
	path := ProjectRegistryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []ProjectRegistryEntry{}, nil
		}
		return nil, fmt.Errorf("read project registry %s: %w", path, err)
	}

	if strings.TrimSpace(string(data)) == "" {
		return []ProjectRegistryEntry{}, nil
	}

	doc := projectRegistryDocument{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse project registry %s: %w", path, err)
	}

	entries := make([]ProjectRegistryEntry, 0, len(doc.Projects))
	for _, entry := range doc.Projects {
		normalized, ok := normalizeProjectRegistryEntry(entry)
		if !ok {
			continue
		}
		entries = append(entries, normalized)
	}
	sortProjectRegistryEntries(entries)
	return entries, nil
}

func UpsertProjectRegistryEntry(entry ProjectRegistryEntry) error {
	normalized, ok := normalizeProjectRegistryEntry(entry)
	if !ok {
		return fmt.Errorf("project path is required")
	}

	entries, err := ReadProjectRegistryEntries()
	if err != nil {
		return err
	}

	replaced := false
	for index := range entries {
		if entries[index].Path == normalized.Path {
			entries[index] = normalized
			replaced = true
			break
		}
	}
	if !replaced {
		entries = append(entries, normalized)
	}

	sortProjectRegistryEntries(entries)
	if err := writeProjectRegistryEntries(ProjectRegistryPath(), entries); err != nil {
		return err
	}
	return nil
}

func normalizeProjectRegistryEntry(entry ProjectRegistryEntry) (ProjectRegistryEntry, bool) {
	entry.Path = strings.TrimSpace(entry.Path)
	if entry.Path == "" {
		return ProjectRegistryEntry{}, false
	}
	entry.Path = filepath.Clean(entry.Path)

	if os.Getenv(ProjectRegistryPathEnvVar) == "" {
		if strings.HasPrefix(entry.Path, "/tmp/") || strings.HasPrefix(entry.Path, filepath.Clean(os.TempDir())) || strings.Contains(entry.Path, "govard/tests") {
			return ProjectRegistryEntry{}, false
		}
	}

	entry.ProjectName = strings.TrimSpace(entry.ProjectName)
	entry.Domain = strings.TrimSpace(entry.Domain)
	for i, d := range entry.ExtraDomains {
		entry.ExtraDomains[i] = strings.TrimSpace(d)
	}
	entry.Framework = strings.TrimSpace(strings.ToLower(entry.Framework))
	entry.LastCommand = strings.TrimSpace(strings.ToLower(entry.LastCommand))
	if entry.LastSeenAt.IsZero() {
		entry.LastSeenAt = time.Now().UTC()
	} else {
		entry.LastSeenAt = entry.LastSeenAt.UTC()
	}
	if entry.ProjectName == "" {
		entry.ProjectName = filepath.Base(entry.Path)
	}
	return entry, true
}

func sortProjectRegistryEntries(entries []ProjectRegistryEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].LastSeenAt.Equal(entries[j].LastSeenAt) {
			return entries[i].Path < entries[j].Path
		}
		return entries[i].LastSeenAt.After(entries[j].LastSeenAt)
	})
}

func writeProjectRegistryEntries(path string, entries []ProjectRegistryEntry) error {
	doc := projectRegistryDocument{
		Version:  projectRegistryVersion,
		Projects: entries,
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal project registry: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create project registry dir %s: %w", dir, err)
	}

	tmpFile, err := os.CreateTemp(dir, "projects-*.tmp")
	if err != nil {
		return fmt.Errorf("create project registry temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	cleanup := func() {
		_ = os.Remove(tmpPath)
	}
	defer cleanup()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write project registry temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close project registry temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, 0644); err != nil {
		return fmt.Errorf("chmod project registry temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace project registry file %s: %w", path, err)
	}
	return nil
}
