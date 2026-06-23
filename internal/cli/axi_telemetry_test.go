package cli

import (
	"testing"

	"github.com/kunchenguid/no-mistakes/internal/telemetry"
)

func TestAxiCommandsEmitPageviews(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		path    string
		command string
	}{
		{"home", []string{"axi"}, "/axi", "axi-home"},
		{"run", []string{"axi", "run", "--intent", "ship the thing"}, "/axi/run", "axi-run"},
		{"respond", []string{"axi", "respond", "--action", "approve"}, "/axi/respond", "axi-respond"},
		{"status", []string{"axi", "status"}, "/axi/status", "axi-status"},
		{"logs", []string{"axi", "logs", "--step", "review"}, "/axi/logs", "axi-logs"},
		{"abort", []string{"axi", "abort"}, "/axi/abort", "axi-abort"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertHeadlessTelemetry(t, tc.args, tc.path, tc.command)
		})
	}
}

func TestGateCommandsEmitPageviews(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		path    string
		command string
	}{
		{"home", []string{"gate"}, "/gate", "gate-home"},
		{"run", []string{"gate", "run", "--intent", "ship the thing"}, "/gate/run", "gate-run"},
		{"respond", []string{"gate", "respond", "--action", "approve"}, "/gate/respond", "gate-respond"},
		{"status", []string{"gate", "status"}, "/gate/status", "gate-status"},
		{"logs", []string{"gate", "logs", "--step", "review"}, "/gate/logs", "gate-logs"},
		{"cancel", []string{"gate", "cancel"}, "/gate/cancel", "gate-cancel"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertHeadlessTelemetry(t, tc.args, tc.path, tc.command)
		})
	}
}

func assertHeadlessTelemetry(t *testing.T, args []string, path, command string) {
	t.Helper()

	tmpDir := t.TempDir()
	t.Setenv("NM_HOME", t.TempDir())
	chdir(t, tmpDir)

	recorder := &telemetryRecorder{}
	restore := telemetry.SetDefaultForTesting(recorder)
	defer restore()

	_, _ = executeCmd(args...)

	if event := recorder.find("pageview", "path", path); event == nil {
		t.Fatalf("expected %s pageview for %v", path, args)
	}
	if event := recorder.find("command", "command", command); event == nil {
		t.Fatalf("expected %s command event alongside the pageview", command)
	}
}

// TestAxiRunPageviewCarriesFlags verifies the run pageview includes the
// flag-derived context an analytics surface can segment on, mirroring how the
// TUI pageview carries entrypoint/run_status.
func TestAxiRunPageviewCarriesFlags(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("NM_HOME", t.TempDir())
	chdir(t, tmpDir)

	recorder := &telemetryRecorder{}
	restore := telemetry.SetDefaultForTesting(recorder)
	defer restore()

	_, _ = executeCmd("axi", "run", "--intent", "ship it", "--yes", "--skip", "lint")

	event := recorder.find("pageview", "path", "/axi/run")
	if event == nil {
		t.Fatal("expected /axi/run pageview")
	}
	if got := event.fields["auto_yes"]; got != true {
		t.Fatalf("auto_yes = %v, want true", got)
	}
	if got := event.fields["has_intent"]; got != true {
		t.Fatalf("has_intent = %v, want true", got)
	}
	if got := event.fields["has_skip"]; got != true {
		t.Fatalf("has_skip = %v, want true", got)
	}
}

// TestAxiLogsPageviewCarriesStep verifies the logs pageview records which step
// and whether a specific run was requested.
func TestAxiLogsPageviewCarriesStep(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("NM_HOME", t.TempDir())
	chdir(t, tmpDir)

	recorder := &telemetryRecorder{}
	restore := telemetry.SetDefaultForTesting(recorder)
	defer restore()

	_, _ = executeCmd("axi", "logs", "--step", "test", "--run", "run-123")

	event := recorder.find("pageview", "path", "/axi/logs")
	if event == nil {
		t.Fatal("expected /axi/logs pageview")
	}
	if got := event.fields["step"]; got != "test" {
		t.Fatalf("step = %v, want test", got)
	}
	if got := event.fields["explicit_run_id"]; got != true {
		t.Fatalf("explicit_run_id = %v, want true", got)
	}
}

func TestAxiLogsPageviewSanitizesInvalidStep(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("NM_HOME", t.TempDir())
	chdir(t, tmpDir)

	recorder := &telemetryRecorder{}
	restore := telemetry.SetDefaultForTesting(recorder)
	defer restore()

	_, _ = executeCmd("axi", "logs", "--step", "secret user text")

	event := recorder.find("pageview", "path", "/axi/logs")
	if event == nil {
		t.Fatal("expected /axi/logs pageview")
	}
	if got := event.fields["step"]; got != "invalid" {
		t.Fatalf("step = %v, want invalid", got)
	}
}

func TestAxiRespondPageviewSanitizesInvalidAction(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("NM_HOME", t.TempDir())
	chdir(t, tmpDir)

	recorder := &telemetryRecorder{}
	restore := telemetry.SetDefaultForTesting(recorder)
	defer restore()

	_, _ = executeCmd("axi", "respond", "--action", "secret user text")

	event := recorder.find("pageview", "path", "/axi/respond")
	if event == nil {
		t.Fatal("expected /axi/respond pageview")
	}
	if got := event.fields["action"]; got != "invalid" {
		t.Fatalf("action = %v, want invalid", got)
	}
}
