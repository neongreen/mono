package main

import (
	"context"
	"dagger/internal/parallel"
)

// Lint all Go projects - runs golangci-lint and uselesswrapper from workspace root and .dagger directory in parallel
func (m *Dagger) Lint(ctx context.Context) error {
	repo := dag.CurrentModule().Source().Directory("..")

	baseContainer := dag.Container().
		From("docker.io/golangci/golangci-lint:v2.6.1").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint")).
		WithEnvVariable("GOPRIVATE", "github.com/neongreen/mono").
		WithEnvVariable("GONOSUMDB", "github.com/neongreen/mono")

	jobs := parallel.New()

	// Lint workspace root with golangci-lint
	jobs = jobs.WithJob("lint projects (golangci-lint)", func(ctx context.Context) error {
		_, err := baseContainer.
			WithMountedDirectory("/src", repo).
			WithWorkdir("/src").
			WithExec([]string{"golangci-lint", "config", "verify"}).
			WithExec([]string{"golangci-lint", "run"}).
			Sync(ctx)
		return err
	})

	// Shared Go container for custom linters
	goContainer := dag.Container().
		From("golang:1.24.7").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithEnvVariable("GOPRIVATE", "github.com/neongreen/mono").
		WithEnvVariable("GONOSUMDB", "github.com/neongreen/mono")

	// Lint workspace root with uselesswrapper
	jobs = jobs.WithJob("lint projects (uselesswrapper)", func(ctx context.Context) error {
		_, err := goContainer.
			WithMountedDirectory("/src", repo).
			WithWorkdir("/src").
			WithExec([]string{"go", "run", "./linters/uselesswrapper/cmd/uselesswrapper", "./..."}).
			Sync(ctx)
		return err
	})

	// Lint workspace root with cobralint
	jobs = jobs.WithJob("lint projects (cobralint)", func(ctx context.Context) error {
		_, err := goContainer.
			WithMountedDirectory("/src", repo).
			WithWorkdir("/src").
			WithExec([]string{"go", "run", "./linters/cobralint/cmd/cobralint", "./..."}).
			Sync(ctx)
		return err
	})

	// Lint .dagger module
	jobs = jobs.WithJob("lint .dagger", func(ctx context.Context) error {
		_, err := baseContainer.
			WithMountedDirectory("/src", repo).
			WithWorkdir("/src/.dagger").
			WithExec([]string{"golangci-lint", "run"}).
			Sync(ctx)
		return err
	})

	return jobs.Run(ctx)
}
