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

// Config holds the settings the provider needs.
type Config struct {
	Bin     string        // path to the claude CLI
	Model   string        // model alias or id; empty uses the account default
	Timeout time.Duration // per run
	Tools   []string      // built-in tools Claude may use
}

// Provider runs prompts through the Claude Code CLI.
type Provider struct {
	cfg Config
}

// New creates a Provider, filling in defaults for empty config values.
func New(cfg Config) *Provider {
	if cfg.Bin == "" {
		cfg.Bin = "claude"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Minute
	}
	if cfg.Tools == nil {
		cfg.Tools = []string{"WebSearch", "WebFetch"}
	}
	return &Provider{cfg: cfg}
}

// Run sends prompt to Claude and asks for output matching jsonSchema.
func (p *Provider) Run(ctx context.Context, prompt, jsonSchema string) (*CLIResult, error) {
	ctx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	tools := strings.Join(p.cfg.Tools, ",")
	args := []string{
		"-p",
		"--output-format", "json",
		"--json-schema", jsonSchema,
		"--tools", tools,
		"--allowedTools", tools,
		"--no-session-persistence",
		"--strict-mcp-config",
	}
	if p.cfg.Model != "" {
		args = append(args, "--model", p.cfg.Model)
	}

	cmd := exec.CommandContext(ctx, p.cfg.Bin, args...)
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
