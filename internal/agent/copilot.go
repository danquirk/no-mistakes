package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
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
	args := a.buildArgs(buildCopilotPrompt(opts.Prompt, opts.JSONSchema))
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
	return finalizeTextResult("copilot", text, opts.JSONSchema, TokenUsage{})
}

func (a *copilotAgent) buildArgs(prompt string) []string {
	args := make([]string, 0, len(a.extraArgs)+12)
	args = append(args, a.extraArgs...)
	args = append(args,
		"-p", prompt,
		"--allow-all",
		"--silent",
		"--no-auto-update",
		"--no-remote",
		"--no-remote-export",
		"--log-level", "error",
		"--stream", "off",
		"--no-color",
	)
	return args
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

func cleanCopilotText(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "● ")
	text = strings.TrimPrefix(text, "• ")
	return strings.TrimSpace(text)
}
