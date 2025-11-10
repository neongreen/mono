package main

import (
	"context"

	"dagger/internal/dagger"
)

type LintersUselesswrapperProject struct {
	Dagger *Dagger // +private
}

func (p *LintersUselesswrapperProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("linters/uselesswrapper")
	return buildProject(ctx, "linters/uselesswrapper", "./cmd/uselesswrapper", src)
}

func (p *LintersUselesswrapperProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("linters/uselesswrapper")
	return testProject(ctx, "linters/uselesswrapper", format, src)
}

func (p *LintersUselesswrapperProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("linters/uselesswrapper")
	return coverageFile(ctx, "linters/uselesswrapper", format, src)
}
