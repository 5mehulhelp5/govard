package engine

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	HookPreUp         = "pre_up"
	HookPostUp        = "post_up"
	HookPreStop       = "pre_stop"
	HookPostStop      = "post_stop"
	HookPreSync       = "pre_sync"
	HookPostSync      = "post_sync"
	HookPreDeploy     = "pre_deploy"
	HookPostDeploy    = "post_deploy"
	HookPreDBConnect  = "pre_db_connect"
	HookPostDBConnect = "post_db_connect"
	HookPreDBImport   = "pre_db_import"
	HookPostDBImport  = "post_db_import"
	HookPreDBDump     = "pre_db_dump"
	HookPostDBDump    = "post_db_dump"
	HookPreDelete     = "pre_delete"
	HookPostDelete    = "post_delete"
)

var allowedHookEvents = map[string]struct{}{
	HookPreUp:         {},
	HookPostUp:        {},
	HookPreStop:       {},
	HookPostStop:      {},
	HookPreSync:       {},
	HookPostSync:      {},
	HookPreDeploy:     {},
	HookPostDeploy:    {},
	HookPreDBConnect:  {},
	HookPostDBConnect: {},
	HookPreDBImport:   {},
	HookPostDBImport:  {},
	HookPreDBDump:     {},
	HookPostDBDump:    {},
	HookPreDelete:     {},
	HookPostDelete:    {},
}

func RunHooks(config Config, event string, stdout, stderr io.Writer) error {
	steps := config.Hooks[event]
	if len(steps) == 0 {
		return nil
	}

	wd, _ := os.Getwd()
	for idx, step := range steps {
		command := strings.TrimSpace(step.Run)
		if command == "" {
			continue
		}

		cmd := exec.Command("bash", "-lc", command)
		cmd.Dir = wd
		cmd.Env = append(os.Environ(),
			"GOVARD_HOOK_EVENT="+event,
			"GOVARD_PROJECT_NAME="+config.ProjectName,
			"GOVARD_DOMAIN="+config.Domain,
			"GOVARD_FRAMEWORK="+config.Framework,
		)
		if stdout != nil {
			cmd.Stdout = stdout
		}
		if stderr != nil {
			cmd.Stderr = stderr
		}

		if err := cmd.Run(); err != nil {
			label := step.Name
			if strings.TrimSpace(label) == "" {
				label = fmt.Sprintf("step #%d", idx+1)
			}
			return fmt.Errorf("hook %s (%s) failed: %w", event, label, err)
		}
	}

	return nil
}
