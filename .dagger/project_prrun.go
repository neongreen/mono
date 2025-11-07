package main

import (
	"context"

	"dagger/internal/dagger"
)

type PrrunProject struct {
	Dagger *Dagger // +private
}

func (p *PrrunProject) Build(ctx context.Context) (*dagger.File, error) {
	return buildProject(ctx, "prrun", ".")
}

func (p *PrrunProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "prrun", format)
}

func (p *PrrunProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "prrun", format)
}

func (p *PrrunProject) Lint(ctx context.Context) (string, error) {
	return lintProject(ctx, "prrun")
}

