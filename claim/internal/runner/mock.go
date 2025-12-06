package runner

import (
	"context"
)

// MockRunner is a mock implementation for testing
type MockRunner struct {
	Response      *ClaimResult
	StepResponse  *StepResult
	ProofResponse *ProofResult
	Err           error
	Calls         []string // Track prompts for assertions
}

// Run returns the pre-configured response
func (m *MockRunner) Run(ctx context.Context, prompt string) (*ClaimResult, error) {
	m.Calls = append(m.Calls, prompt)
	return m.Response, m.Err
}

// RunStep returns the pre-configured step response
func (m *MockRunner) RunStep(ctx context.Context, prompt string) (*StepResult, error) {
	m.Calls = append(m.Calls, prompt)
	return m.StepResponse, m.Err
}

// RunProof returns the pre-configured proof response
func (m *MockRunner) RunProof(ctx context.Context, prompt string) (*ProofResult, error) {
	m.Calls = append(m.Calls, prompt)
	return m.ProofResponse, m.Err
}
