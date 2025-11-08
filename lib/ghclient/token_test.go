package ghclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/neongreen/mono/lib/testhelpers"
)

// Test helper functions moved to lib/testhelpers

func TestGetToken(t *testing.T) {
	t.Run("GITHUB_TOKEN takes precedence", func(t *testing.T) {
		buf := testhelpers.SetupTestLogger(t)
		t.Setenv("GITHUB_TOKEN", "github_token")
		t.Setenv("MISE_GITHUB_TOKEN", "mise_token")

		token := GetToken()
		if token != "github_token" {
			t.Fatalf("GetToken() = %q, want %q", token, "github_token")
		}

		output := buf.String()
		if !strings.Contains(output, "source=GITHUB_TOKEN") {
			t.Fatalf("expected log to mention GITHUB_TOKEN source, got %q", output)
		}
		if strings.Contains(output, "github_token") {
			t.Fatalf("logs must not contain actual token, got %q", output)
		}
	})

	t.Run("MISE_GITHUB_TOKEN used when GITHUB_TOKEN not set", func(t *testing.T) {
		buf := testhelpers.SetupTestLogger(t)
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("MISE_GITHUB_TOKEN", "mise_token")

		token := GetToken()
		if token != "mise_token" {
			t.Fatalf("GetToken() = %q, want %q", token, "mise_token")
		}

		output := buf.String()
		if !strings.Contains(output, "source=MISE_GITHUB_TOKEN") {
			t.Fatalf("expected log to mention MISE_GITHUB_TOKEN source, got %q", output)
		}
		if strings.Contains(output, "mise_token") {
			t.Fatalf("logs must not contain actual token, got %q", output)
		}
	})

	t.Run("returns empty string when no tokens available", func(t *testing.T) {
		buf := testhelpers.SetupTestLogger(t)
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("MISE_GITHUB_TOKEN", "")
		testhelpers.InstallGhStub(t, "", 1)

		token := GetToken()
		if token != "" {
			t.Fatalf("GetToken() = %q, want empty string", token)
		}

		output := buf.String()
		if !strings.Contains(output, "GitHub token unavailable after checking all sources") {
			t.Fatalf("expected log to note unavailable token, got %q", output)
		}
		if !strings.Contains(output, "source=gh_cli") {
			t.Fatalf("expected log to mention gh_cli source, got %q", output)
		}
	})

	t.Run("gh CLI token used when available", func(t *testing.T) {
		buf := testhelpers.SetupTestLogger(t)
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("MISE_GITHUB_TOKEN", "")
		testhelpers.InstallGhStub(t, "cli_token", 0)

		token := GetToken()
		if token != "cli_token" {
			t.Fatalf("GetToken() = %q, want %q", token, "cli_token")
		}

		output := buf.String()
		if !strings.Contains(output, "source=gh_cli") {
			t.Fatalf("expected log to mention gh_cli source, got %q", output)
		}
		if strings.Contains(output, "cli_token") {
			t.Fatalf("logs must not contain actual token, got %q", output)
		}
	})
}

func TestNewHTTPClient(t *testing.T) {
	t.Run("no token returns default client", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("MISE_GITHUB_TOKEN", "")
		testhelpers.InstallGhStub(t, "", 1)

		client := NewHTTPClient(context.Background())
		if client.Timeout != 30*time.Second {
			t.Fatalf("expected timeout 30s, got %v", client.Timeout)
		}
	})

	t.Run("token adds authorization header", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "secret-token")

		var capturedAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client := NewHTTPClient(context.Background())

		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("client.Do() error = %v", err)
		}
		resp.Body.Close()

		if capturedAuth != "Bearer secret-token" {
			t.Fatalf("authorization header = %q, want %q", capturedAuth, "Bearer secret-token")
		}
	})
}
