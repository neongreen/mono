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
	ClaimID        string           `json:"claim_id"`
	Result         string           `json:"result"` // "proven", "unproven", "sorry"
	Bullets        []BulletVerdict  `json:"bullets"`
	Counterexample string           `json:"counterexample"`
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

	// In verbose mode, stream output directly; otherwise use --print for quiet operation
	var output []byte

	if r.Verbose {
		// Verbose mode: use stream-json for real-time streaming
		cmd := exec.CommandContext(ctx, r.Command,
			"--verbose",
			"--output-format", "stream-json",
			"--json-schema", string(schemaJSON),
			"--tools", "", // Disable tools
		)

		if r.Model != "" {
			cmd.Args = append(cmd.Args, "--model", r.Model)
		}

		cmd.Stdin = strings.NewReader(prompt)

		// Stream to stdout/stderr AND capture for parsing
		var stdout, stderr bytes.Buffer
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("claude command failed: %w\nStderr: %s", err, stderr.String())
		}

		output = stdout.Bytes()
	} else{
		// Normal mode: quiet operation with --print and --output-format json
		cmd := exec.CommandContext(ctx, r.Command,
			"--print",
			"--output-format", "json",
			"--json-schema", string(schemaJSON),
			"--tools", "", // Disable tools
		)

		if r.Model != "" {
			cmd.Args = append(cmd.Args, "--model", r.Model)
		}

		cmd.Stdin = strings.NewReader(prompt)

		var err error
		output, err = cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("claude command failed: %w\nOutput: %s", err, output)
		}
	}

	logger.Debug("claude responded", "output_length", len(output))

	// Parse response
	var response struct {
		StructuredOutput json.RawMessage `json:"structured_output"`
	}

	if r.Verbose {
		// stream-json format: newline-delimited JSON, find the last line with structured_output
		lines := bytes.Split(output, []byte("\n"))
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
	} else {
		// Normal JSON format
		if err := json.Unmarshal(output, &response); err != nil {
			return nil, fmt.Errorf("invalid JSON from claude: %w", err)
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
