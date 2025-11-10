package main

import (
	"context"

	"dagger/internal/dagger"
)

type WantProject struct {
	Dagger *Dagger // +private
}

func (p *WantProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("want")
	return buildProject(ctx, "want", "./cmd", src)
}

func (p *WantProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("want")
	return testProject(ctx, "want", format, src)
}

func (p *WantProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("want")
	return coverageFile(ctx, "want", format, src)
}
