package cmd

import (
	"fmt"
	"os/exec"
	"strings"
)

func dockerExecBaseArgs() []string {
	return []string{"exec", "-it"}
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
