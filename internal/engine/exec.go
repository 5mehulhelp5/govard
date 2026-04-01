package engine

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Handoff replaces the current process with the specified binary and arguments.
// This is typically used for interactive shells where we want the terminal
// state and signals to be managed directly by the child process.
func Handoff(binaryPath string, args []string) error {
	// args[0] should be the binary name as seen by the new process
	return syscall.Exec(binaryPath, args, os.Environ())
}

// ExecuteInteractively runs a command with its stdio connected to the current terminal.
// This is used when we don't want to replace the current process.
func ExecuteInteractively(cmd *exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ShellQuote wraps a string in single quotes and escapes existing single quotes
// so it is safe to use in a shell command.
func ShellQuote(raw string) string {
	if raw == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(raw, "'", `'"'"'`) + "'"
}
