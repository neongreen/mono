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
	src := getFilteredSource("lib/ghclient")
	return testProject(ctx, "lib/ghclient", format, src)
}

func (p *LibGhclientProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("lib/ghclient")
	return coverageFile(ctx, "lib/ghclient", format, src)
}
