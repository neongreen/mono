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
	Result         string          `json:"result"` // "proven", "unproven"
	Bullets        []BulletVerdict `json:"bullets"`
	Counterexample string          `json:"counterexample"`
	HasLater       bool            `json:"-"` // true if any bullet is marked @later (not from Claude, set by post-processing)
}

// BulletVerdict is Claude's verdict on a single bullet
type BulletVerdict struct {
	Path             string   `json:"path"`
	Status           string   `json:"status"` // "ok", "trivial", "needs_split", "incomplete", "unsupported", "contradicts", "later"
	RequiredClaims   []string `json:"required_claims"`
	SuggestedRewrite []string `json:"suggested_rewrite"`
}

// Issue represents a single problem found in a proof step
type Issue struct {
	Title       string `json:"title"`       // Short title, e.g. "Race condition not addressed"
	Description string `json:"description"` // Longer explanation
}

// StepResult is the result of checking a single step in incremental mode
type StepResult struct {
	Status string  `json:"status"` // "ok", "incomplete", "unsupported"
	Issues []Issue `json:"issues"` // Problems found (only when status != ok)
}

// ProofResult is the result of checking a proof using the new @proof syntax
type ProofResult struct {
	Result                 string  `json:"result"`                  // "proven", "unproven"
	Issues                 []Issue `json:"issues"`                  // Problems found (only when result != proven)
	InterpretationNarrowed bool    `json:"interpretation_narrowed"` // True if context changed claim meaning
	Stats                  *Stats  `json:"-"`                       // Execution stats (not from Claude)
}

// Stats contains timing and cost information for a proof check
type Stats struct {
	DurationMs    int64   `json:"duration_ms"`     // Wall clock time in milliseconds
	DurationAPIMs int64   `json:"duration_api_ms"` // API call time in milliseconds
	NumTurns      int     `json:"num_turns"`       // Number of conversation turns
	TotalCostUSD  float64 `json:"total_cost_usd"`  // Total cost in USD
	InputTokens   int64   `json:"input_tokens"`    // Total input tokens
	OutputTokens  int64   `json:"output_tokens"`   // Total output tokens
	CacheRead     int64   `json:"cache_read"`      // Cache read input tokens
	CacheCreation int64   `json:"cache_creation"`  // Cache creation input tokens
}

// Runner interface for checking claims
type Runner interface {
	Run(ctx context.Context, prompt string) (*ClaimResult, error)
	RunStep(ctx context.Context, prompt string) (*StepResult, error)
	RunProof(ctx context.Context, prompt string) (*ProofResult, error)
}

// ClaudeRunner runs Claude CLI to check claims
type ClaudeRunner struct {
	Command string // Path to claude command (default "claude")
	Model   string // Optional model override
	Verbose bool   // Show streaming output (thinking tags, etc.)
	WorkDir string // Working directory for Claude (should be set to --root)
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
		// TODO remove all tools in the no-cheating mode
		"--tools", "Read,Grep,Glob",
		"--setting-sources", "project", // Skip user settings (only load project settings)
		"--settings", `{"disableAllHooks": true}`, // Disable all hooks
	}

	if r.Model != "" {
		args = append(args, "--model", r.Model)
	}

	cmd := exec.CommandContext(ctx, r.Command, args...)

	// Set working directory so Claude can access files relative to --root
	if r.WorkDir != "" {
		cmd.Dir = r.WorkDir
	}

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

// RunStep executes Claude for a single step check in incremental mode
// It will retry up to maxRetries times if Claude doesn't return structured output
func (r *ClaudeRunner) RunStep(ctx context.Context, prompt string) (*StepResult, error) {
	const maxRetries = 2
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			logger.Debug("retrying step check", "attempt", attempt+1, "max", maxRetries+1)
		}

		result, err := r.runStepOnce(ctx, prompt)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Only retry on "no structured output" errors
		if !strings.Contains(err.Error(), "did not return structured output") {
			return nil, err
		}
	}

	return nil, lastErr
}

// runStepOnce executes a single attempt at running Claude for a step check
func (r *ClaudeRunner) runStepOnce(ctx context.Context, prompt string) (*StepResult, error) {
	logger.Debug("calling claude for step", "command", r.Command, "prompt_length", len(prompt))

	// Build JSON schema for step result
	schema := buildStepSchema()
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	args := []string{
		"--verbose",
		"--print",
		"--output-format", "stream-json",
		"--json-schema", string(schemaJSON),
		"--setting-sources", "project",
		"--settings", `{"disableAllHooks": true}`,
	}

	if r.Model != "" {
		args = append(args, "--model", r.Model)
	}

	cmd := exec.CommandContext(ctx, r.Command, args...)

	if r.WorkDir != "" {
		cmd.Dir = r.WorkDir
	}

	cmd.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY="+os.Getenv("ANTHROPIC_API_KEY"),
	)

	cmd.Stdin = strings.NewReader(prompt)

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

	// Parse stream-json format
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

	var result StepResult
	if err := json.Unmarshal(response.StructuredOutput, &result); err != nil {
		return nil, fmt.Errorf("invalid step result structure: %w", err)
	}

	logger.Debug("parsed step result", "status", result.Status)

	return &result, nil
}

