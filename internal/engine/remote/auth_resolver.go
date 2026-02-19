package remote

import (
	"os"
	"path/filepath"
	"strings"

	"govard/internal/engine"
)

const (
	AuthMethodSSHAgent = engine.RemoteAuthMethodSSHAgent
	AuthMethodKeychain = engine.RemoteAuthMethodKeychain
	AuthMethodKeyfile  = engine.RemoteAuthMethodKeyfile
)

const remoteKeyPathEnvVar = "GOVARD_REMOTE_KEY_PATH"

// NormalizeAuthMethod normalizes remote auth method aliases to canonical values.
func NormalizeAuthMethod(method string) string {
	return engine.NormalizeRemoteAuthMethod(method)
}

func IsSupportedAuthMethod(method string) bool {
	return engine.IsSupportedRemoteAuthMethod(method)
}

func PersistSSHKeyPath(remoteName string, keyPath string) error {
	return NewKeychainStore().Set(authStoreKey(remoteName), normalizePath(keyPath))
}

// ResolveSSHKeyPath resolves SSH identity path with explicit precedence:
// config key_path -> per-remote env -> global env -> keychain/file store -> default keyfile probes.
func ResolveSSHKeyPath(remoteName string, remoteCfg engine.RemoteConfig) (string, string) {
	if fromConfig := normalizePath(remoteCfg.Auth.KeyPath); fromConfig != "" {
		return fromConfig, "config"
	}

	if fromRemoteEnv := normalizePath(os.Getenv(RemoteKeyPathEnvVar(remoteName))); fromRemoteEnv != "" {
		return fromRemoteEnv, "env:" + RemoteKeyPathEnvVar(remoteName)
	}

	if fromGlobalEnv := normalizePath(os.Getenv(remoteKeyPathEnvVar)); fromGlobalEnv != "" {
		return fromGlobalEnv, "env:" + remoteKeyPathEnvVar
	}

	authMethod := NormalizeAuthMethod(remoteCfg.Auth.Method)
	if authMethod == AuthMethodKeychain {
		if fromStore, err := NewKeychainStore().Get(authStoreKey(remoteName)); err == nil {
			if fromStorePath := normalizePath(fromStore); fromStorePath != "" {
				return fromStorePath, "store:keychain"
			}
		}
	}

	if authMethod == AuthMethodKeyfile {
		if defaultKey := firstExistingSSHKeyfile(); defaultKey != "" {
			return defaultKey, "default:keyfile"
		}
	}

	return "", ""
}

func RemoteKeyPathEnvVar(remoteName string) string {
	if strings.TrimSpace(remoteName) == "" {
		return remoteKeyPathEnvVar
	}

	builder := strings.Builder{}
	builder.WriteString(remoteKeyPathEnvVar)
	builder.WriteString("_")
	for _, r := range strings.ToUpper(strings.TrimSpace(remoteName)) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('_')
	}
	return builder.String()
}

func authStoreKey(remoteName string) string {
	normalized := strings.ToLower(strings.TrimSpace(remoteName))
	if normalized == "" {
		normalized = "default"
	}
	return "remote." + normalized + ".key_path"
}

func firstExistingSSHKeyfile() string {
	candidates := []string{
		"~/.ssh/id_ed25519",
		"~/.ssh/id_ecdsa",
		"~/.ssh/id_rsa",
	}
	for _, candidate := range candidates {
		path := normalizePath(candidate)
		if path == "" {
			continue
		}
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}

func normalizePath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	expanded := os.ExpandEnv(trimmed)
	if expanded == "~" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Clean(home)
		}
		return expanded
	}
	if strings.HasPrefix(expanded, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Clean(filepath.Join(home, strings.TrimPrefix(expanded, "~/")))
		}
	}
	return filepath.Clean(expanded)
}
