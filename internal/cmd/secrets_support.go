package cmd

import (
	"context"
	"fmt"
	"strings"

	"govard/internal/engine"
	"govard/internal/engine/secrets"
)

type secretsSupportDependencies struct {
	NewProvider func(ref engine.ProviderRef) (secrets.Provider, error)
}

var secretsSupportDeps = secretsSupportDependencies{
	NewProvider: secrets.NewProvider,
}

func resolveRemoteConfigSecrets(remoteName string, remoteCfg engine.RemoteConfig) (engine.RemoteConfig, error) {
	resolved := remoteCfg
	providerCache := map[string]secrets.Provider{}

	var err error
	resolved.Host, err = resolveSecretField(
		fmt.Sprintf("remote '%s' field host", remoteName),
		resolved.Host,
		providerCache,
	)
	if err != nil {
		return engine.RemoteConfig{}, err
	}
	resolved.User, err = resolveSecretField(
		fmt.Sprintf("remote '%s' field user", remoteName),
		resolved.User,
		providerCache,
	)
	if err != nil {
		return engine.RemoteConfig{}, err
	}
	resolved.Path, err = resolveSecretField(
		fmt.Sprintf("remote '%s' field path", remoteName),
		resolved.Path,
		providerCache,
	)
	if err != nil {
		return engine.RemoteConfig{}, err
	}
	resolved.Auth.KeyPath, err = resolveSecretField(
		fmt.Sprintf("remote '%s' field auth.key_path", remoteName),
		resolved.Auth.KeyPath,
		providerCache,
	)
	if err != nil {
		return engine.RemoteConfig{}, err
	}
	resolved.Auth.KnownHostsFile, err = resolveSecretField(
		fmt.Sprintf("remote '%s' field auth.known_hosts_file", remoteName),
		resolved.Auth.KnownHostsFile,
		providerCache,
	)
	if err != nil {
		return engine.RemoteConfig{}, err
	}
	resolved.Paths.Media, err = resolveSecretField(
		fmt.Sprintf("remote '%s' field paths.media", remoteName),
		resolved.Paths.Media,
		providerCache,
	)
	if err != nil {
		return engine.RemoteConfig{}, err
	}
	return resolved, nil
}

func resolveSecretField(field string, value string, providerCache map[string]secrets.Provider) (string, error) {
	providerName := secrets.SecretProviderNameForReference(value)
	if providerName == "" {
		return value, nil
	}

	provider := providerCache[providerName]
	if provider == nil {
		ref := engine.ProviderRef{
			Kind: engine.ProviderKindSecrets,
			Name: providerName,
		}
		createdProvider, err := secretsSupportDeps.NewProvider(ref)
		if err != nil {
			return "", fmt.Errorf("%s: initialize provider %q: %w", field, providerName, err)
		}
		provider = createdProvider
		providerCache[providerName] = provider
	}

	secretRef := strings.TrimSpace(value)
	resolvedValue, err := provider.Resolve(context.Background(), secretRef)
	if err != nil {
		return "", fmt.Errorf("%s: %w", field, err)
	}
	return resolvedValue, nil
}

// ResolveRemoteConfigSecretsForTest exposes remote secret resolution for tests.
func ResolveRemoteConfigSecretsForTest(remoteName string, remoteCfg engine.RemoteConfig) (engine.RemoteConfig, error) {
	return resolveRemoteConfigSecrets(remoteName, remoteCfg)
}

// SetSecretsProviderFactoryForTest swaps the secrets provider factory and returns a restore callback.
func SetSecretsProviderFactoryForTest(factory func(ref engine.ProviderRef) (secrets.Provider, error)) func() {
	previous := secretsSupportDeps.NewProvider
	if factory == nil {
		secretsSupportDeps.NewProvider = secrets.NewProvider
	} else {
		secretsSupportDeps.NewProvider = factory
	}
	return func() {
		secretsSupportDeps.NewProvider = previous
	}
}
