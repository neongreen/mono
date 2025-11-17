package main

import (
	"context"

	"dagger/internal/dagger"
)

type LibTomlProject struct {
	Dagger *Dagger // +private
}

func (p *LibTomlProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("lib/toml", dag.CurrentModule().Source().Directory(".."))
	return testProject(ctx, "lib/toml", format, src)
}

func (p *LibTomlProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("lib/toml", dag.CurrentModule().Source().Directory(".."))
	return coverageFile(ctx, "lib/toml", format, src)
}
