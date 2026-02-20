package tunnel

import (
	"fmt"
	"net/url"
	"strings"
)

type cloudflareProvider struct{}

func (provider cloudflareProvider) Name() string {
	return cloudflareProviderName
}

func (provider cloudflareProvider) BuildStartPlan(options StartOptions) (StartPlan, error) {
	targetURL, err := normalizeTargetURL(options.TargetURL)
	if err != nil {
		return StartPlan{}, err
	}

	args := []string{"tunnel", "--url", targetURL}
	if options.NoTLSVerify {
		args = append(args, "--no-tls-verify")
	}

	return StartPlan{
		Binary: "cloudflared",
		Args:   args,
	}, nil
}

func normalizeTargetURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("target URL is required")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid target URL %q: %w", raw, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("target URL %q must include scheme and host", raw)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported target URL scheme %q (allowed: http, https)", parsed.Scheme)
	}

	return parsed.String(), nil
}
