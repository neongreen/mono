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

// JjRunTests runs the jj-run package tests and returns the standard output.
// This requires jj (jujutsu) to be installed in the container.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) JjRunTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	// Use the specified format for gotestsum output
	var args []string
	if format == "testname" {
		args = append([]string{"gotestsum", "--format", format, "--", "-v"}, "./jj-run/...")
	} else {
		args = append([]string{"gotestsum", "--format", format, "--"}, "./jj-run/...")
	}

	return m.jjRunTestContainer().
		WithLabel("suite", "jj-run-tests").
		WithExec(args).
		Stdout(ctx)
}

// ConfTests runs the conf package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) ConfTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "conf", format, "./conf/...")
}

// IngestTests runs the ingest package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) IngestTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "ingest", format, "./ingest/...")
}

// PrintpdfTests runs the printpdf package tests and returns the standard output.
// This requires poppler-utils, imagemagick, and weasyprint to be installed in the container.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) PrintpdfTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	// Use the specified format for gotestsum output
	var args []string
	if format == "testname" {
		args = append([]string{"gotestsum", "--format", format, "--", "-v"}, "./printpdf/...")
	} else {
		args = append([]string{"gotestsum", "--format", format, "--"}, "./printpdf/...")
	}

	return m.printpdfTestContainer().
		WithLabel("suite", "printpdf-tests").
		WithExec(args).
		Stdout(ctx)
}

// PrrunTests runs the prrun package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) PrrunTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "prrun", format, "./prrun/...")
}

// ClaudeTraceTests runs the claude-trace package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) ClaudeTraceTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "claude-trace", format, "./claude-trace/...")
}

// WantTests runs the want package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) WantTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "want", format, "./want/...")
}

// MarkdownFormatTests runs the markdown-format package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) MarkdownFormatTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "markdown-format", format, "./markdown-format/...")
}

// BeadsMergeTests runs the beads-merge package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) BeadsMergeTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "beads-merge", format, "./beads-merge/...")
}

// GhclientTests runs the lib/ghclient package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) GhclientTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "ghclient", format, "./lib/ghclient/...")
}

// GhreleaseTests runs the lib/ghrelease package tests and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) GhreleaseTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	return m.goTest(ctx, "ghrelease", format, "./lib/ghrelease/...")
}

// jjRunTestContainer prepares a Go build container with jj installed for jj-run tests.
func (m *Dagger) jjRunTestContainer() *dagger.Container {
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
		// Install jj (jujutsu) from GitHub releases
		WithExec([]string{"sh", "-c", "curl -L https://github.com/jj-vcs/jj/releases/download/v0.34.0/jj-v0.34.0-x86_64-unknown-linux-musl.tar.gz | tar xz && mv jj /usr/local/bin/"}).
		WithExec([]string{"jj", "--version"}).
		// Install gotestsum with pinned version
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@v1.13.0"})
}

// printpdfTestContainer prepares a Go build container with PDF tools installed for printpdf tests.
func (m *Dagger) printpdfTestContainer() *dagger.Container {
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
		// Install system dependencies for PDF processing
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "poppler-utils", "imagemagick", "python3-pip"}).
		// Install weasyprint for PDF generation
		WithExec([]string{"pip3", "install", "--break-system-packages", "weasyprint"}).
		// Verify tools are available
		WithExec([]string{"which", "pdftoppm"}).
		WithExec([]string{"which", "convert"}).
		WithExec([]string{"which", "weasyprint"}).
		// Install gotestsum with pinned version
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@v1.13.0"})
}

// `All` runs all test suites concurrently.
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
	p.Go(func(ctx context.Context) error {
		_, err := m.JjRunTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.ConfTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.IngestTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.PrintpdfTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.PrrunTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.ClaudeTraceTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.WantTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.MarkdownFormatTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.BeadsMergeTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.GhclientTests(ctx, format)
		return err
	})
	p.Go(func(ctx context.Context) error {
		_, err := m.GhreleaseTests(ctx, format)
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
