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
	return testProject(ctx, "lib/toml", format)
}

func (p *LibTomlProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "lib/toml", format)
}
