package engine

import (
	"fmt"
	"regexp"
	"strings"
)

type ProviderKind string

const (
	ProviderKindTunnel            ProviderKind = "tunnel"
	ProviderKindSecrets           ProviderKind = "secrets"
	ProviderKindBlueprintRegistry ProviderKind = "blueprint_registry"
)

type ProviderRef struct {
	Kind ProviderKind `json:"kind" yaml:"kind"`
	Name string       `json:"name" yaml:"name"`
}

var providerNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

func NormalizeProviderName(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func ValidateProviderRef(ref ProviderRef) error {
	if strings.TrimSpace(string(ref.Kind)) == "" {
		return fmt.Errorf("provider kind is required")
	}

	switch ref.Kind {
	case ProviderKindTunnel, ProviderKindSecrets, ProviderKindBlueprintRegistry:
	default:
		return fmt.Errorf("unsupported provider kind %q", ref.Kind)
	}

	name := NormalizeProviderName(ref.Name)
	if name == "" {
		return fmt.Errorf("provider name is required")
	}
	if !providerNamePattern.MatchString(name) {
		return fmt.Errorf("provider name %q is invalid (allowed: lowercase letters, numbers, hyphen, underscore)", ref.Name)
	}
	return nil
}
