package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/neongreen/mono/claim/internal/logger"
)

// ClaimResult is the result of checking a claim
type ClaimResult struct {
	ClaimID        string          `json:"claim_id"`
	Result         string          `json:"result"` // "proven", "unproven", "sorry"
	Bullets        []BulletVerdict `json:"bullets"`
	Counterexample string          `json:"counterexample"`
}

// BulletVerdict is Claude's verdict on a single bullet
type BulletVerdict struct {
	Path             string   `json:"path"`
	Status           string   `json:"status"` // "ok", "trivial", "needs_split", "needs_claim", "contradicts", "sorry"
	RequiredClaims   []string `json:"required_claims"`
	SuggestedRewrite []string `json:"suggested_rewrite"`
}

// Runner interface for checking claims
type Runner interface {
	Run(ctx context.Context, prompt string) (*ClaimResult, error)
}

// ClaudeRunner runs Claude CLI to check claims
type ClaudeRunner struct {
	Command string // Path to claude command (default "claude")
	Model   string // Optional model override
	Verbose bool   // Show streaming output (thinking tags, etc.)
}

// Run executes Claude with the given prompt and returns the structured result
func (r *ClaudeRunner) Run(ctx context.Context, prompt string) (*ClaimResult, error) {
	logger.Debug("calling claude", "command", r.Command, "model", r.Model, "prompt_length", len(prompt))

	// Build JSON schema for structured output
	schema := buildSchema()
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Always use stream-json to get structured output
	// We need --verbose --print to get output with stream-json format
	// In verbose mode, we also show it to the user via MultiWriter
	args := []string{
		"--verbose",
		"--print",
		"--output-format", "stream-json",
		"--json-schema", string(schemaJSON),
		"--tools", "",                             // Disable all tools
		"--setting-sources", "project",            // Skip user settings (only load project settings)
		"--settings", `{"disableAllHooks": true}`, // Disable all hooks
	}

	if r.Model != "" {
		args = append(args, "--model", r.Model)
	}

	cmd := exec.CommandContext(ctx, r.Command, args...)

	// Preserve API key in environment
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY="+os.Getenv("ANTHROPIC_API_KEY"),
	)

	cmd.Stdin = strings.NewReader(prompt)

	// Always capture output; in verbose mode also stream to stdout/stderr
	var stdout, stderr bytes.Buffer
	if r.Verbose {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("claude command failed: %w\nStderr: %s", err, stderr.String())
	}

	logger.Debug("claude responded", "output_length", len(stdout.Bytes()))

	// Parse stream-json format: newline-delimited JSON
	// Find the last line with structured_output
	var response struct {
		StructuredOutput json.RawMessage `json:"structured_output"`
	}

	lines := bytes.Split(stdout.Bytes(), []byte("\n"))
	for i := len(lines) - 1; i >= 0; i-- {
		if len(lines[i]) == 0 {
			continue
		}
		var event map[string]json.RawMessage
		if err := json.Unmarshal(lines[i], &event); err != nil {
			continue
		}
		if structOut, ok := event["structured_output"]; ok {
			response.StructuredOutput = structOut
			break
		}
	}

	if response.StructuredOutput == nil {
		return nil, fmt.Errorf("claude did not return structured output")
	}

	// Parse claim result
	var result ClaimResult
	if err := json.Unmarshal(response.StructuredOutput, &result); err != nil {
		return nil, fmt.Errorf("invalid claim result structure: %w", err)
	}

	logger.Debug("parsed result", "claim_id", result.ClaimID, "result", result.Result, "bullets", len(result.Bullets))

	return &result, nil
}

// buildSchema creates the JSON schema for Claude's structured output
func buildSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"claim_id": map[string]any{
				"type": "string"},
			"result": map[string]any{
				"type": "string",
				"enum": []string{"proven", "unproven", "sorry"},
			},
			"bullets": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type": "string",
						},
						"status": map[string]any{
							"type": "string",
							"enum": []string{"ok", "trivial", "needs_split", "needs_claim", "contradicts", "sorry"},
						},
						"required_claims": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "string",
							},
						},
						"suggested_rewrite": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "string",
							},
						},
					},
					"required": []string{"path", "status"},
				},
			},
			"counterexample": map[string]any{
				"type": "string",
			},
		},
		"required": []string{"claim_id", "result", "bullets"},
	}
}
