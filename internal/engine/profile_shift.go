package engine

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"govard/internal/conventions"

	"github.com/pterm/pterm"
)

// ProfileShiftInfo holds structured data about a detected environment profile change.
type ProfileShiftInfo struct {
	Shifted         bool
	Reason          string
	PreviousPHP     string
	CurrentPHP      string
	PreviousProfile string
	CurrentProfile  string
	PreviousVersion string
	CurrentVersion  string
	IsInitial       bool
}

// DetectProfileShift detects whether the current config represents a
// profile change compared to the last known state (lock file or registry).
// This is intended to be called early in the pipeline, before containers start.
func DetectProfileShift(config Config) ProfileShiftInfo {
	shifted, reason := checkProfileShiftCleanup(config)
	if !shifted {
		return ProfileShiftInfo{}
	}

	cwd, _ := os.Getwd()
	previousPHP, previousProfile, previousVersion := "", "", ""
	isInitial := false

	lockFile, err := ReadLockFile(LockFilePath(cwd))
	if err == nil {
		previousPHP = strings.TrimSpace(lockFile.Stack.PHPVersion)
		previousProfile = strings.TrimSpace(lockFile.Project.Profile)
		previousVersion = strings.TrimSpace(lockFile.Project.FrameworkVersion)
	} else if entry, ok := GetProjectRegistryEntry(cwd); ok {
		previousPHP = strings.TrimSpace(entry.PHPVersion)
		previousProfile = strings.TrimSpace(entry.Profile)
		previousVersion = strings.TrimSpace(entry.FrameworkVersion)
	} else {
		isInitial = true
	}

	return ProfileShiftInfo{
		Shifted:         true,
		Reason:          reason,
		PreviousPHP:     previousPHP,
		CurrentPHP:      strings.TrimSpace(config.Stack.PHPVersion),
		PreviousProfile: previousProfile,
		CurrentProfile:  strings.TrimSpace(config.Profile),
		PreviousVersion: previousVersion,
		CurrentVersion:  strings.TrimSpace(config.FrameworkVersion),
		IsInitial:       isInitial,
	}
}

// PrepareInfraForShift handles infrastructure cleanup BEFORE containers
// are started. This must be called before `docker compose up` when a profile
// shift is detected, to avoid issues like Redis RDB version incompatibility.
func PrepareInfraForShift(projectName string, config Config) {
	// Force remove Redis/Valkey container to avoid RDB version conflicts (e.g. 7.2 -> 7.0)
	if config.Stack.Features.Cache || config.Stack.Services.Cache != "none" {
		redisContainer := fmt.Sprintf("%s%s", projectName, conventions.RedisSuffix)
		pterm.Info.Println("Removing stale cache container for clean profile start...")
		_ = exec.Command("docker", "rm", "-f", redisContainer).Run()
	}
}

func checkProfileShiftCleanup(config Config) (bool, string) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, ""
	}

	previousPHP := ""
	previousProfile := ""
	previousVersion := ""
	foundPrevious := false

	lockFile, err := ReadLockFile(LockFilePath(cwd))
	if err == nil {
		previousPHP = strings.TrimSpace(lockFile.Stack.PHPVersion)
		previousProfile = strings.TrimSpace(lockFile.Project.Profile)
		previousVersion = strings.TrimSpace(lockFile.Project.FrameworkVersion)
		foundPrevious = true
	} else {
		// Fallback to project registry
		if entry, ok := GetProjectRegistryEntry(cwd); ok {
			previousPHP = strings.TrimSpace(entry.PHPVersion)
			previousProfile = strings.TrimSpace(entry.Profile)
			previousVersion = strings.TrimSpace(entry.FrameworkVersion)
			foundPrevious = true
		}
	}

	currentPHP := strings.TrimSpace(config.Stack.PHPVersion)
	currentProfile := strings.TrimSpace(config.Profile)
	currentVersion := strings.TrimSpace(config.FrameworkVersion)

	if !foundPrevious {
		return true, "Initial configuration"
	}

	if previousPHP != "" && currentPHP != "" && previousPHP != currentPHP {
		return true, fmt.Sprintf("PHP version changed: %s -> %s", previousPHP, currentPHP)
	}

	if previousProfile != currentProfile {
		return true, fmt.Sprintf("Profile changed: %q -> %q", previousProfile, currentProfile)
	}

	if previousVersion != "" && currentVersion != "" && previousVersion != currentVersion {
		return true, fmt.Sprintf("Version changed: %s -> %s", previousVersion, currentVersion)
	}

	return false, ""
}
