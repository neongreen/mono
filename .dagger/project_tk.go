package main

import (
	"context"

	"dagger/internal/dagger"
)

type TkProject struct {
	Dagger *Dagger // +private
}

func (p *TkProject) Build(ctx context.Context) (*dagger.File, error) {
	return buildProject(ctx, "tk", ".")
}

func (p *TkProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "tk", format)
}

func (p *TkProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "tk", format)
}
