package remote

import (
	"errors"
	"os/exec"
	"strings"
)

const (
	FailureCategoryNetwork    = "network"
	FailureCategoryAuth       = "auth"
	FailureCategoryPermission = "permission"
	FailureCategoryHostKey    = "host_key"
	FailureCategoryDependency = "dependency"
	FailureCategoryUnknown    = "unknown"
)

type FailureDetails struct {
	Category string
	Hint     string
}

func ClassifyFailure(err error, output string) FailureDetails {
	combined := strings.ToLower(strings.TrimSpace(output + "\n" + errorText(err)))

	switch {
	case containsAny(combined,
		"host key verification failed",
		"remote host identification has changed",
		"offending key in",
	):
		return FailureDetails{
			Category: FailureCategoryHostKey,
			Hint:     "Verify host fingerprint and known_hosts. Use strict host key mode with a trusted known_hosts file.",
		}
	case containsAny(combined,
		"permission denied (publickey",
		"no supported authentication methods available",
		"too many authentication failures",
	):
		return FailureDetails{
			Category: FailureCategoryAuth,
			Hint:     "Check SSH key/agent configuration and remote user access.",
		}
	case containsAny(combined,
		"could not resolve hostname",
		"name or service not known",
		"temporary failure in name resolution",
		"connection timed out",
		"connection refused",
		"network is unreachable",
		"no route to host",
		"connection reset by peer",
		"operation timed out",
	):
		return FailureDetails{
			Category: FailureCategoryNetwork,
			Hint:     "Check DNS, VPN/network route, host/port, and firewall rules.",
		}
	case containsAny(combined,
		"command not found",
		"rsync: not found",
	):
		return FailureDetails{
			Category: FailureCategoryDependency,
			Hint:     "Install missing remote dependency (for example rsync) or adjust the workflow.",
		}
	case containsAny(combined,
		"permission denied",
		"operation not permitted",
	):
		return FailureDetails{
			Category: FailureCategoryPermission,
			Hint:     "Check remote filesystem permissions and target path ownership.",
		}
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 255 {
		return FailureDetails{
			Category: FailureCategoryAuth,
			Hint:     "SSH connection failed. Check host, user, port, and authentication settings.",
		}
	}

	return FailureDetails{
		Category: FailureCategoryUnknown,
		Hint:     "Review command output for details and retry with corrected remote settings.",
	}
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func containsAny(text string, patterns ...string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}
