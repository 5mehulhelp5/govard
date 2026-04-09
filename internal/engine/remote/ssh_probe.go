package remote

import (
	"errors"
	"os/exec"

	"govard/internal/engine"
)

// SSHProbeError wraps an SSH probe failure with classified details.
type SSHProbeError struct {
	Err     error
	Output  string
	Details FailureDetails
}

func (e *SSHProbeError) Error() string {
	return e.Err.Error()
}

func (e *SSHProbeError) Unwrap() error {
	return e.Err
}

// ProbeSSHAuth performs a quick non-interactive SSH connectivity check.
// It returns nil if key-based authentication succeeds, or an *SSHProbeError
// with classified failure details. Use IsAuthFailure to distinguish
// authentication problems from network/other errors.
func ProbeSSHAuth(remoteName string, remoteCfg engine.RemoteConfig) error {
	args := BuildSSHArgs(remoteName, remoteCfg, false, false)
	args = append(args, "-o", "ConnectTimeout=5", RemoteTarget(remoteCfg), "true")
	cmd := exec.Command("ssh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &SSHProbeError{
			Err:     err,
			Output:  string(output),
			Details: ClassifyFailure(err, string(output)),
		}
	}
	return nil
}

// IsAuthFailure reports whether err represents an SSH authentication failure.
func IsAuthFailure(err error) bool {
	var probeErr *SSHProbeError
	if errors.As(err, &probeErr) {
		return probeErr.Details.Category == FailureCategoryAuth
	}
	return false
}

// IsNetworkFailure reports whether err represents an SSH network failure.
func IsNetworkFailure(err error) bool {
	var probeErr *SSHProbeError
	if errors.As(err, &probeErr) {
		return probeErr.Details.Category == FailureCategoryNetwork
	}
	return false
}
