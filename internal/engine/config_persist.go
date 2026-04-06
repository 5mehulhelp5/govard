package engine

import (
	"os"
	"strings"
)

// PrepareConfigForWrite removes runtime-derived values so persisted config stays portable.
func PrepareConfigForWrite(config Config) Config {
	writable := config

	defaultXdebugSession := "PHPSTORM"
	if profileResult, err := ResolveRuntimeProfile(writable.Framework, writable.FrameworkVersion); err == nil {
		candidate := strings.TrimSpace(profileResult.Profile.XdebugSession)
		if candidate != "" {
			defaultXdebugSession = candidate
		}
	}
	if strings.EqualFold(strings.TrimSpace(writable.Stack.XdebugSession), strings.TrimSpace(defaultXdebugSession)) {
		writable.Stack.XdebugSession = ""
	}
	if strings.EqualFold(strings.TrimSpace(writable.Stack.WebServer), strings.TrimSpace(writable.Stack.Services.WebServer)) {
		writable.Stack.WebServer = ""
	}

	// Strip web server version for the server not in use.
	// Hybrid mode uses both nginx and apache, so both versions are kept.
	activeWebServer := strings.ToLower(strings.TrimSpace(writable.Stack.Services.WebServer))
	if activeWebServer != "" && activeWebServer != "hybrid" {
		if activeWebServer != "nginx" {
			writable.Stack.NginxVersion = ""
		}
		if activeWebServer != "apache" {
			writable.Stack.ApacheVersion = ""
		}
	}

	if writable.Stack.Services.Cache == "none" {
		writable.Stack.Services.Cache = ""
	}
	if writable.Stack.Services.Search == "none" {
		writable.Stack.Services.Search = ""
	}
	if writable.Stack.Services.Queue == "none" {
		writable.Stack.Services.Queue = ""
	}

	if writable.Stack.Services.DB == "none" {
		writable.Stack.Services.DB = ""
	}

	// Double-ensure redundant fields are zeroed for serialization (though they are yaml:"-")
	writable.Stack.DBType = ""
	writable.Stack.Features.Cache = false
	writable.Stack.Features.Search = false
	writable.Stack.Features.Queue = false

	if writable.Stack.UserID <= 0 {
		writable.Stack.UserID = 0
	} else {
		uid := os.Getuid()
		if uid >= 0 && writable.Stack.UserID == uid {
			writable.Stack.UserID = 0
		}
	}

	if writable.Stack.GroupID <= 0 {
		writable.Stack.GroupID = 0
	} else {
		gid := os.Getgid()
		if gid >= 0 && writable.Stack.GroupID == gid {
			writable.Stack.GroupID = 0
		}
	}

	for name, remote := range writable.Remotes {
		defaultMethod := NormalizeRemoteAuthMethod("")
		if NormalizeRemoteAuthMethod(remote.Auth.Method) == defaultMethod {
			remote.Auth.Method = ""
		}
		// Strip capabilities block if no capability is disabled (all-true = default behavior).
		// Storing capabilities: {files: true, media: true, db: true} is redundant noise.
		if remote.Capabilities != nil {
			caps := remote.Capabilities
			anyDisabled := (caps.Files != nil && !*caps.Files) ||
				(caps.Media != nil && !*caps.Media) ||
				(caps.DB != nil && !*caps.DB)
			if !anyDisabled {
				remote.Capabilities = nil
			}
		}
		writable.Remotes[name] = remote
	}

	if slicesEqual(writable.Stack.ChownDirList, GetDefaultChownDirList(writable.Framework)) {
		writable.Stack.ChownDirList = nil
	}

	// Strip ComposerVersion when it matches the auto-derived default
	if writable.Stack.ComposerVersion != "" {
		derivedDefault := "latest"
		if writable.Stack.PHPVersion != "" && !IsNumericDotVersionAtLeast(writable.Stack.PHPVersion, "7.2.5") {
			derivedDefault = "2.2"
		}
		if profileResult, err := ResolveRuntimeProfile(writable.Framework, writable.FrameworkVersion); err == nil {
			if profileResult.Profile.ComposerVersion != "" {
				derivedDefault = profileResult.Profile.ComposerVersion
			}
		}
		if writable.Stack.ComposerVersion == derivedDefault {
			writable.Stack.ComposerVersion = ""
		}
	}

	// Strip PHPVersion when it matches the auto-derived default
	if writable.Stack.PHPVersion != "" {
		var derivedDefault string
		if profileResult, err := ResolveRuntimeProfile(writable.Framework, writable.FrameworkVersion); err == nil {
			derivedDefault = profileResult.Profile.PHPVersion
		}
		if writable.Stack.PHPVersion == derivedDefault {
			writable.Stack.PHPVersion = ""
		}
	}

	// Strip NodeVersion when it matches the auto-derived default
	if writable.Stack.NodeVersion != "" {
		derivedDefault := "24"
		if profileResult, err := ResolveRuntimeProfile(writable.Framework, writable.FrameworkVersion); err == nil {
			if profileResult.Profile.NodeVersion != "" {
				derivedDefault = profileResult.Profile.NodeVersion
			}
		}
		if writable.Stack.NodeVersion == derivedDefault {
			writable.Stack.NodeVersion = ""
		}
	}

	// Strip DBVersion when it matches the auto-derived default
	if writable.Stack.DBVersion != "" {
		var derivedDefault string
		if profileResult, err := ResolveRuntimeProfile(writable.Framework, writable.FrameworkVersion); err == nil {
			derivedDefault = profileResult.Profile.DBVersion
		}
		if writable.Stack.DBVersion == derivedDefault {
			writable.Stack.DBVersion = ""
		}
	}

	// Strip Cache, Search, and Queue versions if they match defaults
	if profileResult, err := ResolveRuntimeProfile(writable.Framework, writable.FrameworkVersion); err == nil {
		if writable.Stack.WebRoot == profileResult.Profile.WebRoot {
			writable.Stack.WebRoot = ""
		}
		if writable.Stack.CacheVersion == profileResult.Profile.CacheVersion {
			writable.Stack.CacheVersion = ""
		}
		if writable.Stack.SearchVersion == profileResult.Profile.SearchVersion {
			writable.Stack.SearchVersion = ""
		}
		if writable.Stack.QueueVersion == profileResult.Profile.QueueVersion {
			writable.Stack.QueueVersion = ""
		}
	}

	return writable
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
