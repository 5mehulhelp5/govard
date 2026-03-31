package tunnel

import (
	"context"
	"fmt"
	"strings"

	"govard/internal/engine"
)

const (
	cloudflareProviderName = "cloudflare"
)

// StartOptions controls tunnel startup behavior.
type StartOptions struct {
	TargetURL   string
	NoTLSVerify bool
	HostHeader  string
}

// StartPlan describes the executable invocation required to start a tunnel.
type StartPlan struct {
	Binary string
	Args   []string
	Env    []string
}

// CommandString renders a human-readable command string.
func (plan StartPlan) CommandString() string {
	parts := []string{}
	if strings.TrimSpace(plan.Binary) != "" {
		parts = append(parts, plan.Binary)
	}
	parts = append(parts, plan.Args...)
	return strings.Join(parts, " ")
}

// Provider builds tunnel startup plans.
type Provider interface {
	Name() string
	BuildStartPlan(options StartOptions) (StartPlan, error)
}

// RuntimeProvider is reserved for providers that start and track tunnel lifecycle directly.
type RuntimeProvider interface {
	Provider
	Start(ctx context.Context, options StartOptions) error
}

// NewProvider resolves a tunnel provider by provider reference.
func NewProvider(ref engine.ProviderRef) (Provider, error) {
	if err := engine.ValidateProviderRef(ref); err != nil {
		return nil, err
	}
	if ref.Kind != engine.ProviderKindTunnel {
		return nil, fmt.Errorf("unsupported provider kind %q", ref.Kind)
	}

	switch engine.NormalizeProviderName(ref.Name) {
	case cloudflareProviderName, "cloudflared", "cloudflare-tunnel":
		return cloudflareProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported tunnel provider %q", ref.Name)
	}
}
