package main

import (
	"context"
	"fmt"

	"dagger/internal/dagger"
)

// buildProject builds a Go project for the current platform.
// projectName: the project directory (e.g., "tk", "want")
// buildTarget: the build target ("." or "./cmd")
func buildProject(ctx context.Context, projectName string, buildTarget string) (*dagger.File, error) {
	repo := dag.CurrentModule().Source().Directory("..")

	return dag.Container().
		From("golang:1.24.7").
		WithMountedDirectory("/src", repo).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithWorkdir(fmt.Sprintf("/src/%s", projectName)).
		WithEnvVariable("GOPRIVATE", "github.com/neongreen/mono").
		WithEnvVariable("GONOSUMDB", "github.com/neongreen/mono").
		WithEnvVariable("GOWORK", "off").
		WithExec([]string{"go", "build", "-o", fmt.Sprintf("/output/%s", projectName), buildTarget}).
		File(fmt.Sprintf("/output/%s", projectName)), nil
}

// testProject runs tests for a Go project.
// projectName: the project directory (e.g., "tk", "want")
// format: gotestsum format
// Always generates coverage.out internally (minimal overhead)
func testProject(ctx context.Context, projectName string, format string) (string, error) {
	ctr := testContainer(projectName, format)
	return ctr.Stdout(ctx)
}

// testContainer creates a container that runs tests with coverage.
// This container is reused by both Test() and Coverage() methods.
func testContainer(projectName string, format string) *dagger.Container {
	repo := dag.CurrentModule().Source().Directory("..")

	var args []string
	if format == "testname" {
		args = []string{"gotestsum", "--format", format, "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", fmt.Sprintf("./%s/...", projectName)}
	} else {
		args = []string{"gotestsum", "--format", format, "--", "-coverprofile=coverage.out", "-covermode=atomic", fmt.Sprintf("./%s/...", projectName)}
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
		WithExec(args)
}

// coverageFile returns the coverage file for a project.
// projectName: the project directory (e.g., "tk", "want")
// format: gotestsum format
func coverageFile(ctx context.Context, projectName string, format string) (*dagger.File, error) {
	ctr := testContainer(projectName, format)
	return ctr.File("coverage.out"), nil
}
