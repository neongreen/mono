package readability

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// PureMDEngine uses the pure.md service for content extraction
type PureMDEngine struct {
	client *http.Client
}

// NewPureMDEngine creates a new pure.md engine
func NewPureMDEngine() *PureMDEngine {
	return &PureMDEngine{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the engine name
func (e *PureMDEngine) Name() string {
	return "pure-md"
}

// Extract extracts readable content using pure.md service
func (e *PureMDEngine) Extract(html []byte, sourceURL string) ([]byte, error) {
	if sourceURL == "" {
		return nil, fmt.Errorf("pure-md engine requires a source URL")
	}

	// pure.md works by prepending https://pure.md/ to the URL
	pureURL := "https://pure.md/" + sourceURL

	resp, err := e.client.Get(pureURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from pure.md: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pure.md returned status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read pure.md response: %w", err)
	}

	return content, nil
}

// IsAvailable checks if pure.md is accessible
func (e *PureMDEngine) IsAvailable() error {
	// We assume it's always available since it's a web service
	// Could add a connectivity check here if needed
	return nil
}
