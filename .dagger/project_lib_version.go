package main

import (
	"context"

	"dagger/internal/dagger"
)

// LibVersionProject represents the lib/version library
type LibVersionProject struct {
	Dagger *Dagger // +private
}

// Test runs tests for lib/version
func (p *LibVersionProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "lib/version", format)
}

// Coverage runs tests for lib/version and returns coverage file
func (p *LibVersionProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "lib/version", format)
}
