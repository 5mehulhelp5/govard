package remote

import (
	"strconv"

	"govard/internal/engine"
)

func BuildSSHArgs(remoteName string, remoteCfg engine.RemoteConfig, forwardAgent bool) []string {
	args := []string{
		"-o", "LogLevel=ERROR",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
	}
	if remoteCfg.Auth.StrictHostKey {
		args = append(args, "-o", "StrictHostKeyChecking=yes")
		if remoteCfg.Auth.KnownHostsFile != "" {
			args = append(args, "-o", "UserKnownHostsFile="+remoteCfg.Auth.KnownHostsFile)
		}
	} else {
		args = append(args, "-o", "StrictHostKeyChecking=no")
		args = append(args, "-o", "UserKnownHostsFile=/dev/null")
	}
	if remoteCfg.Port > 0 {
		args = append(args, "-p", strconv.Itoa(remoteCfg.Port))
	}
	if forwardAgent {
		args = append(args, "-A")
	}
	if keyPath, _ := ResolveSSHKeyPath(remoteName, remoteCfg); keyPath != "" {
		args = append(args, "-i", keyPath)
	}
	return args
}
