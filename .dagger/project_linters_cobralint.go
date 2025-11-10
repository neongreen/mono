package main

import (
	"context"

	"dagger/internal/dagger"
)

type LintersCobralintProject struct {
	Dagger *Dagger // +private
}

func (p *LintersCobralintProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("linters/cobralint")
	return buildProject(ctx, "linters/cobralint", "./cmd/cobralint", src)
}

func (p *LintersCobralintProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("linters/cobralint")
	return testProject(ctx, "linters/cobralint", format, src)
}

func (p *LintersCobralintProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("linters/cobralint")
	return coverageFile(ctx, "linters/cobralint", format, src)
}
