package main

import (
	"context"

	"dagger/internal/dagger"
)

type LintersCobralintProject struct {
	Dagger *Dagger // +private
}

func (p *LintersCobralintProject) Build(ctx context.Context) (*dagger.File, error) {
	return buildProject(ctx, "linters/cobralint", "./cmd/cobralint")
}

func (p *LintersCobralintProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "linters/cobralint", format)
}

func (p *LintersCobralintProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "linters/cobralint", format)
}