// buildSchema creates the JSON schema for structured output
// Note: Codex requires all properties to be in 'required' array and additionalProperties: false
func buildSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"claim_id": map[string]any{
				"type": "string"},
			"result": map[string]any{
				"type": "string",
				"enum": []string{"proven", "unproven"},
			},
			"bullets": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"path": map[string]any{
							"type": "string",
						},
					"status": map[string]any{
						"type": "string",
						"enum": []string{"ok", "trivial", "needs_split", "incomplete", "unsupported", "contradicts"},
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
					"required": []string{"path", "status", "required_claims", "suggested_rewrite"},
				},
			},
			"counterexample": map[string]any{
				"type": "string",
			},
		},
		"required": []string{"claim_id", "result", "bullets", "counterexample"},
	}
}

// buildStepSchema creates the JSON schema for step checking in incremental mode
func buildStepSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"status": map[string]any{
				"type": "string",
				"enum": []string{"ok", "incomplete", "unsupported"},
			},
			"issues": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"title": map[string]any{
							"type": "string",
						},
						"description": map[string]any{
							"type": "string",
						},
					},
					"required": []string{"title", "description"},
				},
			},
		},
		"required": []string{"status", "issues"},
	}
}

// buildProofSchema creates the JSON schema for the new @proof syntax
func buildProofSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"result": map[string]any{
				"type": "string",
				"enum": []string{"proven", "unproven"},
			},
			"issues": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"title": map[string]any{
							"type": "string",
						},
						"description": map[string]any{
							"type": "string",
						},
					},
					"required": []string{"title", "description"},
				},
			},
			"interpretation_narrowed": map[string]any{
				"type":        "boolean",
				"description": "True if the provided context narrows or changes the claim's meaning",
			},
		},
		"required": []string{"result", "issues", "interpretation_narrowed"},
	}
}

// RunProof executes Claude for the new @proof syntax
func (r *ClaudeRunner) RunProof(ctx context.Context, prompt string) (*ProofResult, error) {
	const maxRetries = 2
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			logger.Debug("retrying proof check", "attempt", attempt+1, "max", maxRetries+1)
		}

		result, err := r.runProofOnce(ctx, prompt)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !strings.Contains(err.Error(), "did not return structured output") {
			return nil, err
		}
	}

	return nil, lastErr
}

// runProofOnce executes a single attempt at running Claude for a proof check
func (r *ClaudeRunner) runProofOnce(ctx context.Context, prompt string) (*ProofResult, error) {
	logger.Debug("calling claude for proof", "command", r.Command, "prompt_length", len(prompt))

	schema := buildProofSchema()
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	args := []string{
		"--verbose",
		"--print",
		"--output-format", "stream-json",
		"--json-schema", string(schemaJSON),
		"--tools", "",
		"--setting-sources", "project",
		"--settings", `{"disableAllHooks": true}`,
	}

	if r.Model != "" {
		args = append(args, "--model", r.Model)
	}

	cmd := exec.CommandContext(ctx, r.Command, args...)

	// Run proof checking in an empty temp directory to avoid cache pollution.
	// Proof checking doesn't need filesystem access (tools are disabled) - it only
	// sees the prompt text. Running in a fresh directory ensures we don't get
	// cached responses from previous runs in different codebases.
	tempDir, err := os.MkdirTemp("", "claim-proof-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir for proof check: %w", err)
	}
	defer os.RemoveAll(tempDir)
	cmd.Dir = tempDir
	logger.Debug("running proof check in temp dir", "dir", tempDir)

	cmd.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY="+os.Getenv("ANTHROPIC_API_KEY"),
	)

	cmd.Stdin = strings.NewReader(prompt)

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

	// Always dump full output to temp file for debugging
	dumpClaudeOutput(stdout.Bytes())

	// Parse stream-json format
	var response struct {
		StructuredOutput json.RawMessage `json:"structured_output"`
	}
	var stats *Stats

	lines := bytes.Split(stdout.Bytes(), []byte("\n"))
	for i := len(lines) - 1; i >= 0; i-- {
		if len(lines[i]) == 0 {
			continue
		}
		var event map[string]json.RawMessage
		if err := json.Unmarshal(lines[i], &event); err != nil {
			continue
		}
		if structOut, ok := event["structured_output"]; ok && response.StructuredOutput == nil {
			response.StructuredOutput = structOut
		}
		// Parse stats from result event
		if typeRaw, ok := event["type"]; ok {
			var eventType string
			if err := json.Unmarshal(typeRaw, &eventType); err == nil && eventType == "result" {
				stats = parseStats(lines[i])
			}
		}
	}

	if response.StructuredOutput == nil {
		return nil, fmt.Errorf("claude did not return structured output")
	}

	var result ProofResult
	if err := json.Unmarshal(response.StructuredOutput, &result); err != nil {
		return nil, fmt.Errorf("invalid proof result structure: %w", err)
	}

	result.Stats = stats
	logger.Debug("parsed proof result", "result", result.Result)

	return &result, nil
}

