package main

import (
	"context"

	"dagger/internal/dagger"
)

type TkProject struct {
	Dagger *Dagger // +private
}

func (p *TkProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("tk", dag.CurrentModule().Source().Directory(".."))
	return buildProject(ctx, "tk", ".", src)
}

func (p *TkProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	src := dag.CurrentModule().Source().Directory("..")
	return testProjectWithSetup(ctx, "tk", format, src, func(ctr *dagger.Container) *dagger.Container {
		// Install DuckDB CLI for testing
		return ctr.
			WithExec([]string{"apt-get", "update"}).
			WithExec([]string{"apt-get", "install", "-y", "wget", "unzip"}).
			WithExec([]string{"wget", "https://github.com/duckdb/duckdb/releases/download/v1.1.3/duckdb_cli-linux-amd64.zip"}).
			WithExec([]string{"unzip", "duckdb_cli-linux-amd64.zip"}).
			WithExec([]string{"mv", "duckdb", "/usr/local/bin/duckdb"}).
			WithExec([]string{"chmod", "+x", "/usr/local/bin/duckdb"})
	})
}

func (p *TkProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	src := dag.CurrentModule().Source().Directory("..")
	return coverageFileWithSetup(ctx, "tk", format, src, func(ctr *dagger.Container) *dagger.Container {
		// Install DuckDB CLI for testing
		return ctr.
			WithExec([]string{"apt-get", "update"}).
			WithExec([]string{"apt-get", "install", "-y", "wget", "unzip"}).
			WithExec([]string{"wget", "https://github.com/duckdb/duckdb/releases/download/v1.1.3/duckdb_cli-linux-amd64.zip"}).
			WithExec([]string{"unzip", "duckdb_cli-linux-amd64.zip"}).
			WithExec([]string{"mv", "duckdb", "/usr/local/bin/duckdb"}).
			WithExec([]string{"chmod", "+x", "/usr/local/bin/duckdb"})
	})
}
