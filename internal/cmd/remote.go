package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"govard/internal/engine"
	"govard/internal/engine/remote"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage remote environments",
}

var remoteAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add or update a remote environment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		startedAt := time.Now()
		config := loadWritableConfig()
		name := args[0]

		host, _ := cmd.Flags().GetString("host")
		user, _ := cmd.Flags().GetString("user")
		path, _ := cmd.Flags().GetString("path")
		port, _ := cmd.Flags().GetInt("port")
		protected, _ := cmd.Flags().GetBool("protected")
		environment, _ := cmd.Flags().GetString("environment")
		capabilitiesRaw, _ := cmd.Flags().GetString("capabilities")
		authMethodRaw, _ := cmd.Flags().GetString("auth-method")
		keyPathRaw, _ := cmd.Flags().GetString("key-path")
		strictHostKey, _ := cmd.Flags().GetBool("strict-host-key")
		knownHostsFile, _ := cmd.Flags().GetString("known-hosts-file")

		if host == "" || user == "" || path == "" {
			pterm.Error.Println("host, user, and path are required")
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:  "remote.add",
				Status:     remote.RemoteAuditStatusFailure,
				Category:   "validation",
				Remote:     name,
				DurationMS: time.Since(startedAt).Milliseconds(),
				Message:    "missing required flags: host/user/path",
			})
			return
		}

		environment = engine.NormalizeRemoteEnvironment(environment)
		if !engine.IsValidRemoteEnvironment(environment) {
			pterm.Error.Printf("unsupported remote environment '%s' (allowed: dev, staging, prod)\n", environment)
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:  "remote.add",
				Status:     remote.RemoteAuditStatusFailure,
				Category:   "validation",
				Remote:     name,
				DurationMS: time.Since(startedAt).Milliseconds(),
				Message:    "unsupported remote environment",
			})
			return
		}
		capabilities, err := engine.ParseRemoteCapabilitiesCSV(capabilitiesRaw)
		if err != nil {
			pterm.Error.Println(err.Error())
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:  "remote.add",
				Status:     remote.RemoteAuditStatusFailure,
				Category:   "validation",
				Remote:     name,
				DurationMS: time.Since(startedAt).Milliseconds(),
				Message:    err.Error(),
			})
			return
		}
		authMethod := remote.NormalizeAuthMethod(authMethodRaw)
		if !remote.IsSupportedAuthMethod(authMethod) {
			message := fmt.Sprintf("unsupported auth method '%s' (allowed: keychain, ssh-agent, keyfile)", authMethodRaw)
			pterm.Error.Println(message)
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:  "remote.add",
				Status:     remote.RemoteAuditStatusFailure,
				Category:   "validation",
				Remote:     name,
				DurationMS: time.Since(startedAt).Milliseconds(),
				Message:    message,
			})
			return
		}

		keyPath := strings.TrimSpace(keyPathRaw)
		if keyPath != "" && authMethod == remote.AuthMethodKeychain {
			if err := remote.PersistSSHKeyPath(name, keyPath); err != nil {
				pterm.Warning.Printf("Could not persist key path in auth store (%v); falling back to config auth.key_path.\n", err)
			} else {
				pterm.Info.Printf("Stored SSH key path for remote '%s' in auth store.\n", name)
				keyPath = ""
			}
		}

		if knownHostsFile != "" && !strictHostKey {
			strictHostKey = true
			pterm.Warning.Println("known-hosts-file specified: enabling strict host key checking.")
		}

		if config.Remotes == nil {
			config.Remotes = map[string]engine.RemoteConfig{}
		}

		config.Remotes[name] = engine.RemoteConfig{
			Host:         host,
			User:         user,
			Path:         path,
			Port:         port,
			Environment:  environment,
			Protected:    protected,
			Capabilities: capabilities,
			Auth: engine.RemoteAuth{
				Method:         authMethod,
				KeyPath:        keyPath,
				StrictHostKey:  strictHostKey,
				KnownHostsFile: knownHostsFile,
			},
		}

		saveConfig(config)
		pterm.Success.Printf(
			"Remote '%s' saved (environment=%s, capabilities=%s, auth=%s, protected=%t, strict_host_key=%t).\n",
			name,
			environment,
			strings.Join(engine.RemoteCapabilityList(config.Remotes[name]), ","),
			authMethod,
			protected,
			strictHostKey,
		)
		writeRemoteAuditEvent(remote.AuditEvent{
			Operation:  "remote.add",
			Status:     remote.RemoteAuditStatusSuccess,
			Remote:     name,
			DurationMS: time.Since(startedAt).Milliseconds(),
			Message:    "remote saved",
		})
	},
}

