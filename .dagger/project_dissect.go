package main

import (
	"context"

	"dagger/internal/dagger"
)

type DissectProject struct {
	Dagger *Dagger // +private
}

func (p *DissectProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("dissect", dag.CurrentModule().Source().Directory(".."))
	return buildProject(ctx, "dissect", "./cmd", src)
}

func (p *DissectProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("dissect", dag.CurrentModule().Source().Directory(".."))
	return testProject(ctx, "dissect", format, src)
}

func (p *DissectProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("dissect", dag.CurrentModule().Source().Directory(".."))
	return coverageFile(ctx, "dissect", format, src)
}
