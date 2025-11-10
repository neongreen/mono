package main

import (
	"context"

	"dagger/internal/dagger"
)

type TkProject struct {
	Dagger *Dagger // +private
}

func (p *TkProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("tk", dag.CurrentModule().Source().Directory(".."))
	return buildProject(ctx, "tk", ".", src)
}

func (p *TkProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("tk", dag.CurrentModule().Source().Directory(".."))
	return testProject(ctx, "tk", format, src)
}

func (p *TkProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("tk", dag.CurrentModule().Source().Directory(".."))
	return coverageFile(ctx, "tk", format, src)
}
