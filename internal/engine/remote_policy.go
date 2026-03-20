package engine

import (
	"fmt"
	"strings"
)

const (
	RemoteEnvDev     = "dev"
	RemoteEnvStaging = "staging"
	RemoteEnvProd    = "prod"

	RemoteCapabilityFiles  = "files"
	RemoteCapabilityMedia  = "media"
	RemoteCapabilityDB     = "db"
	RemoteCapabilityCache  = "cache"
	RemoteCapabilityDeploy = "deploy"
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

func normalizeRemoteCapabilities(capabilities RemoteCapabilities) RemoteCapabilities {
	if !capabilities.Files && !capabilities.Media && !capabilities.DB && !capabilities.Cache && !capabilities.Deploy {
		return defaultRemoteCapabilities()
	}
	return capabilities
}

func defaultRemoteCapabilities() RemoteCapabilities {
	return RemoteCapabilities{
		Files:  true,
		Media:  true,
		DB:     true,
		Cache:  true,
		Deploy: true,
	}
}

func RemoteCapabilityEnabled(remoteCfg RemoteConfig, capability string) bool {
	switch strings.ToLower(strings.TrimSpace(capability)) {
	case RemoteCapabilityFiles:
		return remoteCfg.Capabilities.Files
	case RemoteCapabilityMedia:
		return remoteCfg.Capabilities.Media
	case RemoteCapabilityDB:
		return remoteCfg.Capabilities.DB
	case RemoteCapabilityCache:
		return remoteCfg.Capabilities.Cache
	case RemoteCapabilityDeploy:
		return remoteCfg.Capabilities.Deploy
	default:
		return false
	}
}

func RemoteCapabilityList(remoteCfg RemoteConfig) []string {
	names := make([]string, 0, 5)
	for _, name := range []string{
		RemoteCapabilityFiles,
		RemoteCapabilityMedia,
		RemoteCapabilityDB,
		RemoteCapabilityCache,
		RemoteCapabilityDeploy,
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

func ParseRemoteCapabilitiesCSV(raw string) (RemoteCapabilities, error) {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	if len(parts) == 1 {
		switch strings.ToLower(strings.TrimSpace(parts[0])) {
		case "", "all":
			return defaultRemoteCapabilities(), nil
		}
	}

	parsed := RemoteCapabilities{}
	for _, part := range parts {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" {
			continue
		}
		switch name {
		case RemoteCapabilityFiles:
			parsed.Files = true
		case RemoteCapabilityMedia:
			parsed.Media = true
		case RemoteCapabilityDB:
			parsed.DB = true
		case RemoteCapabilityCache:
			parsed.Cache = true
		case RemoteCapabilityDeploy:
			parsed.Deploy = true
		default:
			return RemoteCapabilities{}, fmt.Errorf("unsupported remote capability: %s", name)
		}
	}

	if !parsed.Files && !parsed.Media && !parsed.DB && !parsed.Cache && !parsed.Deploy {
		return RemoteCapabilities{}, fmt.Errorf("at least one remote capability is required")
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
