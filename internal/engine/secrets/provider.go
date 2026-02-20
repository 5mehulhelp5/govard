package secrets

import (
	"context"
	"fmt"
	"strings"

	"govard/internal/engine"
)

const (
	onePasswordProviderName = "1password"
)

// Provider resolves secret references into concrete values.
type Provider interface {
	Name() string
	Resolve(ctx context.Context, ref string) (string, error)
}

// IsSecretReference returns true when the value is a supported secret ref.
func IsSecretReference(raw string) bool {
	return SecretProviderNameForReference(raw) != ""
}

// SecretProviderNameForReference returns the provider name inferred from a secret ref.
func SecretProviderNameForReference(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if strings.HasPrefix(trimmed, "op://") {
		return onePasswordProviderName
	}
	return ""
}

// NewProvider constructs a secrets provider from a normalized provider reference.
func NewProvider(ref engine.ProviderRef) (Provider, error) {
	if err := engine.ValidateProviderRef(ref); err != nil {
		return nil, err
	}
	if ref.Kind != engine.ProviderKindSecrets {
		return nil, fmt.Errorf("unsupported provider kind %q", ref.Kind)
	}

	switch engine.NormalizeProviderName(ref.Name) {
	case onePasswordProviderName, "op", "onepassword":
		return NewOPProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported secrets provider %q", ref.Name)
	}
}
