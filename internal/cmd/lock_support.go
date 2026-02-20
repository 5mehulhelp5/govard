package cmd

import (
	"fmt"
	"os"

	"govard/internal/engine"
)

func evaluateUpLockWarnings(cwd string, config engine.Config) []string {
	lockPath := engine.LockFilePath(cwd)
	expected, err := engine.ReadLockFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return []string{fmt.Sprintf("Lock check skipped: %v", err)}
	}

	current, err := engine.BuildLockFileFromConfig(cwd, config, Version, lockDependencies)
	if err != nil {
		return []string{fmt.Sprintf("Lock check skipped: %v", err)}
	}
	return buildUpLockWarnings(expected, current)
}

func buildUpLockWarnings(expected engine.LockFile, current engine.LockFile) []string {
	result := engine.CompareLockFile(expected, current)
	if result.Compliant {
		return nil
	}
	warnings := make([]string, 0, len(result.Mismatches))
	for _, mismatch := range result.Mismatches {
		warnings = append(warnings, "Lockfile mismatch: "+mismatch)
	}
	return warnings
}

// BuildUpLockWarningsForTest exposes lock warning rendering for tests.
func BuildUpLockWarningsForTest(expected engine.LockFile, current engine.LockFile) []string {
	return buildUpLockWarnings(expected, current)
}
