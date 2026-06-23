//go:build !windows

package process

import "os/exec"

// HideWindow suppresses visible child-process windows on platforms that support
// that concept. It is a no-op outside Windows.
func HideWindow(cmd *exec.Cmd) {}
