package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCopilotReadsReferencedPromptFileForMatchingAndLogging(t *testing.T) {
	dir := t.TempDir()
	promptPath := filepath.Join(dir, "prompt.md")
	if err := os.WriteFile(promptPath, []byte("full pipeline prompt"), 0o644); err != nil {
		t.Fatalf("write prompt: %v", err)
	}
	logPath := filepath.Join(dir, "fakeagent.log")
	t.Setenv("FAKEAGENT_LOG", logPath)

	scenario := &Scenario{Actions: []Action{
		{Match: "full pipeline prompt", Text: "matched full prompt"},
		{Text: "fallback"},
	}}
	args := []string{"-p", copilotWrapperPrompt(promptPath)}

	var code int
	stdout := captureStdout(t, func() {
		code = runCopilot(args, scenario)
	})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if got := strings.TrimSpace(stdout); got != "matched full prompt" {
		t.Fatalf("stdout = %q, want matched full prompt", got)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	var inv invocation
	if err := json.Unmarshal(data, &inv); err != nil {
		t.Fatalf("unmarshal invocation: %v", err)
	}
	if inv.Prompt != "full pipeline prompt" {
		t.Fatalf("logged prompt = %q, want full prompt", inv.Prompt)
	}
}

func copilotWrapperPrompt(path string) string {
	return "Read the complete no-mistakes task instructions from this file, execute them, and make your final response satisfy any output contract in that file. Do not ask clarification questions. File: " + filepath.Clean(path)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
	return string(out)
}