var remoteExecCmd = &cobra.Command{
	Use:   "exec [name] -- <command>",
	Short: "Execute a command on a remote environment",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		startedAt := time.Now()
		remoteName := args[0]
		config := loadFullConfig()
		remoteCfg, err := ensureRemoteKnown(config, remoteName)
		if err != nil {
			pterm.Error.Println(err.Error())
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:  "remote.exec",
				Status:     remote.RemoteAuditStatusFailure,
				Category:   "validation",
				Remote:     remoteName,
				DurationMS: time.Since(startedAt).Milliseconds(),
				Message:    err.Error(),
			})
			return
		}

		commandLine := strings.TrimSpace(strings.Join(args[1:], " "))
		if commandLine == "" {
			pterm.Error.Println("missing remote command after '--'")
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:  "remote.exec",
				Status:     remote.RemoteAuditStatusFailure,
				Category:   "validation",
				Remote:     remoteName,
				DurationMS: time.Since(startedAt).Milliseconds(),
				Message:    "missing remote command after '--'",
			})
			return
		}

		sshCmd := remote.BuildSSHExecCommand(remoteName, remoteCfg, true, commandLine)
		sshCmd.Stdin = os.Stdin
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr

		if err := sshCmd.Run(); err != nil {
			pterm.Error.Printf("Remote exec failed: %v\n", err)
			details := remote.ClassifyFailure(err, "")
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:  "remote.exec",
				Status:     remote.RemoteAuditStatusFailure,
				Category:   details.Category,
				Remote:     remoteName,
				DurationMS: time.Since(startedAt).Milliseconds(),
				Message:    err.Error(),
			})
			return
		}
		writeRemoteAuditEvent(remote.AuditEvent{
			Operation:  "remote.exec",
			Status:     remote.RemoteAuditStatusSuccess,
			Remote:     remoteName,
			DurationMS: time.Since(startedAt).Milliseconds(),
		})
	},
}

var remoteTestCmd = &cobra.Command{
	Use:   "test [name]",
	Short: "Test SSH connectivity to a remote",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName := args[0]
		config := loadFullConfig()
		remoteCfg, err := ensureRemoteKnown(config, remoteName)
		if err != nil {
			pterm.Error.Println(err.Error())
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation: "remote.test",
				Status:    remote.RemoteAuditStatusFailure,
				Category:  "validation",
				Remote:    remoteName,
				Message:   err.Error(),
			})
			return
		}
		pterm.Info.Printf(
			"Remote profile: environment=%s, capabilities=%s, auth=%s, protected=%t, strict_host_key=%t\n",
			remoteCfg.Environment,
			strings.Join(engine.RemoteCapabilityList(remoteCfg), ","),
			remote.NormalizeAuthMethod(remoteCfg.Auth.Method),
			remoteCfg.Protected,
			remoteCfg.Auth.StrictHostKey,
		)

		testArgs := remote.BuildSSHArgs(remoteName, remoteCfg, false)
		testArgs = append(testArgs, "-o", "ConnectTimeout=5", remote.RemoteTarget(remoteCfg), "echo govard-remote-ok")
		testCmd := exec.Command("ssh", testArgs...)
		sshStartedAt := time.Now()
		output, err := testCmd.CombinedOutput()
		sshDuration := time.Since(sshStartedAt)
		if err != nil {
			details := reportRemoteCommandFailure("SSH connectivity", err, output, sshDuration, false)
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:  "remote.test.ssh",
				Status:     remote.RemoteAuditStatusFailure,
				Category:   details.Category,
				Remote:     remoteName,
				DurationMS: sshDuration.Milliseconds(),
				Message:    err.Error(),
			})
			return
		}
		if !strings.Contains(string(output), "govard-remote-ok") {
			details := reportRemoteCommandFailure(
				"SSH connectivity",
				fmt.Errorf("unexpected SSH probe response"),
				output,
				sshDuration,
				false,
			)
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:  "remote.test.ssh",
				Status:     remote.RemoteAuditStatusFailure,
				Category:   details.Category,
				Remote:     remoteName,
				DurationMS: sshDuration.Milliseconds(),
				Message:    "unexpected SSH probe response",
			})
			return
		}
		pterm.Success.Printf("SSH connectivity check passed (%s).\n", formatDuration(sshDuration))
		writeRemoteAuditEvent(remote.AuditEvent{
			Operation:  "remote.test.ssh",
			Status:     remote.RemoteAuditStatusSuccess,
			Remote:     remoteName,
			DurationMS: sshDuration.Milliseconds(),
		})

		rsyncArgs := remote.BuildSSHArgs(remoteName, remoteCfg, false)
		rsyncArgs = append(rsyncArgs, "-o", "ConnectTimeout=5", remote.RemoteTarget(remoteCfg), "command -v rsync >/dev/null 2>&1 && echo govard-rsync-ok")
		rsyncCmd := exec.Command("ssh", rsyncArgs...)
		rsyncStartedAt := time.Now()
		rsyncOutput, err := rsyncCmd.CombinedOutput()
		rsyncDuration := time.Since(rsyncStartedAt)
		if err != nil {
			details := reportRemoteCommandFailure("Remote rsync availability", err, rsyncOutput, rsyncDuration, true)
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:  "remote.test.rsync",
				Status:     remote.RemoteAuditStatusWarning,
				Category:   details.Category,
				Remote:     remoteName,
				DurationMS: rsyncDuration.Milliseconds(),
				Message:    err.Error(),
			})
			return
		}
		if strings.Contains(string(rsyncOutput), "govard-rsync-ok") {
			pterm.Success.Printf("Remote rsync availability check passed (%s).\n", formatDuration(rsyncDuration))
			writeRemoteAuditEvent(remote.AuditEvent{
				Operation:  "remote.test.rsync",
				Status:     remote.RemoteAuditStatusSuccess,
				Remote:     remoteName,
				DurationMS: rsyncDuration.Milliseconds(),
			})
			return
		}

		details := reportRemoteCommandFailure(
			"Remote rsync availability",
			fmt.Errorf("unexpected rsync probe response"),
			rsyncOutput,
			rsyncDuration,
			true,
		)
		writeRemoteAuditEvent(remote.AuditEvent{
			Operation:  "remote.test.rsync",
			Status:     remote.RemoteAuditStatusWarning,
			Category:   details.Category,
			Remote:     remoteName,
			DurationMS: rsyncDuration.Milliseconds(),
			Message:    "unexpected rsync probe response",
		})
	},
}

