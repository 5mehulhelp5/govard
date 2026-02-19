package remote

import (
	"fmt"
	"os/exec"
	"strings"

	"govard/internal/engine"
)

func RemoteTarget(remoteCfg engine.RemoteConfig) string {
	return fmt.Sprintf("%s@%s", remoteCfg.User, remoteCfg.Host)
}

func BuildSSHExecCommand(remoteName string, remoteCfg engine.RemoteConfig, forwardAgent bool, remoteCommand string) *exec.Cmd {
	args := BuildSSHArgs(remoteName, remoteCfg, forwardAgent)
	args = append(args, RemoteTarget(remoteCfg), remoteCommand)
	return exec.Command("ssh", args...)
}

func runRemoteCapture(remoteName string, remoteCfg engine.RemoteConfig, remoteCommand string) (string, error) {
	cmd := BuildSSHExecCommand(remoteName, remoteCfg, true, remoteCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("remote command failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func BuildRsyncCommand(
	remoteName string,
	source string,
	destination string,
	remoteCfg engine.RemoteConfig,
	deleteFiles bool,
	resume bool,
	includePatterns []string,
	excludePatterns []string,
) *exec.Cmd {
	args := []string{"-az"}
	if deleteFiles {
		args = append(args, "--delete")
	}
	if resume {
		args = append(args, "--partial", "--append-verify")
	}
	for _, pattern := range includePatterns {
		trimmed := strings.TrimSpace(pattern)
		if trimmed == "" {
			continue
		}
		args = append(args, "--include", trimmed)
	}
	for _, pattern := range excludePatterns {
		trimmed := strings.TrimSpace(pattern)
		if trimmed == "" {
			continue
		}
		args = append(args, "--exclude", trimmed)
	}

	sshArgs := append([]string{"ssh"}, BuildSSHArgs(remoteName, remoteCfg, false)...)
	args = append(args, "-e", strings.Join(sshArgs, " "))
	args = append(args, source, destination)

	return exec.Command("rsync", args...)
}
