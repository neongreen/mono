package readability

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// RemoteServiceEngine is a generic engine for remote readability services
type RemoteServiceEngine struct {
	name       string
	serviceURL string
	client     *http.Client
}

// NewRemoteServiceEngine creates a new remote service engine
func NewRemoteServiceEngine(name, serviceURL string) *RemoteServiceEngine {
	return &RemoteServiceEngine{
		name:       name,
		serviceURL: serviceURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewPureMDEngine creates a new pure.md engine
func NewPureMDEngine() *RemoteServiceEngine {
	return NewRemoteServiceEngine("pure-md", "https://pure.md/")
}

// NewJinaEngine creates a new Jina Reader engine
func NewJinaEngine() *RemoteServiceEngine {
	return NewRemoteServiceEngine("jina", "https://r.jina.ai/")
}

// Name returns the engine name
func (e *RemoteServiceEngine) Name() string {
	return e.name
}

// Extract extracts readable content using the remote service
func (e *RemoteServiceEngine) Extract(html []byte, sourceURL string) ([]byte, error) {
	if sourceURL == "" {
		return nil, fmt.Errorf("%s engine requires a source URL", e.name)
	}

	// Construct the service URL by prepending the service base URL to the source URL
	fullURL := e.serviceURL + sourceURL

	resp, err := e.client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from %s: %w", e.name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned status %d", e.name, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response: %w", e.name, err)
	}

	return content, nil
}

// IsAvailable checks if the remote service is accessible
func (e *RemoteServiceEngine) IsAvailable() error {
	// We assume it's always available since it's a web service
	// Could add a connectivity check here if needed
	return nil
}
