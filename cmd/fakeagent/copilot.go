package main

import (
	"fmt"
	"os"
	"strings"
)

func runCopilot(args []string, scenario *Scenario) int {
	prompt := extractCopilotPrompt(args)
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
