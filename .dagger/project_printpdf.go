package main

import (
	"context"

	"dagger/internal/dagger"
)

type PrintpdfProject struct {
	Dagger *Dagger // +private
}

func (p *PrintpdfProject) Build(ctx context.Context) (*dagger.File, error) {
	src := getFilteredSource("printpdf", dag.CurrentModule().Source().Directory(".."))
	return buildProject(ctx, "printpdf", "./cmd", src)
}

// Test runs tests for printpdf with PDF tools installed.
// Note: Golden tests (visual regression) are skipped as they're sensitive to environment differences.
func (p *PrintpdfProject) Test(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (string, error) {
	repo := dag.CurrentModule().Source().Directory("..")

	// Test only specific packages (skip golden tests)
	var args []string
	if format == "testname" {
		args = []string{"gotestsum", "--format", format, "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", "./printpdf/pkg/converter", "./printpdf/pkg/fetcher"}
	} else {
		args = []string{"gotestsum", "--format", format, "--", "-coverprofile=coverage.out", "-covermode=atomic", "./printpdf/pkg/converter", "./printpdf/pkg/fetcher"}
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
		// Install PDF processing tools
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "poppler-utils", "imagemagick", "python3-pip"}).
		WithExec([]string{"pip3", "install", "--break-system-packages", "weasyprint"}).
		WithExec([]string{"which", "pdftoppm"}).
		WithExec([]string{"which", "convert"}).
		WithExec([]string{"which", "weasyprint"}).
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@v1.13.0"}).
		WithExec(args).
		Stdout(ctx)
}

func (p *PrintpdfProject) Coverage(ctx context.Context,
	// +optional
	// +default="testname"
	format string,
) (*dagger.File, error) {
	repo := dag.CurrentModule().Source().Directory("..")

	// Test only specific packages (skip golden tests)
	var args []string
	if format == "testname" {
		args = []string{"gotestsum", "--format", format, "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", "./printpdf/pkg/converter", "./printpdf/pkg/fetcher"}
	} else {
		args = []string{"gotestsum", "--format", format, "--", "-coverprofile=coverage.out", "-covermode=atomic", "./printpdf/pkg/converter", "./printpdf/pkg/fetcher"}
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
		// Install PDF processing tools
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "poppler-utils", "imagemagick", "python3-pip"}).
		WithExec([]string{"pip3", "install", "--break-system-packages", "weasyprint"}).
		WithExec([]string{"which", "pdftoppm"}).
		WithExec([]string{"which", "convert"}).
		WithExec([]string{"which", "weasyprint"}).
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@v1.13.0"}).
		WithExec(args)

	return container.File("coverage.out"), nil
}
