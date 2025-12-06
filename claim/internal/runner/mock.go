package runner

import (
	"context"
)

// MockRunner is a mock implementation for testing
type MockRunner struct {
	Response *ClaimResult
	Err      error
	Calls    []string // Track prompts for assertions
}

// Run returns the pre-configured response
func (m *MockRunner) Run(ctx context.Context, prompt string) (*ClaimResult, error) {
	m.Calls = append(m.Calls, prompt)
	return m.Response, m.Err
}