// parseStats extracts stats from a result event
func parseStats(line []byte) *Stats {
	var event struct {
		DurationMs    int64   `json:"duration_ms"`
		DurationAPIMs int64   `json:"duration_api_ms"`
		NumTurns      int     `json:"num_turns"`
		TotalCostUSD  float64 `json:"total_cost_usd"`
		Usage         struct {
			InputTokens           int64 `json:"input_tokens"`
			OutputTokens          int64 `json:"output_tokens"`
			CacheReadInputTokens  int64 `json:"cache_read_input_tokens"`
			CacheCreationTokens   int64 `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(line, &event); err != nil {
		return nil
	}
	return &Stats{
		DurationMs:    event.DurationMs,
		DurationAPIMs: event.DurationAPIMs,
		NumTurns:      event.NumTurns,
		TotalCostUSD:  event.TotalCostUSD,
		InputTokens:   event.Usage.InputTokens,
		OutputTokens:  event.Usage.OutputTokens,
		CacheRead:     event.Usage.CacheReadInputTokens,
		CacheCreation: event.Usage.CacheCreationTokens,
	}
}

// dumpClaudeOutput writes the raw Claude output to a temp file for debugging
func dumpClaudeOutput(output []byte) {
	// Write to a fixed location so it's easy to find
	dumpPath := "/tmp/claim-claude-output.jsonl"
	if err := os.WriteFile(dumpPath, output, 0644); err != nil {
		logger.Debug("failed to dump claude output", "error", err)
		return
	}
	logger.Debug("dumped claude output", "path", dumpPath)
}

// CodexRunner runs codex CLI to check claims
type CodexRunner struct {
	Command string // Path to codex command (default "codex")
	Verbose bool   // Show streaming output
	WorkDir string // Working directory for Codex (should be set to --root)
}

// Run executes codex with the given prompt and returns the structured result
func (r *CodexRunner) Run(ctx context.Context, prompt string) (*ClaimResult, error) {
	logger.Debug("calling codex", "command", r.Command, "prompt_length", len(prompt))

	// Build JSON schema for structured output
	schema := buildSchema()
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Write schema to a temporary file
	schemaFile, err := os.CreateTemp("", "claim-schema-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp schema file: %w", err)
	}
	defer os.Remove(schemaFile.Name())
	defer schemaFile.Close()

	if _, err := schemaFile.Write(schemaJSON); err != nil {
		return nil, fmt.Errorf("failed to write schema: %w", err)
	}
	schemaFile.Close()

	// codex exec uses --output-schema for structured output
	args := []string{
		"exec",
		"--output-schema", schemaFile.Name(),
		"--json", // Output as JSONL for parsing
		"-",      // Read prompt from stdin
	}

	cmd := exec.CommandContext(ctx, r.Command, args...)

	// Set working directory so Codex can access files relative to --root
	if r.WorkDir != "" {
		cmd.Dir = r.WorkDir
	}

	cmd.Stdin = strings.NewReader(prompt)

	// Always capture output; in verbose mode also stream to stderr
	var stdout, stderr bytes.Buffer
	if r.Verbose {
		cmd.Stdout = &stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("codex command failed: %w\nStderr: %s", err, stderr.String())
	}

	logger.Debug("codex responded", "output_length", len(stdout.Bytes()))

	// Parse JSONL output - look for item.completed with type agent_message
	var resultJSON string
	lines := bytes.Split(stdout.Bytes(), []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var event struct {
			Type string `json:"type"`
			Item struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"item"`
		}
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}
		if event.Type == "item.completed" && event.Item.Type == "agent_message" {
			resultJSON = event.Item.Text
			break
		}
	}

	if resultJSON == "" {
		return nil, fmt.Errorf("codex did not return structured output")
	}

	// Parse claim result from the item text
	var result ClaimResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return nil, fmt.Errorf("invalid claim result structure: %w", err)
	}

	logger.Debug("parsed result", "claim_id", result.ClaimID, "result", result.Result, "bullets", len(result.Bullets))

	// Log bullet paths for debugging
	bulletPaths := make([]string, len(result.Bullets))
	for i, b := range result.Bullets {
		bulletPaths[i] = b.Path
	}
	logger.Debug("bullet paths from codex", "paths", bulletPaths)

	return &result, nil
}

// RunStep executes codex for a single step check in incremental mode
func (r *CodexRunner) RunStep(ctx context.Context, prompt string) (*StepResult, error) {
	// For now, return a simple not-implemented error
	// Codex incremental mode can be implemented later if needed
	return nil, fmt.Errorf("incremental mode not yet implemented for Codex")
}

// RunProof executes codex for the new @proof syntax
func (r *CodexRunner) RunProof(ctx context.Context, prompt string) (*ProofResult, error) {
	// For now, return a simple not-implemented error
	// Codex proof mode can be implemented later if needed
	return nil, fmt.Errorf("proof mode not yet implemented for Codex")
}
