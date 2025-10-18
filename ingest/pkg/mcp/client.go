package mcp

import (
	"context"
	"net/http"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Option modifies client construction.
type Option func(*clientOptions)

type clientOptions struct {
	impl        *sdkmcp.Implementation
	healthCheck HealthCheck
}

// WithImplementation overrides the implementation metadata sent to the server.
func WithImplementation(impl *sdkmcp.Implementation) Option {
	return func(o *clientOptions) {
		o.impl = impl
	}
}

// WithHealthCheck registers a callback invoked after each successful handshake.
// Returning an error causes the connection attempt to fail (and be retried).
func WithHealthCheck(fn HealthCheck) Option {
	return func(o *clientOptions) {
		o.healthCheck = fn
	}
}

// Client wraps an MCP client with ingest-specific configuration.
type Client struct {
	cfg         Config
	sdkClient   *sdkmcp.Client
	httpClient  *http.Client
	retry       RetryConfig
	healthCheck HealthCheck
}

// NewClient creates a client capable of connecting to an MCP server via SSE.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	cfg = cfg.withDefaults()

	var options clientOptions
	for _, opt := range opts {
		opt(&options)
	}
	impl := options.impl
	if impl == nil {
		impl = &sdkmcp.Implementation{Name: "ingest-mcp-client", Version: "dev"}
	}

	httpClient := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: newHeaderRoundTripper(http.DefaultTransport, cfg),
	}

	return &Client{
		cfg:         cfg,
		sdkClient:   sdkmcp.NewClient(impl, nil),
		httpClient:  httpClient,
		retry:       cfg.Retry,
		healthCheck: options.healthCheck,
	}, nil
}

// Connect establishes a session to the configured server.
func (c *Client) Connect(ctx context.Context) (*Session, error) {
	backoff := c.retry.InitialBackoff
	var lastErr error

	for attempt := 1; attempt <= c.retry.MaxAttempts; attempt++ {
		session, err := c.connectOnce(ctx)
		if err == nil {
			return session, nil
		}
		lastErr = err

		// If we've exhausted attempts or the context is done, stop.
		if attempt == c.retry.MaxAttempts || ctx.Err() != nil {
			break
		}

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}

		backoff *= 2
		if backoff > c.retry.MaxBackoff {
			backoff = c.retry.MaxBackoff
		}
	}

	return nil, lastErr
}

func (c *Client) connectOnce(ctx context.Context) (*Session, error) {
	transport := &sdkmcp.SSEClientTransport{
		Endpoint:   c.cfg.Endpoint,
		HTTPClient: c.httpClient,
	}

	cs, err := c.sdkClient.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}

	if c.healthCheck != nil {
		if err := c.healthCheck(ctx, cs); err != nil {
			cs.Close()
			return nil, err
		}
	}

	return &Session{session: cs}, nil
}

// Session represents an active MCP session.
type Session struct {
	session *sdkmcp.ClientSession
}

// Close terminates the session.
func (s *Session) Close() error {
	return s.session.Close()
}

// CallTool invokes a tool exposed by the MCP server.
func (s *Session) CallTool(ctx context.Context, name string, args map[string]any) (*sdkmcp.CallToolResult, error) {
	params := &sdkmcp.CallToolParams{
		Name:      name,
		Arguments: args,
	}
	return s.session.CallTool(ctx, params)
}

// ListTools returns metadata about tools exposed by the server.
func (s *Session) ListTools(ctx context.Context) ([]*sdkmcp.Tool, error) {
	resp, err := s.session.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}
	return resp.Tools, nil
}

// Internal client session access for advanced use cases.
func (s *Session) raw() *sdkmcp.ClientSession {
	return s.session
}

func newHeaderRoundTripper(base http.RoundTripper, cfg Config) http.RoundTripper {
	headers := make(http.Header)
	for k, v := range cfg.Headers {
		headers.Set(k, v)
	}
	if cfg.AuthToken != "" {
		headers.Set("Authorization", "Bearer "+cfg.AuthToken)
	}

	return &headerRoundTripper{
		base:    base,
		headers: headers,
	}
}

type headerRoundTripper struct {
	base    http.RoundTripper
	headers http.Header
}

func (h *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, values := range h.headers {
		for _, v := range values {
			req.Header.Set(k, v)
		}
	}
	return h.base.RoundTrip(req)
}

// HealthCheck is invoked after a session is established. Implementations should
// perform lightweight verification, e.g. ListTools or CallTool with a noop.
type HealthCheck func(context.Context, *sdkmcp.ClientSession) error
