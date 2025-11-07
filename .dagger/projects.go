package main

import (
	"context"

	"dagger/internal/parallel"
)

// Build all Go projects in parallel
func (m *Dagger) Build(ctx context.Context) error {
	jobs := parallel.New()
	p := m.Project()
	jobs = jobs.WithJob("build tk", func(ctx context.Context) error {
		_, err := p.Tk().Build(ctx)
		return err
	})
	jobs = jobs.WithJob("build want", func(ctx context.Context) error {
		_, err := p.Want().Build(ctx)
		return err
	})
	jobs = jobs.WithJob("build conf", func(ctx context.Context) error {
		_, err := p.Conf().Build(ctx)
		return err
	})
	jobs = jobs.WithJob("build dissect", func(ctx context.Context) error {
		_, err := p.Dissect().Build(ctx)
		return err
	})
	jobs = jobs.WithJob("build ingest", func(ctx context.Context) error {
		_, err := p.Ingest().Build(ctx)
		return err
	})
	jobs = jobs.WithJob("build printpdf", func(ctx context.Context) error {
		_, err := p.Printpdf().Build(ctx)
		return err
	})
	jobs = jobs.WithJob("build prrun", func(ctx context.Context) error {
		_, err := p.Prrun().Build(ctx)
		return err
	})
	jobs = jobs.WithJob("build claude-trace", func(ctx context.Context) error {
		_, err := p.ClaudeTrace().Build(ctx)
		return err
	})
	jobs = jobs.WithJob("build markdown-format", func(ctx context.Context) error {
		_, err := p.MarkdownFormat().Build(ctx)
		return err
	})
	jobs = jobs.WithJob("build beads-merge", func(ctx context.Context) error {
		_, err := p.BeadsMerge().Build(ctx)
		return err
	})
	jobs = jobs.WithJob("build jj-run", func(ctx context.Context) error {
		_, err := p.JjRun().Build(ctx)
		return err
	})
	return jobs.Run(ctx)
}

// Test all Go projects in parallel
func (m *Dagger) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) error {
	jobs := parallel.New()
	p := m.Project()
	jobs = jobs.WithJob("test tk", func(ctx context.Context) error {
		_, err := p.Tk().Test(ctx, format)
		return err
	})
	jobs = jobs.WithJob("test want", func(ctx context.Context) error {
		_, err := p.Want().Test(ctx, format)
		return err
	})
	jobs = jobs.WithJob("test conf", func(ctx context.Context) error {
		_, err := p.Conf().Test(ctx, format)
		return err
	})
	jobs = jobs.WithJob("test dissect", func(ctx context.Context) error {
		_, err := p.Dissect().Test(ctx, format)
		return err
	})
	jobs = jobs.WithJob("test ingest", func(ctx context.Context) error {
		_, err := p.Ingest().Test(ctx, format)
		return err
	})
	jobs = jobs.WithJob("test printpdf", func(ctx context.Context) error {
		_, err := p.Printpdf().Test(ctx, format)
		return err
	})
	jobs = jobs.WithJob("test prrun", func(ctx context.Context) error {
		_, err := p.Prrun().Test(ctx, format)
		return err
	})
	jobs = jobs.WithJob("test claude-trace", func(ctx context.Context) error {
		_, err := p.ClaudeTrace().Test(ctx, format)
		return err
	})
	jobs = jobs.WithJob("test markdown-format", func(ctx context.Context) error {
		_, err := p.MarkdownFormat().Test(ctx, format)
		return err
	})
	jobs = jobs.WithJob("test beads-merge", func(ctx context.Context) error {
		_, err := p.BeadsMerge().Test(ctx, format)
		return err
	})
	jobs = jobs.WithJob("test jj-run", func(ctx context.Context) error {
		_, err := p.JjRun().Test(ctx, format)
		return err
	})
	return jobs.Run(ctx)
}

// Lint all Go projects in parallel
func (m *Dagger) Lint(ctx context.Context) error {
	// Limit parallelism to avoid OOM (matching Dagger's approach)
	jobs := parallel.New().WithLimit(3)
	p := m.Project()
	
	// Verify golangci-lint config first
	jobs = jobs.WithJob("verify golangci-lint config", func(ctx context.Context) error {
		_, err := dag.Container().
			From("golangci/golangci-lint:v2.6.1").
			WithMountedDirectory("/workspace", dag.CurrentModule().Source().Directory("..")).
			WithWorkdir("/workspace").
			WithExec([]string{"golangci-lint", "config", "verify"}).
			Sync(ctx)
		return err
	})
	
	jobs = jobs.WithJob("lint tk", func(ctx context.Context) error {
		_, err := p.Tk().Lint(ctx)
		return err
	})
	jobs = jobs.WithJob("lint want", func(ctx context.Context) error {
		_, err := p.Want().Lint(ctx)
		return err
	})
	jobs = jobs.WithJob("lint conf", func(ctx context.Context) error {
		_, err := p.Conf().Lint(ctx)
		return err
	})
	jobs = jobs.WithJob("lint dissect", func(ctx context.Context) error {
		_, err := p.Dissect().Lint(ctx)
		return err
	})
	jobs = jobs.WithJob("lint ingest", func(ctx context.Context) error {
		_, err := p.Ingest().Lint(ctx)
		return err
	})
	jobs = jobs.WithJob("lint printpdf", func(ctx context.Context) error {
		_, err := p.Printpdf().Lint(ctx)
		return err
	})
	jobs = jobs.WithJob("lint prrun", func(ctx context.Context) error {
		_, err := p.Prrun().Lint(ctx)
		return err
	})
	jobs = jobs.WithJob("lint claude-trace", func(ctx context.Context) error {
		_, err := p.ClaudeTrace().Lint(ctx)
		return err
	})
	jobs = jobs.WithJob("lint markdown-format", func(ctx context.Context) error {
		_, err := p.MarkdownFormat().Lint(ctx)
		return err
	})
	jobs = jobs.WithJob("lint beads-merge", func(ctx context.Context) error {
		_, err := p.BeadsMerge().Lint(ctx)
		return err
	})
	jobs = jobs.WithJob("lint jj-run", func(ctx context.Context) error {
		_, err := p.JjRun().Lint(ctx)
		return err
	})
	return jobs.Run(ctx)
}
