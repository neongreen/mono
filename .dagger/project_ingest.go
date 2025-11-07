package main

import (
	"context"

	"dagger/internal/dagger"
)

type IngestProject struct {
	Dagger *Dagger // +private
}

func (p *IngestProject) Build(ctx context.Context) (*dagger.File, error) {
	return buildProject(ctx, "ingest", "./cmd")
}

// Test runs tests for ingest including integration tests.
func (p *IngestProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	repo := dag.CurrentModule().Source().Directory("..")

	// Run both unit tests and integration tests
	var args []string
	if format == "testname" {
		args = []string{"gotestsum", "--format", format, "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", "./ingest/..."}
	} else {
		args = []string{"gotestsum", "--format", format, "--", "-coverprofile=coverage.out", "-covermode=atomic", "./ingest/..."}
	}

	return dag.Container().
		From("golang:1.24.7").
		WithMountedDirectory("/src", repo).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/go/bin", dag.CacheVolume("go-bin")).
		WithWorkdir("/src").
		WithEnvVariable("GOPRIVATE", "github.com/neongreen/mono").
		WithEnvVariable("GONOSUMDB", "github.com/neongreen/mono").
		WithEnvVariable("GOWORK", "off").
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@v1.13.0"}).
		WithExec(args).
		Stdout(ctx)
}

func (p *IngestProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	repo := dag.CurrentModule().Source().Directory("..")

	// Run both unit tests and integration tests
	var args []string
	if format == "testname" {
		args = []string{"gotestsum", "--format", format, "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", "./ingest/..."}
	} else {
		args = []string{"gotestsum", "--format", format, "--", "-coverprofile=coverage.out", "-covermode=atomic", "./ingest/..."}
	}

	container := dag.Container().
		From("golang:1.24.7").
		WithMountedDirectory("/src", repo).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/go/bin", dag.CacheVolume("go-bin")).
		WithWorkdir("/src").
		WithEnvVariable("GOPRIVATE", "github.com/neongreen/mono").
		WithEnvVariable("GONOSUMDB", "github.com/neongreen/mono").
		WithEnvVariable("GOWORK", "off").
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@v1.13.0"}).
		WithExec(args)

	return container.File("coverage.out"), nil
}
