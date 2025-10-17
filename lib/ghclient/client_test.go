package ghclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewClientAddsAuthorizationHeader(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "secret-token")
	ctx := context.Background()

	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]string{"login": "octocat"})
	}))
	t.Cleanup(server.Close)

	client := NewClient(ctx)
	client.BaseURL = mustParseURL(server.URL + "/")
	client.UploadURL = client.BaseURL

	_, _, err := client.Users.Get(ctx, "octocat")
	if err != nil {
		// go-github treats HTTP 200 with empty body as success; any other error is unexpected.
		t.Fatalf("Repositories.List returned error: %v", err)
	}

	if capturedAuth != "Bearer secret-token" {
		t.Fatalf("authorization header = %q, want %q", capturedAuth, "Bearer secret-token")
	}
}

func TestNewEnterpriseClientUsesCustomURLs(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "enterprise-token")
	ctx := context.Background()

	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]string{"login": "enterprise-user"})
	}))
	t.Cleanup(server.Close)

	client, err := NewEnterpriseClient(ctx, server.URL, server.URL)
	if err != nil {
		t.Fatalf("NewEnterpriseClient error: %v", err)
	}
	client.BaseURL = mustParseURL(server.URL + "/")
	client.UploadURL = client.BaseURL

	_, _, err = client.Users.Get(ctx, "enterprise-user")
	if err != nil {
		t.Fatalf("Repositories.List returned error: %v", err)
	}

	if receivedPath != "/users/enterprise-user" {
		t.Fatalf("expected enterprise path '/users/enterprise-user', got %q", receivedPath)
	}
}

func mustParseURL(raw string) *url.URL {
	parsed, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return parsed
}
