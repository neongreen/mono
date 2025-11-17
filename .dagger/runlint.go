package main

import (
	"context"
	"dagger/internal/dagger"
	"dagger/internal/parallel"
)

// Lint all Go projects - runs golangci-lint and uselesswrapper from workspace root and .dagger directory in parallel
//
// Uses --new-from-rev to grandfather existing gosec violations at commit 0c23a5a5
// (when gosec was first enabled). This allows 18 documented violations while
// preventing new ones. See tk-330 for details.
//
// Projects to lint can be specified as arguments (e.g., "tk", "conf").
// If no projects specified, lints all projects.
func (m *Dagger) Lint(ctx context.Context,
	// +optional
	projects []string,
) error {
	repo := dag.CurrentModule().Source().Directory("..")

	// Determine lint paths based on projects argument
	lintPaths := []string{"./..."}
	if len(projects) > 0 {
		lintPaths = make([]string, len(projects))
		for i, project := range projects {
			lintPaths[i] = "./" + project + "/..."
		}
	}

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
		args := []string{"golangci-lint", "run", "--new-from-rev=0c23a5a5"}
		args = append(args, lintPaths...)
		_, err := baseContainer.
			WithMountedDirectory("/src", repo).
			WithWorkdir("/src").
			WithExec([]string{"golangci-lint", "config", "verify"}).
			WithExec(args).
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
		args := []string{"go", "run", "./linters/uselesswrapper/cmd/uselesswrapper"}
		args = append(args, lintPaths...)
		_, err := goContainer.
			WithMountedDirectory("/src", repo).
			WithWorkdir("/src").
			WithExec(args).
			Sync(ctx)
		return err
	})

	// Lint tk project with cobralint (only if linting all projects or tk specifically)
	shouldLintTk := len(projects) == 0
	if !shouldLintTk {
		for _, p := range projects {
			if p == "tk" {
				shouldLintTk = true
				break
			}
		}
	}
	if shouldLintTk {
		jobs = jobs.WithJob("lint tk (cobralint)", func(ctx context.Context) error {
			_, err := goContainer.
				WithMountedDirectory("/src", repo).
				WithWorkdir("/src").
				WithExec([]string{"go", "run", "./linters/cobralint/cmd/cobralint", "./tk/..."}).
				Sync(ctx)
			return err
		})
	}

	// Lint .dagger module (only if linting all projects or dagger specifically)
	shouldLintDagger := len(projects) == 0
	if !shouldLintDagger {
		for _, p := range projects {
			if p == "dagger" || p == ".dagger" {
				shouldLintDagger = true
				break
			}
		}
	}
	if shouldLintDagger {
		jobs = jobs.WithJob("lint .dagger", func(ctx context.Context) error {
			_, err := baseContainer.
				WithMountedDirectory("/src", repo).
				WithWorkdir("/src/.dagger").
				WithExec([]string{"golangci-lint", "run", "--new-from-rev=0c23a5a5"}).
				Sync(ctx)
			return err
		})
	}

	return jobs.Run(ctx)
}

// ModernizeFix runs golangci-lint with --fix to automatically fix modernize issues
// Returns a Changeset showing what was fixed
func (m *Dagger) ModernizeFix(ctx context.Context) (*dagger.Changeset, error) {
	repo := dag.CurrentModule().Source().Directory("..")

	// Run golangci-lint with --fix flag for modernize linter
	fixedDir := dag.Container().
		From("docker.io/golangci/golangci-lint:v2.6.1").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint")).
		WithEnvVariable("GOPRIVATE", "github.com/neongreen/mono").
		WithEnvVariable("GONOSUMDB", "github.com/neongreen/mono").
		WithMountedDirectory("/src", repo).
		WithWorkdir("/src").
		WithExec([]string{
			"golangci-lint", "run",
			"--fix",
			"--enable-only=modernize",
		}).
		Directory("/src")

	// Return changeset showing what was fixed
	return fixedDir.Changes(repo), nil
}

// AutoFix applies all automatic code fixes: modernize, staticcheck, goimports, cargo fmt, and go mod tidy
// Returns a Changeset showing all fixes applied
func (m *Dagger) AutoFix(ctx context.Context) (*dagger.Changeset, error) {
	repo := dag.CurrentModule().Source().Directory("..")

	// Do all Go operations in a single container with golangci-lint
	goFixed := dag.Container().
		From("docker.io/golangci/golangci-lint:v2.6.1").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint")).
		WithEnvVariable("GOPRIVATE", "github.com/neongreen/mono").
		WithEnvVariable("GONOSUMDB", "github.com/neongreen/mono").
		WithMountedDirectory("/src", repo).
		WithWorkdir("/src").
		// Step 1: Run modernize and staticcheck fixes
		WithExec([]string{
			"golangci-lint", "run",
			"--fix",
			"--enable-only=modernize,staticcheck",
		}).
		// Step 2: Install and run goimports
		WithExec([]string{"go", "install", "golang.org/x/tools/cmd/goimports@latest"}).
		WithExec([]string{"/go/bin/goimports", "-w", "."}).
		// Step 3: Run go mod tidy
		WithExec([]string{"go", "mod", "tidy"}).
		Directory("/src")

	// Step 4: Format Rust code in mdbook-comments (if it exists)
	fullyFixed := dag.Container().
		From("rust:1.90").
		WithMountedDirectory("/src", goFixed).
		WithWorkdir("/src").
		WithExec([]string{"sh", "-c", "if [ -f mdbook-comments/Cargo.toml ]; then cd mdbook-comments && rustup component add rustfmt && cargo fmt; fi"}).
		Directory("/src")

	// Return changeset showing all fixes
	return fullyFixed.Changes(repo), nil
}
