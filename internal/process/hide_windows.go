//go:build windows

package process

import (
	"os/exec"
	"syscall"
)

// HideWindow prevents short-lived subprocesses from flashing a console window
// and stealing focus on Windows.
func HideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}
