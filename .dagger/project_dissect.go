package main

import (
	"context"

	"dagger/internal/dagger"
)

type DissectProject struct {
	Dagger *Dagger // +private
}

func (p *DissectProject) Build(ctx context.Context) (*dagger.File, error) {
	return buildProject(ctx, "dissect", "./cmd")
}

func (p *DissectProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "dissect", format)
}

func (p *DissectProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "dissect", format)
}
