package mcp

import (
	"context"
	"errors"
	"math/rand"
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
	if backoff <= 0 {
		backoff = time.Second
	}
	maxBackoff := c.retry.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 5 * time.Second
	}
	maxAttempts := c.retry.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		session, err := c.connectOnce(ctx)
		if err == nil {
			session.SessionRetry = c.retry
			return session, nil
		}
		lastErr = err

		if !shouldRetry(err) {
			break
		}

		if attempt == maxAttempts || ctx.Err() != nil {
			break
		}

		sleep := jitter(backoff)
		if err := sleepContext(ctx, sleep); err != nil {
			return nil, err
		}

		backoff = nextBackoff(backoff, maxBackoff)
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
	SessionRetry RetryConfig
	session      *sdkmcp.ClientSession
}

// Close terminates the session.
func (s *Session) Close() error {
	return s.session.Close()
}

// CallTool invokes a tool exposed by the MCP server.
func (s *Session) CallTool(ctx context.Context, name string, args map[string]any) (*sdkmcp.CallToolResult, error) {
	var res *sdkmcp.CallToolResult
	op := func() error {
		params := &sdkmcp.CallToolParams{
			Name:      name,
			Arguments: args,
		}
		resp, err := s.session.CallTool(ctx, params)
		if err != nil {
			return err
		}
		res = resp
		return nil
	}
	if err := s.retryOperation(ctx, op); err != nil {
		return nil, err
	}
	return res, nil
}

// ListTools returns metadata about tools exposed by the server.
func (s *Session) ListTools(ctx context.Context) ([]*sdkmcp.Tool, error) {
	var tools []*sdkmcp.Tool
	op := func() error {
		resp, err := s.session.ListTools(ctx, nil)
		if err != nil {
			return err
		}
		tools = resp.Tools
		return nil
	}
	if err := s.retryOperation(ctx, op); err != nil {
		return nil, err
	}
	return tools, nil
}

// ListResources returns resources exposed by the server.
func (s *Session) ListResources(ctx context.Context, params *sdkmcp.ListResourcesParams) (*sdkmcp.ListResourcesResult, error) {
	var result *sdkmcp.ListResourcesResult
	op := func() error {
		resp, err := s.session.ListResources(ctx, params)
		if err != nil {
			return err
		}
		result = resp
		return nil
	}
	if err := s.retryOperation(ctx, op); err != nil {
		return nil, err
	}
	return result, nil
}

// ReadResource fetches the contents of a specific resource URI.
func (s *Session) ReadResource(ctx context.Context, params *sdkmcp.ReadResourceParams) (*sdkmcp.ReadResourceResult, error) {
	var result *sdkmcp.ReadResourceResult
	op := func() error {
		resp, err := s.session.ReadResource(ctx, params)
		if err != nil {
			return err
		}
		result = resp
		return nil
	}
	if err := s.retryOperation(ctx, op); err != nil {
		return nil, err
	}
	return result, nil
}

// Internal client session access for advanced use cases.
func (s *Session) raw() *sdkmcp.ClientSession {
	return s.session
}

func (s *Session) retryOperation(ctx context.Context, op func() error) error {
	backoff := s.SessionRetry.InitialBackoff
	if backoff <= 0 {
		backoff = time.Second
	}
	maxBackoff := s.SessionRetry.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 5 * time.Second
	}
	maxAttempts := s.SessionRetry.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := op(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if !shouldRetry(lastErr) {
			break
		}
		if attempt == maxAttempts || ctx.Err() != nil {
			break
		}
		sleep := jitter(backoff)
		if err := sleepContext(ctx, sleep); err != nil {
			return err
		}
		backoff = nextBackoff(backoff, maxBackoff)
	}
	return lastErr
}

func jitter(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}
	factor := 0.6 + rand.Float64()*0.8 // [0.6,1.4)
	return time.Duration(float64(base) * factor)
}

func nextBackoff(current, max time.Duration) time.Duration {
	if current <= 0 {
		current = time.Second
	}
	if max <= 0 {
		max = 5 * time.Second
	}
	current *= 2
	if current > max {
		return max
	}
	return current
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if temp, ok := err.(interface{ Temporary() bool }); ok {
		return temp.Temporary()
	}
	return true
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
