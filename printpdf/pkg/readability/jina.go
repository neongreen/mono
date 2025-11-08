package readability

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// JinaEngine uses the Jina Reader service for content extraction
type JinaEngine struct {
	client *http.Client
}

// NewJinaEngine creates a new Jina Reader engine
func NewJinaEngine() *JinaEngine {
	return &JinaEngine{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the engine name
func (e *JinaEngine) Name() string {
	return "jina"
}

// Extract extracts readable content using Jina Reader service
func (e *JinaEngine) Extract(html []byte, sourceURL string) ([]byte, error) {
	if sourceURL == "" {
		return nil, fmt.Errorf("jina engine requires a source URL")
	}

	// Jina Reader works by prepending https://r.jina.ai/ to the URL
	jinaURL := "https://r.jina.ai/" + sourceURL

	resp, err := e.client.Get(jinaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from jina: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jina returned status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read jina response: %w", err)
	}

	return content, nil
}

// IsAvailable checks if Jina Reader is accessible
func (e *JinaEngine) IsAvailable() error {
	// We assume it's always available since it's a web service
	// Could add a connectivity check here if needed
	return nil
}
