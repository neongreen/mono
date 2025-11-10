package main

import (
	"context"

	"dagger/internal/dagger"
)

type PrrunProject struct {
	Dagger *Dagger // +private
}

func (p *PrrunProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("prrun")
	return buildProject(ctx, "prrun", ".", src)
}

func (p *PrrunProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("prrun")
	return testProject(ctx, "prrun", format, src)
}

func (p *PrrunProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("prrun")
	return coverageFile(ctx, "prrun", format, src)
}
