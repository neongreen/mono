package ghclient

import (
	"context"

	"github.com/google/go-github/v61/github"
)

// NewClient constructs a go-github client configured with our shared HTTP settings.
func NewClient(ctx context.Context) *github.Client {
	if ctx == nil {
		ctx = context.Background()
	}
	httpClient := NewHTTPClient(ctx)
	return github.NewClient(httpClient)
}

// NewEnterpriseClient builds a client for GitHub Enterprise deployments using custom URLs.
func NewEnterpriseClient(ctx context.Context, baseURL, uploadURL string) (*github.Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	httpClient := NewHTTPClient(ctx)
	return github.NewEnterpriseClient(baseURL, uploadURL, httpClient)
}
