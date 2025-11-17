package main

import (
	"context"

	"dagger/internal/dagger"
)

type JjRunProject struct {
	Dagger *Dagger // +private
}

func (p *JjRunProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("jj-run", dag.CurrentModule().Source().Directory(".."))
	return buildProject(ctx, "jj-run", ".", src)
}

// Test runs tests for jj-run with jujutsu installed.
func (p *JjRunProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	repo := dag.CurrentModule().Source().Directory("..")

	var args []string
	if format == "testname" {
		args = []string{"gotestsum", "--format", format, "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", "./jj-run/..."}
	} else {
		args = []string{"gotestsum", "--format", format, "--", "-coverprofile=coverage.out", "-covermode=atomic", "./jj-run/..."}
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
		// Install jj (jujutsu)
		WithExec([]string{"sh", "-c", "curl -L https://github.com/jj-vcs/jj/releases/download/v0.34.0/jj-v0.34.0-x86_64-unknown-linux-musl.tar.gz | tar xz && mv jj /usr/local/bin/"}).
		WithExec([]string{"jj", "--version"}).
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@v1.13.0"}).
		WithExec(args).
		Stdout(ctx)
}

//nolint:unparam // ctx kept for Dagger interface consistency
func (p *JjRunProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	repo := dag.CurrentModule().Source().Directory("..")

	var args []string
	if format == "testname" {
		args = []string{"gotestsum", "--format", format, "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", "./jj-run/..."}
	} else {
		args = []string{"gotestsum", "--format", format, "--", "-coverprofile=coverage.out", "-covermode=atomic", "./jj-run/..."}
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
		// Install jj (jujutsu)
		WithExec([]string{"sh", "-c", "curl -L https://github.com/jj-vcs/jj/releases/download/v0.34.0/jj-v0.34.0-x86_64-unknown-linux-musl.tar.gz | tar xz && mv jj /usr/local/bin/"}).
		WithExec([]string{"jj", "--version"}).
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@v1.13.0"}).
		WithExec(args)

	return container.File("coverage.out"), nil
}
