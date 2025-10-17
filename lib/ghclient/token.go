package ghclient

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// GetToken retrieves a GitHub token from environment variables or gh CLI.
func GetToken() string {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		slog.Debug("GitHub token resolved", "source", "GITHUB_TOKEN")
		return token
	}
	if token := os.Getenv("MISE_GITHUB_TOKEN"); token != "" {
		slog.Debug("GitHub token resolved", "source", "MISE_GITHUB_TOKEN")
		return token
	}

	output, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		slog.Debug("Failed to retrieve GitHub token via gh CLI", "source", "gh_cli", "error", err)
		slog.Debug("GitHub token unavailable after checking all sources")
		return ""
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		slog.Debug("gh CLI returned empty GitHub token output", "source", "gh_cli")
		slog.Debug("GitHub token unavailable after checking all sources")
		return ""
	}

	slog.Debug("GitHub token resolved", "source", "gh_cli")
	return token
}

// NewHTTPClient returns an HTTP client configured with authentication headers.
func NewHTTPClient(ctx context.Context) *http.Client {
	if ctx == nil {
		ctx = context.Background()
	}

	token := GetToken()
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if token == "" {
		return client
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	authenticatedClient := oauth2.NewClient(ctx, ts)
	authenticatedClient.Timeout = client.Timeout
	return authenticatedClient
}
