package main

import (
	"context"

	"dagger/internal/dagger"
)

type BeadsMergeProject struct {
	Dagger *Dagger // +private
}

func (p *BeadsMergeProject) Build(ctx context.Context) (*dagger.File, error) {
	return buildProject(ctx, "beads-merge", ".")
}

func (p *BeadsMergeProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "beads-merge", format)
}

func (p *BeadsMergeProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "beads-merge", format)
}

func (p *BeadsMergeProject) Lint(ctx context.Context) (string, error) {
	return lintProject(ctx, "beads-merge")
}

