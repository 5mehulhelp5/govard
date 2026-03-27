package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CleanupStaleComposeFiles removes compose files in ~/.govard/compose that haven't been modified recently.
// It returns the number of files removed and any error encountered.
func CleanupStaleComposeFiles(maxAge time.Duration) (int, error) {
	composeDir := filepath.Join(GovardHomeDir(), "compose")
	f, err := os.Open(composeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("open compose directory: %w", err)
	}
	defer f.Close()

	removedCount := 0
	now := time.Now()

	for {
		entries, err := f.ReadDir(100)
		if err != nil && err != io.EOF {
			return removedCount, fmt.Errorf("read compose directory: %w", err)
		}
		if len(entries) == 0 {
			break
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yml.hash") {
				continue
			}

			filePath := filepath.Join(composeDir, name)
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if now.Sub(info.ModTime()) > maxAge {
				if err := os.Remove(filePath); err == nil {
					removedCount++
				}
			}
		}
	}

	return removedCount, nil
}

// AutoCleanupComposeFiles triggers a background cleanup if needed (e.g., once a day).
func AutoCleanupComposeFiles() {
	// If the directory doesn't exist, there's nothing to cleanup.
	// We don't want to create it here as that would have side effects for diagnostics.
	composeDir := filepath.Join(GovardHomeDir(), "compose")
	if info, err := os.Stat(composeDir); err != nil || !info.IsDir() {
		return
	}

	lastCleanupFile := filepath.Join(composeDir, ".last_cleanup")
	if info, err := os.Stat(lastCleanupFile); err == nil {
		if time.Since(info.ModTime()) < 24*time.Hour {
			return
		}
	}

	// Create/touch the file
	_ = os.WriteFile(lastCleanupFile, []byte(time.Now().String()), 0600)

	// Run in background
	go func() {
		defer func() {
			if r := recover(); r != nil {
				_ = r // satisfy linter for empty branch in background recovery
			}
		}()
		_, _ = CleanupStaleComposeFiles(14 * 24 * time.Hour)
	}()
}

// CheckComposeSpam returns an error if there are too many files in the compose directory.
func CheckComposeSpam(threshold int) error {
	composeDir := filepath.Join(GovardHomeDir(), "compose")
	f, err := os.Open(composeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	// Use Readdirnames with a limit to avoid loading everything into memory if it's huge
	names, err := f.Readdirnames(threshold + 1)
	if err != nil && err != io.EOF {
		return err
	}

	if len(names) > threshold {
		return fmt.Errorf("directory has more than %d files", threshold)
	}

	return nil
}
