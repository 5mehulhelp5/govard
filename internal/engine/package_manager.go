package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func ResolveNodePackageManager(root string) string {
	if pm, ok := readDeclaredNodePackageManager(filepath.Join(root, "package.json")); ok {
		return pm
	}

	type candidate struct {
		file string
		pm   string
	}

	for _, candidate := range []candidate{
		{file: "pnpm-workspace.yaml", pm: "pnpm"},
		{file: "pnpm-lock.yaml", pm: "pnpm"},
		{file: "yarn.lock", pm: "yarn"},
		{file: "bun.lock", pm: "bun"},
		{file: "bun.lockb", pm: "bun"},
		{file: "package-lock.json", pm: "npm"},
	} {
		if _, err := os.Stat(filepath.Join(root, candidate.file)); err == nil {
			return candidate.pm
		}
	}

	return "npm"
}

func readDeclaredNodePackageManager(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	var pkg struct {
		PackageManager string `json:"packageManager"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", false
	}

	raw := strings.TrimSpace(pkg.PackageManager)
	if raw == "" {
		return "", false
	}

	for _, supported := range []string{"pnpm", "yarn", "bun", "npm"} {
		if raw == supported || strings.HasPrefix(raw, supported+"@") {
			return supported, true
		}
	}

	return "", false
}
