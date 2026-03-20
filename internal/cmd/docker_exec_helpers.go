package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func dockerExecBaseArgs() []string {
	// If we are running in an integration test, always use non-interactive mode
	if os.Getenv("GOVARD_TEST_RUNTIME") == "true" {
		return []string{"exec", "-i"}
	}

	// Check if we have a TTY for stdin/stdout
	if isTerminal(os.Stdin) && isTerminal(os.Stdout) {
		return []string{"exec", "-it"}
	}
	return []string{"exec", "-i"}
}

func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func ensureContainerReadyForExec(containerName string, serviceLabel string) error {
	inspect := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
	output, err := inspect.CombinedOutput()
	state := strings.ToLower(strings.TrimSpace(string(output)))

	if state == "true" {
		return nil
	}
	if state == "false" {
		return fmt.Errorf(
			"%s container %s is stopped. Run `govard env up` (or `govard env restart`) and retry",
			serviceLabel,
			containerName,
		)
	}

	if err != nil {
		return fmt.Errorf(
			"%s container %s is unknown. Run `govard env up` (or `govard env restart`) and retry",
			serviceLabel,
			containerName,
		)
	}

	return fmt.Errorf(
		"%s container %s is %s. Run `govard env up` (or `govard env restart`) and retry",
		serviceLabel,
		containerName,
		state,
	)
}
