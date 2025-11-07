package main

import (
	"context"
	"dagger/internal/parallel"
)

// Lint all Go projects - runs golangci-lint from workspace root and .dagger directory in parallel
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

	// Lint workspace root
	jobs = jobs.WithJob("lint projects", func(ctx context.Context) error {
		_, err := baseContainer.
			WithMountedDirectory("/src", repo).
			WithWorkdir("/src").
			WithExec([]string{"golangci-lint", "config", "verify"}).
			WithExec([]string{"golangci-lint", "run"}).
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
