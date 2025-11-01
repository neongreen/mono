// A generated module for DaggerTk functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"fmt"

	"dagger/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

type Dagger struct{}

// TkTests runs the tk package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) TkTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "tk", format, "./tk/...")
}

// DissectTests runs the dissect package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) DissectTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "dissect", format, "./dissect/...")
}

// TomlTests runs the lib/toml package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) TomlTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "toml", format, "./lib/toml/...")
}

// `All` runs the tk, dissect, and toml test suites concurrently.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) All(ctx context.Context,
	// +optional
	// +default="testname"
	format string) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(func(ctx context.Context) error {
		_, err := m.TkTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.DissectTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.TomlTests(ctx, format)
		return err
	})

	return p.Wait()
}

func (m *Dagger) goTest(ctx context.Context, label string, format string, patterns ...string) (string, error) {
	// Use the specified format for gotestsum output
	// Include -v flag for verbose test output when using testname format for better visibility
	var args []string
	if format == "testname" {
		args = append([]string{"gotestsum", "--format", format, "--", "-v"}, patterns...)
	} else {
		args = append([]string{"gotestsum", "--format", format, "--"}, patterns...)
	}
	return m.testContainer().
		WithLabel("suite", fmt.Sprintf("%s-tests", label)).
		WithExec(args).
		Stdout(ctx)
}

// testContainer prepares a Go build container with the monorepo mounted.
func (m *Dagger) testContainer() *dagger.Container {
	repo := dag.CurrentModule().Source().Directory("..")

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
		// Install gotestsum with pinned version
		// Cached in /go/bin volume, so it won't be rebuilt on every run
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@v1.13.0"})
}
