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
	case "", "staging", "stage", "qa", "uat", "test":
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
	if !capabilities.Files && !capabilities.Media && !capabilities.DB && !capabilities.Deploy {
		return defaultRemoteCapabilities()
	}
	return capabilities
}

func defaultRemoteCapabilities() RemoteCapabilities {
	return RemoteCapabilities{
		Files:  true,
		Media:  true,
		DB:     true,
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
	case RemoteCapabilityDeploy:
		return remoteCfg.Capabilities.Deploy
	default:
		return false
	}
}

func RemoteCapabilityList(remoteCfg RemoteConfig) []string {
	names := make([]string, 0, 4)
	for _, name := range []string{
		RemoteCapabilityFiles,
		RemoteCapabilityMedia,
		RemoteCapabilityDB,
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
		case RemoteCapabilityDeploy:
			parsed.Deploy = true
		default:
			return RemoteCapabilities{}, fmt.Errorf("unsupported remote capability: %s", name)
		}
	}

	if !parsed.Files && !parsed.Media && !parsed.DB && !parsed.Deploy {
		return RemoteCapabilities{}, fmt.Errorf("at least one remote capability is required")
	}
	return parsed, nil
}

func RemoteWriteBlocked(remoteCfg RemoteConfig) (bool, string) {
	environment := NormalizeRemoteEnvironment(remoteCfg.Environment)
	switch {
	case remoteCfg.Protected:
		return true, "explicit protected flag"
	case environment == RemoteEnvProd:
		return true, "production environment protection"
	default:
		return false, ""
	}
}
