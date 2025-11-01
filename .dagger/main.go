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
func (m *Dagger) TkTests(ctx context.Context) (string, error) {
	return m.goTest(ctx, "tk", "./tk/...")
}

// DissectTests runs the dissect package tests and returns the standard output.
func (m *Dagger) DissectTests(ctx context.Context) (string, error) {
	return m.goTest(ctx, "dissect", "./dissect/...")
}

// TomlTests runs the lib/toml package tests and returns the standard output.
func (m *Dagger) TomlTests(ctx context.Context) (string, error) {
	return m.goTest(ctx, "toml", "./lib/toml/...")
}

// `All` runs the tk, dissect, and toml test suites concurrently.
func (m *Dagger) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(func(ctx context.Context) error {
		_, err := m.TkTests(ctx)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.DissectTests(ctx)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.TomlTests(ctx)
		return err
	})

	return p.Wait()
}

func (m *Dagger) goTest(ctx context.Context, label string, patterns ...string) (string, error) {
	args := append([]string{"gotestsum", "--format", "pkgname", "--"}, patterns...)
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
		WithWorkdir("/src").
		WithEnvVariable("GOPRIVATE", "github.com/neongreen/mono").
		WithEnvVariable("GONOSUMDB", "github.com/neongreen/mono").
		WithEnvVariable("GOWORK", "off").
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@latest"})
}
