package main

import (
	"context"

	"dagger/internal/dagger"
)

type MarkdownFormatProject struct {
	Dagger *Dagger // +private
}

func (p *MarkdownFormatProject) Build(ctx context.Context) (*dagger.File, error) {
	return buildProject(ctx, "markdown-format", ".")
}

func (p *MarkdownFormatProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "markdown-format", format)
}

func (p *MarkdownFormatProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "markdown-format", format)
}
