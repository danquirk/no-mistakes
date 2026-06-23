package cli

import "github.com/kunchenguid/no-mistakes/internal/telemetry"

type headlessSurface struct {
	name  string
	short string
	long  string
	// cancelSub is the cancel subcommand verb (e.g. "abort", "cancel").
	cancelSub string
	// cancelledKey is the output key that reports the cancel result, kept in the
	// surface's own vocabulary so `gate cancel` reports `cancelled` while `axi
	// abort` reports `aborted`.
	cancelledKey string
}

var (
	axiSurface = headlessSurface{
		name:  "axi",
		short: "Agent interface: drive no-mistakes from an autonomous agent",
		long: "Agent eXperience Interface for no-mistakes. Prints token-efficient TOON\n" +
			"to stdout and is driven entirely by flags (no interactive prompts).\n" +
			"Running `no-mistakes axi` with no subcommand shows the current state.",
		cancelSub:    "abort",
		cancelledKey: "aborted",
	}
	gateSurface = headlessSurface{
		name:  "gate",
		short: "Headless gate interface: validate and respond without the TUI",
		long: "Headless gate interface for no-mistakes. Prints machine-readable output\n" +
			"to stdout and is driven entirely by flags (no interactive prompts).\n" +
			"Running `no-mistakes gate` with no subcommand shows the current state.",
		cancelSub:    "cancel",
		cancelledKey: "cancelled",
	}
)

func (s headlessSurface) command(sub string) string {
	if sub == "" {
		return "no-mistakes " + s.name
	}
	return "no-mistakes " + s.name + " " + sub
}

func (s headlessSurface) telemetryName(sub string) string {
	if sub == "" {
		return s.name + "-home"
	}
	return s.name + "-" + sub
}

func (s headlessSurface) telemetryPath(sub string) string {
	if sub == "" {
		return "/" + s.name
	}
	return "/" + s.name + "/" + sub
}

func trackHeadlessSurface(surface headlessSurface, sub string, fields telemetry.Fields, fn func() error) error {
	telemetry.Pageview(surface.telemetryPath(sub), fields)
	return trackCommand(surface.telemetryName(sub), fn)
}
