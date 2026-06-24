package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kunchenguid/no-mistakes/internal/shellenv"
)

// copilotAgent spawns GitHub Copilot CLI for each invocation. Copilot supports
// non-interactive prompt mode with -p/--prompt; structured output is requested
// through the same inline final-output contract used by text-only backends.
type copilotAgent struct {
	bin       string
	extraArgs []string
}

func (a *copilotAgent) Name() string { return "copilot" }

func (a *copilotAgent) Run(ctx context.Context, opts RunOpts) (*Result, error) {
	return runWithRetry(ctx, "copilot", opts, claudeMaxRetries, classifyTransient, nil, func() (*Result, error) {
		return a.runOnce(ctx, opts)
	})
}

func (a *copilotAgent) Close() error { return nil }

func (a *copilotAgent) runOnce(ctx context.Context, opts RunOpts) (*Result, error) {
	promptPath, err := writeCopilotPromptFile(opts.Prompt, opts.JSONSchema)
	if err != nil {
		return nil, err
	}
	defer os.Remove(promptPath)

	args := a.buildArgs(shortCopilotPrompt(promptPath))
	cmd := exec.CommandContext(ctx, a.bin, args...)
	cmd.Dir = opts.CWD
	cmd.Env = gitSafeEnv(opts.CWD)
	shellenv.ConfigureShellCommand(cmd)

	out, err := cmd.CombinedOutput()
	text := cleanCopilotText(string(out))
	if err != nil {
		if text != "" {
			return nil, fmt.Errorf("copilot exited: %w: %s", err, text)
		}
		return nil, fmt.Errorf("copilot exited: %w", err)
	}
	if opts.OnChunk != nil && text != "" {
		opts.OnChunk(text)
	}
	return finalizeCopilotResult(text, opts.JSONSchema)
}

func (a *copilotAgent) buildArgs(prompt string) []string {
	args := make([]string, 0, len(a.extraArgs)+12)
	args = append(args, a.extraArgs...)
	args = append(args,
		"-p", prompt,
		"--allow-all",
		"--silent",
		"--no-ask-user",
		"--no-auto-update",
		"--no-remote",
		"--no-remote-export",
		"--log-level", "error",
		"--stream", "off",
		"--no-color",
	)
	return args
}

func writeCopilotPromptFile(prompt string, schema json.RawMessage) (string, error) {
	f, err := os.CreateTemp("", "no-mistakes-copilot-prompt-*.md")
	if err != nil {
		return "", fmt.Errorf("copilot prompt temp file: %w", err)
	}
	path := f.Name()
	if _, err := f.WriteString(buildCopilotPrompt(prompt, schema)); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("copilot prompt temp file write: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("copilot prompt temp file close: %w", err)
	}
	return path, nil
}

func shortCopilotPrompt(promptPath string) string {
	return "Read the complete no-mistakes task instructions from this file, execute them, and make your final response satisfy any output contract in that file. Do not ask clarification questions. File: " + filepath.Clean(promptPath)
}

func buildCopilotPrompt(prompt string, schema json.RawMessage) string {
	if len(schema) == 0 {
		return prompt
	}
	pretty, err := json.MarshalIndent(json.RawMessage(schema), "", "  ")
	if err != nil {
		pretty = []byte(schema)
	}
	return prompt + "\n\n## no-mistakes final output contract\n\n" +
		"When the iteration is complete, your final assistant response must be only valid JSON matching this JSON Schema. " +
		"Do not wrap it in Markdown fences. Do not include prose before or after the JSON object.\n\n" +
		string(pretty)
}

func finalizeCopilotResult(text string, schema json.RawMessage) (*Result, error) {
	if text == "" {
		return nil, fmt.Errorf("copilot returned no text output")
	}
	if len(schema) == 0 {
		return &Result{Text: text}, nil
	}
	output, err := parseStructuredTextOutput(text, schema)
	if err != nil {
		repaired := repairCopilotWrappedJSON(text)
		if repaired != text {
			if output, repairErr := parseStructuredTextOutput(repaired, schema); repairErr == nil {
				return &Result{Output: output, Text: text}, nil
			}
		}
		return nil, fmt.Errorf("copilot output parse: %w", err)
	}
	return &Result{Output: output, Text: text}, nil
}

func cleanCopilotText(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "● ")
		line = strings.TrimPrefix(line, "• ")
		lines[i] = line
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func repairCopilotWrappedJSON(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return text
	}
	var b strings.Builder
	b.WriteString(text[:start])
	inString := false
	escaped := false
	for _, r := range text[start : end+1] {
		switch {
		case escaped:
			escaped = false
			b.WriteRune(r)
		case r == '\\':
			escaped = true
			b.WriteRune(r)
		case r == '"':
			inString = !inString
			b.WriteRune(r)
		case inString && (r == '\n' || r == '\r'):
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}
	b.WriteString(text[end+1:])
	return b.String()
}
