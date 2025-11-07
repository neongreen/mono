package main

import (
	"context"

	"dagger/internal/dagger"
)

type ConfProject struct {
	Dagger *Dagger // +private
}

func (p *ConfProject) Build(ctx context.Context) (*dagger.File, error) {
	return buildProject(ctx, "conf", "./cmd")
}

func (p *ConfProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "conf", format)
}

func (p *ConfProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "conf", format)
}

func (p *ConfProject) Lint(ctx context.Context) (string, error) {
	return lintProject(ctx, "conf")
}

