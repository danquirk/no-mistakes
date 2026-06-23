package agent

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCopilotBuildArgs(t *testing.T) {
	a := &copilotAgent{extraArgs: []string{"--model", "gpt-5.4"}}
	args := a.buildArgs("read prompt file")
	joined := strings.Join(args, "\x00")

	for _, want := range []string{
		"--model\x00gpt-5.4",
		"-p\x00read prompt file",
		"--allow-all",
		"--silent",
		"--no-ask-user",
		"--no-auto-update",
		"--no-remote",
		"--no-remote-export",
		"--stream\x00off",
		"--no-color",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("args = %q, missing %q", args, want)
		}
	}
}

func TestShortCopilotPromptReferencesPromptFile(t *testing.T) {
	got := shortCopilotPrompt(`C:\Temp\prompt.md`)
	for _, want := range []string{
		"complete no-mistakes task instructions",
		"final response satisfy any output contract",
		`C:\Temp\prompt.md`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("short prompt missing %q in:\n%s", want, got)
		}
	}
}

func TestBuildCopilotPromptAddsSchemaContract(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"summary":{"type":"string"}}}`)
	got := buildCopilotPrompt("review code", schema)
	for _, want := range []string{
		"review code",
		"final output contract",
		"valid JSON matching this JSON Schema",
		`"summary"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("prompt missing %q in:\n%s", want, got)
		}
	}
}

func TestCleanCopilotTextStripsPromptModeBullet(t *testing.T) {
	if got := cleanCopilotText("  ● OK\n• done\n"); got != "OK\ndone" {
		t.Fatalf("cleanCopilotText() = %q, want cleaned lines", got)
	}
}

func TestFinalizeCopilotResultRepairsWrappedJSONStrings(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"findings":{"type":"array","items":{"type":"object","properties":{"description":{"type":"string"}}}}},"required":["findings"]}`)
	text := "notes\n● {\"findings\":[{\"description\":\"wrapped\n  output\"}]}"
	result, err := finalizeCopilotResult(cleanCopilotText(text), schema)
	if err != nil {
		t.Fatalf("finalizeCopilotResult() error = %v", err)
	}
	if !strings.Contains(string(result.Output), "wrapped") {
		t.Fatalf("output = %s, want repaired JSON", result.Output)
	}
	if got := result.Text; !strings.Contains(got, "wrapped\noutput") {
		t.Fatalf("Text = %q, want original text preserved", got)
	}
}

func TestRepairCopilotWrappedJSONNoopsWithoutObject(t *testing.T) {
	if got := repairCopilotWrappedJSON("plain text"); got != "plain text" {
		t.Fatalf("cleanCopilotText() = %q, want OK", got)
	}
}
