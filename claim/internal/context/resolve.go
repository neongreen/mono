package context

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/neongreen/mono/claim/internal/logger"
)

// ResolvedContext contains the fetched context content
type ResolvedContext struct {
	Specification string // Original @context specification
	Content       string // Resolved content from Claude
}

// Resolver resolves @context specifications to actual content
type Resolver struct {
	Command string // Path to claude command
	Model   string // Model to use (e.g., "haiku" for faster/cheaper resolution)
	WorkDir string // Working directory for resolution
}

// Resolve fetches the context specified by the @context block
// It uses Claude to interpret natural language context specifications
func (r *Resolver) Resolve(ctx context.Context, specification string) (*ResolvedContext, error) {
	if strings.TrimSpace(specification) == "" {
		return nil, nil
	}

	logger.Debug("resolving context", "spec_length", len(specification))

	prompt := buildContextPrompt(specification)

	args := []string{
		"--print",
		"--output-format", "json",
		"--tools", "Read,Grep,Glob",
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
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("context resolution failed: %w\nStderr: %s", err, stderr.String())
	}

	// Parse JSON response - claude --output-format json returns {result: string, ...}
	var response struct {
		Result string `json:"result"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse context response: %w", err)
	}

	resolved := &ResolvedContext{
		Specification: specification,
		Content:       response.Result,
	}

	logger.Debug("resolved context", "content_length", len(resolved.Content))
	return resolved, nil
}

func buildContextPrompt(specification string) string {
	return fmt.Sprintf(`You are a context fetcher. Your job is to retrieve and OUTPUT the actual code/content specified below.

# Context Specification

%s

# Instructions

1. Read the specification carefully
2. Use the available tools (Read, Grep, Glob) to fetch the requested content
3. OUTPUT THE ACTUAL CODE/CONTENT - not just line numbers or file references
4. Before each piece of content, include a one-line header, e.g.:
   - "Results of grep 'pattern' in file.go:"
   - "Function Foo in path/to/file.go:"
5. Then output the FULL content in a code block with appropriate language tag:
   - For functions: the entire function body including signature
   - For grep: all matching lines with context
   - For files: the requested portion of the file
6. Do NOT just say "Lines 10-20" - actually include those lines
7. Use markdown code fences with language tags for code (e.g. '''go, '''typescript)
8. No explanation beyond the headers - just the actual content

Return the ACTUAL CONTENT now:`, specification)
}
