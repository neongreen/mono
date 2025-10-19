package mcp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestClientConnectAndCallTool(t *testing.T) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test-server", Version: "v0.0.1"}, nil)
	sdkmcp.AddTool(server, &sdkmcp.Tool{Name: "echo", Description: "echo back text"}, func(ctx context.Context, req *sdkmcp.CallToolRequest, args map[string]any) (*sdkmcp.CallToolResult, any, error) {
		text, _ := args["text"].(string)
		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{Text: text},
			},
		}, nil, nil
	})

	var capturedAuth string
	var capturedCustom string
	handler := sdkmcp.NewSSEHandler(func(r *http.Request) *sdkmcp.Server {
		capturedAuth = r.Header.Get("Authorization")
		capturedCustom = r.Header.Get("X-Trace")
		return server
	}, nil)

	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	cfg := Config{
		Endpoint:  httpServer.URL,
		AuthToken: "linear-secret",
		Headers: map[string]string{
			"X-Trace": "ingest-test",
		},
		Timeout: 5 * time.Second,
	}

	client, err := NewClient(cfg, WithHealthCheck(func(ctx context.Context, session *sdkmcp.ClientSession) error {
		// Basic health: ensure the server exposes the echo tool.
		tools, err := session.ListTools(ctx, nil)
		if err != nil {
			return err
		}
		if len(tools.Tools) == 0 {
			return errors.New("no tools returned")
		}
		return nil
	}))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer session.Close()

	result, err := session.CallTool(ctx, "echo", map[string]any{"text": "hello"})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	text := result.Content[0].(*sdkmcp.TextContent).Text
	if text != "hello" {
		t.Fatalf("expected echo result 'hello', got %q", text)
	}

	if want := "Bearer linear-secret"; capturedAuth != want {
		t.Fatalf("expected Authorization header %q, got %q", want, capturedAuth)
	}
	if capturedCustom != "ingest-test" {
		t.Fatalf("expected custom header to propagate, got %q", capturedCustom)
	}
}

func TestClientRequiresEndpoint(t *testing.T) {
	_, err := NewClient(Config{})
	if err == nil {
		t.Fatal("expected validation error for missing endpoint")
	}
}

func TestClientConnectRetries(t *testing.T) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "retry-server", Version: "v0.0.1"}, nil)
	sdkmcp.AddTool(server, &sdkmcp.Tool{Name: "noop", Description: "returns nothing"}, func(ctx context.Context, req *sdkmcp.CallToolRequest, args map[string]any) (*sdkmcp.CallToolResult, any, error) {
		return &sdkmcp.CallToolResult{}, nil, nil
	})

	sseHandler := sdkmcp.NewSSEHandler(func(*http.Request) *sdkmcp.Server { return server }, nil)

	handshakeAttempts := 0
	failingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") == "text/event-stream" {
			handshakeAttempts++
			if handshakeAttempts == 1 {
				http.Error(w, "temporary failure", http.StatusBadGateway)
				return
			}
		}
		sseHandler.ServeHTTP(w, r)
	})

	srv := httptest.NewServer(failingHandler)
	defer srv.Close()

	cfg := Config{
		Endpoint: srv.URL,
		Retry: RetryConfig{
			MaxAttempts:    3,
			InitialBackoff: 5 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	session, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer session.Close()

	if handshakeAttempts != 2 {
		t.Fatalf("expected 2 handshake attempts, got %d", handshakeAttempts)
	}
}

func TestClientHealthCheckRetry(t *testing.T) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "health-server", Version: "v0.0.1"}, nil)
	sdkmcp.AddTool(server, &sdkmcp.Tool{Name: "noop", Description: "noop"}, func(ctx context.Context, req *sdkmcp.CallToolRequest, args map[string]any) (*sdkmcp.CallToolResult, any, error) {
		return &sdkmcp.CallToolResult{}, nil, nil
	})

	handler := sdkmcp.NewSSEHandler(func(*http.Request) *sdkmcp.Server { return server }, nil)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	healthAttempts := 0
	healthCheck := func(ctx context.Context, session *sdkmcp.ClientSession) error {
		healthAttempts++
		if healthAttempts == 1 {
			return errors.New("health check failed")
		}
		_, err := session.ListTools(ctx, nil)
		return err
	}

	cfg := Config{
		Endpoint: srv.URL,
		Retry: RetryConfig{
			MaxAttempts:    2,
			InitialBackoff: 5 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
		},
	}

	client, err := NewClient(cfg, WithHealthCheck(healthCheck))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	session, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer session.Close()

	if healthAttempts != 2 {
		t.Fatalf("expected health check to run twice, got %d", healthAttempts)
	}
}
