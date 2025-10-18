package mcp

import (
	"os"
	"testing"
	"time"
)

func TestResolveConfigMergesEnv(t *testing.T) {
	t.Setenv("INGEST_MCP_ENDPOINT", "https://fallback.example/sse")
	t.Setenv("INGEST_LINEAR_MCP_ENDPOINT", "https://linear.example/sse")
	t.Setenv("INGEST_LINEAR_MCP_TOKEN", "linear-token")
	t.Setenv("INGEST_LINEAR_MCP_HEADERS", "X-Env=linear,X-Trace=trace-id")
	t.Setenv("INGEST_MCP_TIMEOUT", "5s")
	t.Setenv("INGEST_MCP_RETRY_MAX_ATTEMPTS", "4")
	t.Setenv("INGEST_MCP_RETRY_INITIAL_BACKOFF", "10ms")
	t.Setenv("INGEST_MCP_RETRY_MAX_BACKOFF", "40ms")

	cfg, err := ResolveConfig("linear", Config{})
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}

	if cfg.Endpoint != "https://linear.example/sse" {
		t.Fatalf("expected provider endpoint, got %q", cfg.Endpoint)
	}
	if cfg.AuthToken != "linear-token" {
		t.Fatalf("expected token from env, got %q", cfg.AuthToken)
	}
	if cfg.Timeout != 5*time.Second {
		t.Fatalf("expected timeout override, got %v", cfg.Timeout)
	}
	if cfg.Retry.MaxAttempts != 4 || cfg.Retry.InitialBackoff != 10*time.Millisecond || cfg.Retry.MaxBackoff != 40*time.Millisecond {
		t.Fatalf("unexpected retry config %#v", cfg.Retry)
	}
	if cfg.Headers["X-Env"] != "linear" || cfg.Headers["X-Trace"] != "trace-id" {
		t.Fatalf("expected headers from env, got %#v", cfg.Headers)
	}
}

func TestResolveConfigOverridesTakePriority(t *testing.T) {
	t.Setenv("INGEST_MCP_ENDPOINT", "https://env.example/sse")
	cfg, err := ResolveConfig("", Config{
		Endpoint: "https://override.example/sse",
		Headers:  map[string]string{"X-Custom": "value"},
	})
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Endpoint != "https://override.example/sse" {
		t.Fatalf("expected override endpoint, got %q", cfg.Endpoint)
	}
	if cfg.Headers["X-Custom"] != "value" {
		t.Fatalf("override header missing, got %#v", cfg.Headers)
	}
}

func TestResolveConfigMissingEndpointFails(t *testing.T) {
	os.Unsetenv("INGEST_MCP_ENDPOINT")
	if _, err := ResolveConfig("", Config{}); err == nil {
		t.Fatal("expected validation error")
	}
}
