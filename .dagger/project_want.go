package main

import (
	"context"

	"dagger/internal/dagger"
)

type WantProject struct {
	Dagger *Dagger // +private
}

func (p *WantProject) Build(ctx context.Context) (*dagger.File, error) {
	return buildProject(ctx, "want", "./cmd")
}

func (p *WantProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "want", format)
}

func (p *WantProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "want", format)
}

func (p *WantProject) Lint(ctx context.Context) (string, error) {
	return lintProject(ctx, "want")
}

