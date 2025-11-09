package main

import (
	"context"

	"dagger/internal/dagger"
)

type LintersUselesswrapperProject struct {
	Dagger *Dagger // +private
}

func (p *LintersUselesswrapperProject) Build(ctx context.Context) (*dagger.File, error) {
	return buildProject(ctx, "linters/uselesswrapper", "./cmd/uselesswrapper")
}

func (p *LintersUselesswrapperProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "linters/uselesswrapper", format)
}

func (p *LintersUselesswrapperProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "linters/uselesswrapper", format)
}
