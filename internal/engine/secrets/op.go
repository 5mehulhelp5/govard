package secrets

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CommandRunner executes a command and returns combined output.
type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

type opProvider struct {
	runner CommandRunner
}

// NewOPProvider creates the default 1Password CLI provider.
func NewOPProvider() Provider {
	return NewOPProviderWithRunner(defaultOPCommandRunner)
}

// NewOPProviderWithRunner creates a 1Password provider with a custom command runner.
func NewOPProviderWithRunner(runner CommandRunner) Provider {
	if runner == nil {
		runner = defaultOPCommandRunner
	}
	return &opProvider{runner: runner}
}

func (provider *opProvider) Name() string {
	return onePasswordProviderName
}

func (provider *opProvider) Resolve(ctx context.Context, ref string) (string, error) {
	trimmedRef := strings.TrimSpace(ref)
	if !IsSecretReference(trimmedRef) {
		return "", fmt.Errorf("unsupported secret reference %q", ref)
	}

	output, err := provider.runner(ctx, "op", "read", trimmedRef)
	if err != nil {
		details := strings.TrimSpace(string(output))
		if details == "" {
			return "", fmt.Errorf("op read %q failed: %w", trimmedRef, err)
		}
		return "", fmt.Errorf("op read %q failed: %w (%s)", trimmedRef, err, details)
	}

	value := strings.TrimSpace(string(output))
	if value == "" {
		return "", fmt.Errorf("op read %q returned an empty value", trimmedRef)
	}
	return value, nil
}

func defaultOPCommandRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	return command.CombinedOutput()
}
