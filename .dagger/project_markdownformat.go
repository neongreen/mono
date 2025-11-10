package main

import (
	"context"

	"dagger/internal/dagger"
)

type MarkdownFormatProject struct {
	Dagger *Dagger // +private
}

func (p *MarkdownFormatProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("markdown-format")
	return buildProject(ctx, "markdown-format", ".", src)
}

func (p *MarkdownFormatProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("markdown-format")
	return testProject(ctx, "markdown-format", format, src)
}

func (p *MarkdownFormatProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("markdown-format")
	return coverageFile(ctx, "markdown-format", format, src)
}
