//go:build !windows

package desktop

import (
	"os"
	"os/exec"
	"syscall"
)

func setSysProcAttrForDetach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	// Fully detach stdio so the child does not receive SIGPIPE or
	// get torn down when the parent Wails process quits.
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err == nil {
		cmd.Stdin = devNull
		cmd.Stdout = devNull
		cmd.Stderr = devNull
	}
}
