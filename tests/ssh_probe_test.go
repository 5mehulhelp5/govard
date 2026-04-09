package tests

import (
	"errors"
	"fmt"
	"testing"

	"govard/internal/engine/remote"
)

func TestSSHProbeErrorIsAuthFailure(t *testing.T) {
	err := &remote.SSHProbeError{
		Err:    errors.New("exit status 255"),
		Output: "Permission denied (publickey).",
		Details: remote.FailureDetails{
			Category: remote.FailureCategoryAuth,
			Hint:     "Check SSH key/agent configuration and remote user access.",
		},
	}

	if !remote.IsAuthFailure(err) {
		t.Fatal("expected IsAuthFailure to return true for auth probe error")
	}
	if remote.IsNetworkFailure(err) {
		t.Fatal("expected IsNetworkFailure to return false for auth probe error")
	}
}

func TestSSHProbeErrorIsNetworkFailure(t *testing.T) {
	err := &remote.SSHProbeError{
		Err:    errors.New("exit status 255"),
		Output: "Connection timed out",
		Details: remote.FailureDetails{
			Category: remote.FailureCategoryNetwork,
			Hint:     "Check DNS, VPN/network route, host/port, and firewall rules.",
		},
	}

	if remote.IsAuthFailure(err) {
		t.Fatal("expected IsAuthFailure to return false for network probe error")
	}
	if !remote.IsNetworkFailure(err) {
		t.Fatal("expected IsNetworkFailure to return true for network probe error")
	}
}

func TestSSHProbeErrorUnwrap(t *testing.T) {
	innerErr := errors.New("exit status 255")
	probeErr := &remote.SSHProbeError{
		Err:    innerErr,
		Output: "Permission denied (publickey).",
		Details: remote.FailureDetails{
			Category: remote.FailureCategoryAuth,
		},
	}

	if !errors.Is(probeErr, innerErr) {
		t.Fatal("expected SSHProbeError to unwrap to inner error")
	}
}

func TestIsAuthFailureWithNonProbeError(t *testing.T) {
	err := errors.New("some random error")
	if remote.IsAuthFailure(err) {
		t.Fatal("expected IsAuthFailure to return false for non-SSHProbeError")
	}
}

func TestIsAuthFailureWithNil(t *testing.T) {
	if remote.IsAuthFailure(nil) {
		t.Fatal("expected IsAuthFailure to return false for nil")
	}
}

func TestSSHProbeErrorWrappedInFmtErrorf(t *testing.T) {
	innerErr := &remote.SSHProbeError{
		Err:    errors.New("exit status 255"),
		Output: "Permission denied (publickey).",
		Details: remote.FailureDetails{
			Category: remote.FailureCategoryAuth,
		},
	}

	wrappedErr := fmt.Errorf("SSH connection failed: %w", innerErr)

	if !remote.IsAuthFailure(wrappedErr) {
		t.Fatal("expected IsAuthFailure to detect auth failure through wrapped error")
	}
}
