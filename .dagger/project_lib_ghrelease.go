package main

import (
	"context"

	"dagger/internal/dagger"
)

type LibGhreleaseProject struct {
	Dagger *Dagger // +private
}

func (p *LibGhreleaseProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("lib/ghrelease")
	return testProject(ctx, "lib/ghrelease", format, src)
}

func (p *LibGhreleaseProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("lib/ghrelease")
	return coverageFile(ctx, "lib/ghrelease", format, src)
}
