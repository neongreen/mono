package main

import (
	"context"

	"dagger/internal/dagger"
)

type ConfProject struct {
	Dagger *Dagger // +private
}

func (p *ConfProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("conf", dag.CurrentModule().Source().Directory(".."))
	return buildProject(ctx, "conf", "./cmd/main.go", src)
}

func (p *ConfProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("conf", dag.CurrentModule().Source().Directory(".."))
	return testProject(ctx, "conf", format, src)
}

func (p *ConfProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("conf", dag.CurrentModule().Source().Directory(".."))
	return coverageFile(ctx, "conf", format, src)
}
