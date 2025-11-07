package main

import (
	"context"

	"dagger/internal/dagger"
)

// LibConfigschemaProject represents the lib/configschema library
type LibConfigschemaProject struct {
	Dagger *Dagger // +private
}

// Test runs tests for lib/configschema
func (p *LibConfigschemaProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "lib/configschema", format)
}

// Coverage runs tests for lib/configschema and returns coverage file
func (p *LibConfigschemaProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "lib/configschema", format)
}
