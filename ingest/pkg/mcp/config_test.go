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

func TestResolveConfigUsesProviderDefaults(t *testing.T) {
	t.Setenv("INGEST_MCP_ENDPOINT", "")
	t.Setenv("INGEST_LINEAR_MCP_ENDPOINT", "")
	cfg, err := ResolveConfig("linear", Config{})
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Endpoint != "https://mcp.linear.app/sse" {
		t.Fatalf("expected linear default endpoint, got %q", cfg.Endpoint)
	}

	t.Setenv("INGEST_GITHUB_MCP_ENDPOINT", "")
	cfg, err = ResolveConfig("github", Config{})
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.Endpoint != "https://api.githubcopilot.com/mcp/" {
		t.Fatalf("expected github default endpoint, got %q", cfg.Endpoint)
	}
}

func TestResolveConfigUsesGitHubTokenFallbacks(t *testing.T) {
	t.Setenv("INGEST_GITHUB_MCP_ENDPOINT", "https://api.githubcopilot.com/mcp/")
	t.Setenv("INGEST_GITHUB_MCP_TOKEN", "")
	t.Setenv("INGEST_MCP_TOKEN", "")
	t.Setenv("MISE_GITHUB_TOKEN", "mise-token")
	t.Setenv("GITHUB_TOKEN", "gh-token")

	cfg, err := ResolveConfig("github", Config{})
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.AuthToken != "mise-token" {
		t.Fatalf("expected token from MISE_GITHUB_TOKEN, got %q", cfg.AuthToken)
	}

	t.Setenv("MISE_GITHUB_TOKEN", "")
	cfg, err = ResolveConfig("github", Config{})
	if err != nil {
		t.Fatalf("ResolveConfig: %v", err)
	}
	if cfg.AuthToken != "gh-token" {
		t.Fatalf("expected token from GITHUB_TOKEN, got %q", cfg.AuthToken)
	}
}

func TestResolveConfigUnknownProviderRequiresEndpoint(t *testing.T) {
	t.Setenv("INGEST_UNKNOWN_MCP_ENDPOINT", "")
	if _, err := ResolveConfig("unknown", Config{}); err == nil {
		t.Fatal("expected error for unknown provider without endpoint")
	}

	cfg, err := ResolveConfig("unknown", Config{Endpoint: "https://example.com"})
	if err != nil {
		t.Fatalf("unexpected error when endpoint provided: %v", err)
	}
	if cfg.Endpoint != "https://example.com" {
		t.Fatalf("expected override endpoint, got %q", cfg.Endpoint)
	}
}
