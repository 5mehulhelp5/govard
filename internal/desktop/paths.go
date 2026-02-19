package desktop

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func FindRepoRoot() (string, error) {
	start, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if root, ok := findRootFrom(start); ok {
		return root, nil
	}

	if exe, err := os.Executable(); err == nil {
		if root, ok := findRootFrom(filepath.Dir(exe)); ok {
			return root, nil
		}
	}

	return "", fmt.Errorf("could not locate repository root from %s", start)
}

func ResolveAssets() (fs.FS, error) {
	root, err := FindRepoRoot()
	if err != nil {
		return nil, err
	}

	candidates := []string{
		filepath.Join(root, "desktop", "frontend", "dist"),
		filepath.Join(root, "desktop", "frontend"),
	}

	for _, candidate := range candidates {
		if isDir(candidate) {
			return os.DirFS(candidate), nil
		}
	}

	return nil, fmt.Errorf("desktop frontend assets not found under %s", filepath.Join(root, "desktop", "frontend"))
}

func exists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func findRootFrom(start string) (string, bool) {
	dir := start
	for {
		if exists(filepath.Join(dir, "go.mod")) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}
