package cli

import (
	"strings"
	"testing"
)

func TestGateCommandRegistered(t *testing.T) {
	out, err := executeCmd("gate", "--help")
	if err != nil {
		t.Fatalf("gate --help: %v\n%s", err, out)
	}
	for _, want := range []string{
		"Headless gate interface",
		"run",
		"respond",
		"status",
		"logs",
		"cancel",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("gate help missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "axi") || strings.Contains(out, "AXI") {
		t.Fatalf("gate help should not mention AXI:\n%s", out)
	}
}

func TestGateSurfaceHelpUsesGateCommands(t *testing.T) {
	gate := stepView{
		Name:         "review",
		Status:       "awaiting_approval",
		FindingsJSON: findingsJSON(t, nil, "needs review"),
	}

	out := axiDoc(gateFields(gate, gateSurface)...)
	for _, want := range []string{
		"no-mistakes gate respond --action approve",
		"no-mistakes gate respond --action fix --findings <ids>",
		"no-mistakes gate respond --action skip",
		"no-mistakes gate logs --step review --full",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("gate surface help missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "no-mistakes axi") {
		t.Fatalf("gate surface help should not mention AXI:\n%s", out)
	}
}

func TestGateStartRunHelpUsesGateCommand(t *testing.T) {
	got := startRunHelp(gateSurface)
	if !strings.Contains(got, `no-mistakes gate run --intent "the user's goal" --yes`) {
		t.Fatalf("gate start help uses wrong command: %q", got)
	}
}
