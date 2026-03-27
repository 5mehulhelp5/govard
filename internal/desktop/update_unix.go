//go:build !windows

package desktop

import (
	"os/exec"
	"syscall"
)

func setSysProcAttrForDetach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}
