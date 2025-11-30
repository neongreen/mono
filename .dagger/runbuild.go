package main

import (
	"context"
	"dagger/internal/dagger"
	"dagger/internal/parallel"
)

// Build all Go projects in parallel
func (m *Dagger) Build(ctx context.Context) error {
	jobs := parallel.New()
	p := m.Project()

	addBuild := func(name string, buildFn func(context.Context) (*dagger.File, error)) {
		jobs = jobs.WithJob(name, func(ctx context.Context) error {
			_, err := buildFn(ctx)
			return err
		})
	}

	addBuild("build tk", p.Tk().Build)
	addBuild("build want", p.Want().Build)
	addBuild("build conf", p.Conf().Build)
	addBuild("build dissect", p.Dissect().Build)
	addBuild("build ingest", p.Ingest().Build)
	addBuild("build printpdf", p.Printpdf().Build)
	addBuild("build prrun", p.Prrun().Build)
	addBuild("build markdown-format", p.MarkdownFormat().Build)
	addBuild("build jj-run", p.JjRun().Build)

	return jobs.Run(ctx)
}
