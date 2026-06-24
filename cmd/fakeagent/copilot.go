package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	copilotPromptWrapperPrefix = "Read the complete no-mistakes task instructions from this file,"
	copilotPromptFileMarker    = "File: "
)

func runCopilot(args []string, scenario *Scenario) int {
	prompt := extractCopilotPrompt(args)
	var err error
	prompt, err = resolveCopilotPrompt(prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fakeagent: copilot prompt: %v\n", err)
		return 1
	}
	logInvocation("copilot", prompt, args)

	action := scenario.Match(prompt)
	if err := applyAction(action); err != nil {
		return 1
	}
	if action.Structured != nil || action.StructuredRaw != "" {
		fmt.Fprintln(os.Stdout, action.structuredJSON())
		return 0
	}
	fmt.Fprintln(os.Stdout, action.textOrDefault())
	return 0
}

func extractCopilotPrompt(args []string) string {
	for i, arg := range args {
		switch {
		case arg == "-p" || arg == "--prompt":
			if i+1 < len(args) {
				return args[i+1]
			}
		case strings.HasPrefix(arg, "--prompt="):
			return arg[len("--prompt="):]
		}
	}
	return ""
}

func resolveCopilotPrompt(prompt string) (string, error) {
	path, ok := referencedCopilotPromptFile(prompt)
	if !ok {
		return prompt, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read referenced prompt file %q: %w", path, err)
	}
	return string(data), nil
}

func referencedCopilotPromptFile(prompt string) (string, bool) {
	if !strings.HasPrefix(prompt, copilotPromptWrapperPrefix) {
		return "", false
	}
	idx := strings.LastIndex(prompt, copilotPromptFileMarker)
	if idx < 0 {
		return "", false
	}
	return strings.TrimSpace(prompt[idx+len(copilotPromptFileMarker):]), true
}
