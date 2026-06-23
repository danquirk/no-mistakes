package agent

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCopilotBuildArgs(t *testing.T) {
	a := &copilotAgent{extraArgs: []string{"--model", "gpt-5.4"}}
	args := a.buildArgs("do work")
	joined := strings.Join(args, "\x00")

	for _, want := range []string{
		"--model\x00gpt-5.4",
		"-p\x00do work",
		"--allow-all",
		"--silent",
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
	if got := cleanCopilotText("  ● OK\n"); got != "OK" {
		t.Fatalf("cleanCopilotText() = %q, want OK", got)
	}
}
