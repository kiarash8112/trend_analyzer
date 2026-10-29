// Package claude runs analyses through the local Claude Code CLI (`claude -p`),
// so they use the Claude account the CLI is logged in with instead of an API key.
package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CLI settings.
const (
	bin     = "claude"         // the Claude Code CLI, looked up on PATH
	timeout = 10 * time.Minute // per run
)

// tools are the built-in tools Claude may use.
var tools = []string{"WebSearch", "WebFetch"}

// Provider runs prompts through the Claude Code CLI.
type Provider struct {
	model string // model alias or id; empty uses the account default
}

// New creates a Provider. model may be empty.
func New(model string) *Provider {
	return &Provider{model: model}
}

// Run sends prompt to Claude and asks for output matching jsonSchema.
func (p *Provider) Run(ctx context.Context, prompt, jsonSchema string) (*CLIResult, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	allowed := strings.Join(tools, ",")
	args := []string{
		"-p",
		"--output-format", "json",
		"--json-schema", jsonSchema,
		"--tools", allowed,
		"--allowedTools", allowed,
		"--no-session-persistence",
		"--strict-mcp-config",
	}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdin = strings.NewReader(prompt)
	// Run outside the project so its CLAUDE.md and settings don't leak into the analysis.
	cmd.Dir = os.TempDir()
	// Drop any API key so the CLI uses the logged-in Claude subscription.
	cmd.Env = withoutEnv(os.Environ(), "ANTHROPIC_API_KEY")

	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("claude CLI: %w", ctx.Err())
		}
		return nil, fmt.Errorf("claude CLI: %w: %s", err, tail(stderr.String()+stdout.String(), 500))
	}

	var res CLIResult
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		return nil, fmt.Errorf("claude CLI: decode output: %w: %s", err, tail(stdout.String(), 500))
	}
	if res.IsError || res.Subtype != "success" {
		return nil, fmt.Errorf("claude CLI: %s: %s", res.Subtype, tail(res.Result, 500))
	}
	return &res, nil
}

func withoutEnv(env []string, key string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		if !strings.HasPrefix(kv, key+"=") {
			out = append(out, kv)
		}
	}
	return out
}

func tail(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return "..." + s[len(s)-n:]
	}
	return s
}
