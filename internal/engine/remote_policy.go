package engine

import (
	"fmt"
	"strings"
)

const (
	RemoteEnvDev     = "dev"
	RemoteEnvStaging = "staging"
	RemoteEnvProd    = "prod"

	RemoteCapabilityFiles = "files"
	RemoteCapabilityMedia = "media"
	RemoteCapabilityDB    = "db"
)

var validRemoteEnvironments = map[string]struct{}{
	RemoteEnvDev:     {},
	RemoteEnvStaging: {},
	RemoteEnvProd:    {},
}

func NormalizeRemoteEnvironment(value string) string {
	cleaned := strings.ToLower(strings.TrimSpace(value))
	switch cleaned {
	case "", "staging", "stage", "stg", "qa", "uat", "test":
		return RemoteEnvStaging
	case "dev", "development", "local":
		return RemoteEnvDev
	case "prod", "production", "live":
		return RemoteEnvProd
	default:
		return cleaned
	}
}

func IsValidRemoteEnvironment(value string) bool {
	_, ok := validRemoteEnvironments[NormalizeRemoteEnvironment(value)]
	return ok
}

// RemoteCapabilityEnabled returns true if the capability is allowed.
// nil means "not set" → allowed by default.
// Only an explicit false blocks the capability.
func RemoteCapabilityEnabled(remoteCfg RemoteConfig, capability string) bool {
	switch strings.ToLower(strings.TrimSpace(capability)) {
	case RemoteCapabilityFiles:
		return remoteCfg.Capabilities.Files == nil || *remoteCfg.Capabilities.Files
	case RemoteCapabilityMedia:
		return remoteCfg.Capabilities.Media == nil || *remoteCfg.Capabilities.Media
	case RemoteCapabilityDB:
		return remoteCfg.Capabilities.DB == nil || *remoteCfg.Capabilities.DB
	default:
		return false
	}
}

// RemoteCapabilityList returns the list of enabled capabilities.
// nil is treated as enabled (allow-by-default).
func RemoteCapabilityList(remoteCfg RemoteConfig) []string {
	names := make([]string, 0, 3)
	for _, name := range []string{
		RemoteCapabilityFiles,
		RemoteCapabilityMedia,
		RemoteCapabilityDB,
	} {
		if RemoteCapabilityEnabled(remoteCfg, name) {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return []string{"none"}
	}
	return names
}

// ParseRemoteCapabilitiesCSV parses a comma-separated list of capability names to BLOCK.
// An empty string or "none" means block nothing (all enabled).
// Example: "db" means block only db; "files,media" blocks files and media.
func ParseRemoteCapabilitiesCSV(raw string) (RemoteCapabilities, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "none") || strings.EqualFold(raw, "all") {
		return RemoteCapabilities{}, nil
	}

	parsed := RemoteCapabilities{}
	falseVal := false
	for _, part := range strings.Split(raw, ",") {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" {
			continue
		}
		switch name {
		case RemoteCapabilityFiles:
			parsed.Files = &falseVal
		case RemoteCapabilityMedia:
			parsed.Media = &falseVal
		case RemoteCapabilityDB:
			parsed.DB = &falseVal
		default:
			return RemoteCapabilities{}, fmt.Errorf("unsupported remote capability: %s", name)
		}
	}
	return parsed, nil
}

// RemoteWriteBlocked checks whether writes to a remote should be blocked.
// The remoteName is the map key from the config (e.g., "production", "dev").
// If remoteCfg.Protected is explicitly set, it takes precedence.
// Otherwise, remotes whose name normalizes to "prod" are auto-protected.
func RemoteWriteBlocked(remoteName string, remoteCfg RemoteConfig) (bool, string) {
	if remoteCfg.Protected != nil {
		if *remoteCfg.Protected {
			return true, "explicit protected flag"
		}
		return false, ""
	}
	environment := NormalizeRemoteEnvironment(remoteName)
	if environment == RemoteEnvProd {
		return true, "production environment protection (auto)"
	}
	return false, ""
}
