package mcp

import (
	"errors"
	"time"
)

// Config controls how the MCP client connects to a remote server.
type Config struct {
	// Endpoint is the base URL for the server's SSE transport.
	Endpoint string
	// AuthToken, when set, is sent as a bearer token on every HTTP request.
	AuthToken string
	// Headers are additional HTTP headers to attach to requests.
	Headers map[string]string
	// Timeout bounds network operations. Defaults to 30 seconds.
	Timeout time.Duration
	// Retry configures connection retry behaviour.
	Retry RetryConfig
}

func (c Config) validate() error {
	if c.Endpoint == "" {
		return errors.New("mcp: endpoint is required")
	}
	return nil
}

func (c Config) withDefaults() Config {
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	c.Retry = c.Retry.withDefaults()
	return c
}

// RetryConfig controls connection retries.
type RetryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

func (r RetryConfig) withDefaults() RetryConfig {
	if r.MaxAttempts <= 0 {
		r.MaxAttempts = 3
	}
	if r.InitialBackoff <= 0 {
		r.InitialBackoff = 500 * time.Millisecond
	}
	if r.MaxBackoff <= 0 {
		r.MaxBackoff = 5 * time.Second
	}
	return r
}
