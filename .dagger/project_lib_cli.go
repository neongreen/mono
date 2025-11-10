package main

import (
	"context"

	"dagger/internal/dagger"
)

// LibCliProject represents the lib/cli library
type LibCliProject struct {
	Dagger *Dagger // +private
}

// Test runs tests for lib/cli
func (p *LibCliProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("lib/cli", dag.CurrentModule().Source().Directory(".."))
	return testProject(ctx, "lib/cli", format, src)
}

// Coverage runs tests for lib/cli and returns coverage file
func (p *LibCliProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("lib/cli", dag.CurrentModule().Source().Directory(".."))
	return coverageFile(ctx, "lib/cli", format, src)
}
