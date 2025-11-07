package main

import (
	"context"

	"dagger/internal/dagger"
)

type ClaudeTraceProject struct {
	Dagger *Dagger // +private
}

func (p *ClaudeTraceProject) Build(ctx context.Context) (*dagger.File, error) {
	return buildProject(ctx, "claude-trace", "./cmd")
}

func (p *ClaudeTraceProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	return testProject(ctx, "claude-trace", format)
}

func (p *ClaudeTraceProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	return coverageFile(ctx, "claude-trace", format)
}
