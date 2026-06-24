//go:build windows

package agent

import (
	"os/exec"
	"testing"
)

func TestConfigureManagedServerCmdHidesWindow(t *testing.T) {
	cmd := exec.Command("does-not-need-to-exist.exe")
	configureManagedServerCmd(cmd)
	if cmd.SysProcAttr == nil {
		t.Fatal("configureManagedServerCmd did not assign SysProcAttr")
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Fatal("HideWindow = false, want true")
	}
}
