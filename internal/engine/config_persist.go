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
	cacheEnabled := strings.TrimSpace(strings.ToLower(writable.Stack.Services.Cache)) != "" &&
		strings.TrimSpace(strings.ToLower(writable.Stack.Services.Cache)) != "none"
	searchEnabled := strings.TrimSpace(strings.ToLower(writable.Stack.Services.Search)) != "" &&
		strings.TrimSpace(strings.ToLower(writable.Stack.Services.Search)) != "none"
	if writable.Stack.Features.Redis == cacheEnabled {
		writable.Stack.Features.Redis = false
	}
	if writable.Stack.Features.Elasticsearch == searchEnabled {
		writable.Stack.Features.Elasticsearch = false
	}

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
