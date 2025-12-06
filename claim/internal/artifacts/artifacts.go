package artifacts

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Store manages artifact storage for a claim check session
type Store struct {
	Dir string // Base directory for all artifacts
}

// NewStore creates a new artifact store in a temp directory
func NewStore() (*Store, error) {
	// Create a timestamped directory under /tmp/claim-artifacts
	timestamp := time.Now().Format("20060102-150405")
	dir := filepath.Join("/tmp", "claim-artifacts", timestamp)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create artifacts dir: %w", err)
	}

	return &Store{Dir: dir}, nil
}

// SavePrompt saves a prompt for a claim check
func (s *Store) SavePrompt(claimID, prompt string) (string, error) {
	filename := fmt.Sprintf("%s-prompt.md", claimID)
	path := filepath.Join(s.Dir, filename)

	if err := os.WriteFile(path, []byte(prompt), 0644); err != nil {
		return "", fmt.Errorf("failed to save prompt: %w", err)
	}

	return path, nil
}

// SaveClaudeOutput saves the raw Claude output for a claim check
func (s *Store) SaveClaudeOutput(claimID string, output []byte) (string, error) {
	filename := fmt.Sprintf("%s-claude-output.jsonl", claimID)
	path := filepath.Join(s.Dir, filename)

	if err := os.WriteFile(path, output, 0644); err != nil {
		return "", fmt.Errorf("failed to save claude output: %w", err)
	}

	return path, nil
}

// SaveContextResolution saves the context resolution result
func (s *Store) SaveContextResolution(claimID, context string) (string, error) {
	filename := fmt.Sprintf("%s-context.md", claimID)
	path := filepath.Join(s.Dir, filename)

	if err := os.WriteFile(path, []byte(context), 0644); err != nil {
		return "", fmt.Errorf("failed to save context: %w", err)
	}

	return path, nil
}
