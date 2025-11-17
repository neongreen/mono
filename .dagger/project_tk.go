package main

import (
	"context"
	"fmt"

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

// WasmBuild builds the tk WASM binary and returns a directory with the compiled assets
//
//nolint:unparam // ctx kept for Dagger interface consistency
func (p *TkProject) WasmBuild(ctx context.Context) (*dagger.Directory, error) {
	repo := dag.CurrentModule().Source().Directory("..")

	// Build WASM binary
	ctr := dag.Container().
		From("golang:1.24.7").
		WithMountedDirectory("/src", repo).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithWorkdir("/src/tk").
		WithEnvVariable("GOPRIVATE", "github.com/neongreen/mono").
		WithEnvVariable("GONOSUMDB", "github.com/neongreen/mono").
		WithEnvVariable("GOWORK", "off").
		WithEnvVariable("GOOS", "js").
		WithEnvVariable("GOARCH", "wasm").
		WithExec([]string{"go", "build", "-o", "/output/tk.wasm", "./wasm-demo/main.go"})

	// Get wasm_exec.js from Go installation
	wasmExecJS := ctr.File("/usr/local/go/lib/wasm/wasm_exec.js")

	// Create output directory with WASM binary, wasm_exec.js, and HTML files
	return dag.Directory().
		WithFile("tk.wasm", ctr.File("/output/tk.wasm")).
		WithFile("wasm_exec.js", wasmExecJS).
		WithFile("index.html", repo.File("tk/wasm-demo/index.html")), nil
}

// WasmServe builds the tk WASM demo and serves it on the specified port
func (p *TkProject) WasmServe(ctx context.Context,
	// +optional
	// +default=8080
	port int,
) (*dagger.Service, error) {
	// Build the WASM assets
	assets, err := p.WasmBuild(ctx)
	if err != nil {
		return nil, err
	}

	// Convert port to string for command
	portStr := fmt.Sprintf("%d", port)

	// Serve the assets using Python's http.server
	return dag.Container().
		From("python:3.12-slim").
		WithMountedDirectory("/app", assets).
		WithWorkdir("/app").
		WithExposedPort(port).
		WithExec([]string{"python3", "-m", "http.server", "-b", "0.0.0.0", "--directory", ".", portStr}).
		AsService(), nil
}