func init() {
	remoteAddCmd.Flags().String("host", "", "Remote host")
	remoteAddCmd.Flags().String("user", "", "Remote user")
	remoteAddCmd.Flags().String("path", "", "Remote path")
	remoteAddCmd.Flags().Int("port", 22, "Remote port")
	remoteAddCmd.Flags().String("environment", "staging", "Remote environment (dev, staging, prod)")
	remoteAddCmd.Flags().String("capabilities", "files,media,db,deploy", "Remote capabilities (files,media,db,deploy or all)")
	remoteAddCmd.Flags().String("auth-method", remote.AuthMethodKeychain, "Remote auth method (keychain, ssh-agent, keyfile)")
	remoteAddCmd.Flags().String("key-path", "", "SSH private key path (stored in auth store when --auth-method=keychain)")
	remoteAddCmd.Flags().Bool("strict-host-key", false, "Enable strict SSH host key checking")
	remoteAddCmd.Flags().String("known-hosts-file", "", "Custom SSH known_hosts file (implies --strict-host-key)")
	remoteAddCmd.Flags().Bool("protected", false, "Mark remote as protected")

	remoteCmd.AddCommand(remoteAddCmd)
	remoteCmd.AddCommand(remoteExecCmd)
	remoteCmd.AddCommand(remoteTestCmd)
	remoteCmd.AddCommand(remoteAuditCmd)

	rootCmd.AddCommand(remoteCmd)
}

// RemoteCommand exposes the remote command for testing.
func RemoteCommand() *cobra.Command {
	return remoteCmd
}

// RootCommandForTest exposes the root command for tests.
func RootCommandForTest() *cobra.Command {
	return rootCmd
}

func ensureRemoteKnown(config engine.Config, name string) (engine.RemoteConfig, error) {
	remote, ok := config.Remotes[name]
	if !ok {
		return engine.RemoteConfig{}, fmt.Errorf("unknown remote: %s", name)
	}
	return remote, nil
}

func reportRemoteCommandFailure(step string, err error, output []byte, duration time.Duration, warning bool) remote.FailureDetails {
	details := remote.ClassifyFailure(err, string(output))
	message := fmt.Sprintf("%s failed (%s, %s): %v", step, details.Category, formatDuration(duration), err)
	if warning {
		pterm.Warning.Println(message)
	} else {
		pterm.Error.Println(message)
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed != "" {
		if warning {
			pterm.Warning.Println(trimmed)
		} else {
			pterm.Error.Println(trimmed)
		}
	}

	if details.Hint != "" {
		if warning {
			pterm.Warning.Printf("Hint: %s\n", details.Hint)
		} else {
			pterm.Error.Printf("Hint: %s\n", details.Hint)
		}
	}
	return details
}

func formatDuration(duration time.Duration) string {
	return duration.Round(time.Millisecond).String()
}
