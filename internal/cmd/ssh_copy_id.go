package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"govard/internal/engine"
	"govard/internal/engine/remote"

	"github.com/pterm/pterm"
)

// resolvePublicKeyForRemote finds the best SSH public key to copy for a remote.
// If explicitKeyPath is provided, it normalizes and uses that. Otherwise, it
// resolves from config, then falls back to default key file probing.
func resolvePublicKeyForRemote(remoteName string, remoteCfg engine.RemoteConfig, explicitKeyPath string) string {
	pubKeyPath := ""

	if explicitKeyPath != "" {
		pubKeyPath = strings.TrimSuffix(explicitKeyPath, ".pub") + ".pub"
	} else {
		privateKeyPath, _ := remote.ResolveSSHKeyPath(remoteName, remoteCfg)
		if privateKeyPath != "" {
			pubKeyPath = strings.TrimSuffix(privateKeyPath, ".pub") + ".pub"
		}
	}

	if pubKeyPath != "" && fileExists(pubKeyPath) {
		return pubKeyPath
	}

	// Default fallback: probe well-known key types
	candidates := []string{"~/.ssh/id_ed25519.pub", "~/.ssh/id_ecdsa.pub", "~/.ssh/id_rsa.pub"}
	for _, c := range candidates {
		resolved := remote.NormalizePath(c)
		if fileExists(resolved) {
			return resolved
		}
	}

	return ""
}

// copySSHKeyToRemote copies the given public key to the remote server's
// authorized_keys. It uses ssh-copy-id when available, falling back to
// a manual append via SSH.
func copySSHKeyToRemote(remoteName string, remoteCfg engine.RemoteConfig, pubKeyPath string) error {
	// Prefer ssh-copy-id if available
	if sshCopyIdBin, err := exec.LookPath("ssh-copy-id"); err == nil {
		args := []string{"-i", pubKeyPath}
		if remoteCfg.Port > 0 {
			args = append(args, "-p", fmt.Sprintf("%d", remoteCfg.Port))
		}
		args = append(args, remote.RemoteTarget(remoteCfg))
		cmd := exec.Command(sshCopyIdBin, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Fallback for systems without ssh-copy-id
	pubKeyContent, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	setupCmd := fmt.Sprintf(
		"mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '%s' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys",
		strings.TrimSpace(string(pubKeyContent)),
	)
	sshCmd := remote.BuildSSHExecCommand(remoteName, remoteCfg, false, setupCmd)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	return sshCmd.Run()
}

// offerSSHKeyCopyOnAuthFailure probes SSH auth for the given remote and, if
// key-based authentication fails, interactively offers to copy the local
// SSH public key before the caller proceeds to the full SSH connection.
//
// Returns an error only for non-auth failures (network, host key, etc.).
// Auth failures are handled by offering copy-id; if the user declines,
// nil is returned so the caller can fall through to password-based SSH.
func offerSSHKeyCopyOnAuthFailure(remoteName string, remoteCfg engine.RemoteConfig) error {
	probeErr := remote.ProbeSSHAuth(remoteName, remoteCfg)
	if probeErr == nil {
		return nil // Auth OK, nothing to do
	}

	if !remote.IsAuthFailure(probeErr) {
		// Network, host key, or other non-auth error — bubble up
		return fmt.Errorf("SSH connection failed: %w", probeErr)
	}

	// Auth-specific failure — offer to copy key if we're in a terminal
	if !stdinIsTerminal() {
		return nil // Non-interactive, let SSH handle it
	}

	pubKeyPath := resolvePublicKeyForRemote(remoteName, remoteCfg, "")
	if pubKeyPath == "" {
		pterm.Warning.Println("SSH key authentication failed and no local public key found.")
		pterm.Warning.Println("SSH will ask for your password. To set up key auth later, run: govard remote copy-id " + remoteName)
		return nil
	}

	confirmed, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(true).
		Show(fmt.Sprintf(
			"SSH key auth failed for '%s'. Copy your public key (%s) to the remote server?",
			remoteName,
			filepath.Base(pubKeyPath),
		))

	if !confirmed {
		return nil // User declined, let SSH ask for password
	}

	pterm.Info.Printf("Copying public key '%s' to remote '%s' (%s)...\n", pubKeyPath, remoteName, remote.RemoteTarget(remoteCfg))

	if err := copySSHKeyToRemote(remoteName, remoteCfg, pubKeyPath); err != nil {
		pterm.Warning.Printf("Failed to copy SSH key: %v\n", err)
		pterm.Warning.Println("Continuing with password authentication...")
		return nil
	}

	pterm.Success.Printf("SSH key copied to '%s'. Future connections will use key authentication.\n", remoteName)
	return nil
}
