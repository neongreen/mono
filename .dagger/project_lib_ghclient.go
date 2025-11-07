package main

import (
	"context"

	"dagger/internal/dagger"
)

type LibGhclientProject struct {
	Dagger *Dagger // +private
}

func (p *LibGhclientProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "lib/ghclient", format)
}

func (p *LibGhclientProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "lib/ghclient", format)
}
