package main

import (
	"context"

	"dagger/internal/dagger"
)

type ClaudeTraceProject struct {
	Dagger *Dagger // +private
}

func (p *ClaudeTraceProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("claude-trace", dag.CurrentModule().Source().Directory(".."))
	return buildProject(ctx, "claude-trace", "./cmd", src)
}

func (p *ClaudeTraceProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := getFilteredSource("claude-trace", dag.CurrentModule().Source().Directory(".."))
	return testProject(ctx, "claude-trace", format, src)
}

func (p *ClaudeTraceProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := getFilteredSource("claude-trace", dag.CurrentModule().Source().Directory(".."))
	return coverageFile(ctx, "claude-trace", format, src)
}
