//go:build windows

package desktop

import "os/exec"

func setSysProcAttrForDetach(cmd *exec.Cmd) {
	// No-op on Windows: SysProcAttr.Setsid is not available.
}
