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
	return testProject(ctx, "lib/ghrelease", format)
}

func (p *LibGhreleaseProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "lib/ghrelease", format)
}
