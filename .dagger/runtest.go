package main

import (
	"context"

	"dagger/internal/parallel"
)

// Test all Go projects in parallel
func (m *Dagger) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) error {
	jobs := parallel.New()
	p := m.Project()

	addTest := func(name string, testFn func(context.Context, string) (string, error)) {
		jobs = jobs.WithJob(name, func(ctx context.Context) error {
			_, err := testFn(ctx, format)
			return err
		})
	}

	addTest("test tk", p.Tk().Test)
	addTest("test want", p.Want().Test)
	addTest("test conf", p.Conf().Test)
	addTest("test dissect", p.Dissect().Test)
	addTest("test ingest", p.Ingest().Test)
	addTest("test printpdf", p.Printpdf().Test)
	addTest("test prrun", p.Prrun().Test)
	addTest("test claude-trace", p.ClaudeTrace().Test)
	addTest("test markdown-format", p.MarkdownFormat().Test)
	addTest("test jj-run", p.JjRun().Test)
	addTest("test lib/ghclient", p.LibGhclient().Test)
	addTest("test lib/ghrelease", p.LibGhrelease().Test)
	addTest("test lib/toml", p.LibToml().Test)

	return jobs.Run(ctx)
}
