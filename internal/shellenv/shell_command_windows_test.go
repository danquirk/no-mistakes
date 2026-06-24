//go:build windows

package shellenv

import (
	"errors"
	"os/exec"
	"strconv"
	"syscall"
	"testing"
)

// TestIsTaskkillAlreadyGone pins down the locale-independent contract the
// Windows cancel path relies on: taskkill exit code 128 (no matching PID) is
// the only nonzero exit treated as "the child already exited", so every other
// nonzero code falls through to the direct-child-kill backstop instead of
// being swallowed as os.ErrProcessDone. It runs only on Windows; on Linux the
// windows build tag excludes it from `go test ./...`, while `GOOS=windows go
// vet` keeps it compile-checked.
func TestIsTaskkillAlreadyGone(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "exit 128 is already gone", err: exitCodeErr(t, 128), want: true},
		{name: "exit 1 access denied is a real failure", err: exitCodeErr(t, 1), want: false},
		{name: "exec.ErrNotFound is not already-gone", err: exec.ErrNotFound, want: false},
		{name: "wrapped exit 128 still detected", err: wrapErr(exitCodeErr(t, 128)), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTaskkillAlreadyGone(tt.err); got != tt.want {
				t.Fatalf("isTaskkillAlreadyGone(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestConfigureShellCommandHidesWindow(t *testing.T) {
	cmd := exec.Command("does-not-need-to-exist.exe")
	ConfigureShellCommand(cmd)
	if cmd.SysProcAttr == nil {
		t.Fatal("ConfigureShellCommand did not assign SysProcAttr")
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Fatal("HideWindow = false, want true")
	}
}

// exitCodeErr runs `cmd /c exit N` and returns the resulting *exec.ExitError so
// the helper is exercised against a real ProcessState with the chosen code.
func exitCodeErr(t *testing.T, code int) error {
	t.Helper()
	cmd := exec.Command("cmd", "/c", "exit", strconv.Itoa(code))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected exit %d to yield a nonzero-run error", code)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode() != code {
		t.Fatalf("exit code = %d, want %d", exitErr.ExitCode(), code)
	}
	return err
}

type wrappedErr struct{ e error }

func (w wrappedErr) Error() string { return "wrapped: " + w.e.Error() }
func (w wrappedErr) Unwrap() error { return w.e }

func wrapErr(e error) error { return wrappedErr{e: e} }
