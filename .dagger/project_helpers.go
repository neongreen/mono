package main

import (
	"context"
	"fmt"

	"dagger/internal/dagger"
)

// buildProject builds a Go project for the current platform.
// projectName: the project directory (e.g., "tk", "want")
// buildTarget: the build target ("." or "./cmd")
// source: filtered source directory (use getFilteredSource helper)
//
//nolint:unparam // ctx kept for consistency with other helper functions
func buildProject(ctx context.Context, projectName string, buildTarget string, source *dagger.Directory) (*dagger.File, error) {
	return dag.Container().
		From("golang:1.24.7").
		WithMountedDirectory("/src", source).
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
// source: filtered source directory (use getFilteredSource helper)
// Always generates coverage.out internally (minimal overhead)
func testProject(ctx context.Context, projectName string, format string, source *dagger.Directory) (string, error) {
	ctr := testContainer(projectName, format, source)
	return ctr.Stdout(ctx)
}

// testContainer creates a container that runs tests with coverage.
// This container is reused by both Test() and Coverage() methods.
func testContainer(projectName string, format string, source *dagger.Directory) *dagger.Container {
	var args []string
	if format == "testname" {
		args = []string{"gotestsum", "--format", format, "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", fmt.Sprintf("./%s/...", projectName)}
	} else {
		args = []string{"gotestsum", "--format", format, "--", "-coverprofile=coverage.out", "-covermode=atomic", fmt.Sprintf("./%s/...", projectName)}
	}

	return dag.Container().
		From("golang:1.24.7").
		WithMountedDirectory("/src", source).
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

// getFilteredSource returns a filtered source directory for a project.
// projectName: the project directory (e.g., "tk", "want")
// This uses pre-call filtering to only include necessary files for optimal caching.
func getFilteredSource(
	projectName string,
	// Exclude everything except the project itself, lib, and Go module files
	// +ignore=["*", "!go.mod", "!go.sum", "!lib/**"]
	source *dagger.Directory,
) *dagger.Directory {
	// Add the specific project directory back in (ignore can't do this dynamically)
	root := dag.CurrentModule().Source().Directory("..")
	return source.WithDirectory(projectName, root.Directory(projectName))
}

// coverageFile returns the coverage file for a project.
// projectName: the project directory (e.g., "tk", "want")
// format: gotestsum format
// source: filtered source directory (use getFilteredSource helper)
//
//nolint:unparam // ctx kept for consistency with other helper functions
func coverageFile(ctx context.Context, projectName string, format string, source *dagger.Directory) (*dagger.File, error) {
	ctr := testContainer(projectName, format, source)
	return ctr.File("coverage.out"), nil
}

// testContainerWithSetup creates a container that runs tests with coverage, with custom setup steps.
// projectName: the project directory (e.g., "tk", "want")
// format: gotestsum format
// source: source directory (typically full repo for projects needing custom tools)
// setupFn: optional function to add setup steps (tool installation, etc.) to the container
func testContainerWithSetup(projectName string, format string, source *dagger.Directory, setupFn func(*dagger.Container) *dagger.Container) *dagger.Container {
	var args []string
	if format == "testname" {
		args = []string{"gotestsum", "--format", format, "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", fmt.Sprintf("./%s/...", projectName)}
	} else {
		args = []string{"gotestsum", "--format", format, "--", "-coverprofile=coverage.out", "-covermode=atomic", fmt.Sprintf("./%s/...", projectName)}
	}

	ctr := dag.Container().
		From("golang:1.24.7").
		WithMountedDirectory("/src", source).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/go/bin", dag.CacheVolume("go-bin")).
		WithWorkdir("/src").
		WithEnvVariable("GOPRIVATE", "github.com/neongreen/mono").
		WithEnvVariable("GONOSUMDB", "github.com/neongreen/mono").
		WithEnvVariable("GOWORK", "off")

	// Apply custom setup if provided
	if setupFn != nil {
		ctr = setupFn(ctr)
	}

	// Install gotestsum and run tests
	return ctr.
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@v1.13.0"}).
		WithExec(args)
}

// testProjectWithSetup runs tests for a Go project with custom setup.
// projectName: the project directory (e.g., "tk", "want")
// format: gotestsum format
// source: source directory (typically full repo for projects needing custom tools)
// setupFn: optional function to add setup steps (tool installation, etc.) to the container
func testProjectWithSetup(ctx context.Context, projectName string, format string, source *dagger.Directory, setupFn func(*dagger.Container) *dagger.Container) (string, error) {
	ctr := testContainerWithSetup(projectName, format, source, setupFn)
	return ctr.Stdout(ctx)
}

// coverageFileWithSetup returns the coverage file for a project with custom setup.
// projectName: the project directory (e.g., "tk", "want")
// format: gotestsum format
// source: source directory (typically full repo for projects needing custom tools)
// setupFn: optional function to add setup steps (tool installation, etc.) to the container
//
//nolint:unparam // ctx kept for consistency with other helper functions
func coverageFileWithSetup(ctx context.Context, projectName string, format string, source *dagger.Directory, setupFn func(*dagger.Container) *dagger.Container) (*dagger.File, error) {
	ctr := testContainerWithSetup(projectName, format, source, setupFn)
	return ctr.File("coverage.out"), nil
}
