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

// TkCoverage runs the tk package tests and exports the coverage report.
func (m *Dagger) TkCoverage(ctx context.Context) (*dagger.File, error) {
	return m.goTestCoverage(ctx, "tk", "./tk/...")
}

// DissectCoverage runs the dissect package tests and exports the coverage report.
func (m *Dagger) DissectCoverage(ctx context.Context) (*dagger.File, error) {
	return m.goTestCoverage(ctx, "dissect", "./dissect/...")
}

// TomlCoverage runs the lib/toml package tests and exports the coverage report.
func (m *Dagger) TomlCoverage(ctx context.Context) (*dagger.File, error) {
	return m.goTestCoverage(ctx, "toml", "./lib/toml/...")
}

// ConfCoverage runs the conf package tests and exports the coverage report.
func (m *Dagger) ConfCoverage(ctx context.Context) (*dagger.File, error) {
	// Include -race flag for conf as per original workflow
	container := m.testContainer().
		WithLabel("suite", "conf-tests").
		WithExec([]string{"gotestsum", "--format", "testname", "--", "-v", "-race", "-coverprofile=coverage.out", "-covermode=atomic", "./conf/..."})

	return container.File("coverage.out"), nil
}

// IngestCoverage runs the ingest package tests and exports the coverage report.
func (m *Dagger) IngestCoverage(ctx context.Context) (*dagger.File, error) {
	return m.goTestCoverage(ctx, "ingest", "./ingest/...", "./ingest/cmd", "./ingest/pkg/integration")
}

// PrintpdfCoverage runs the printpdf package tests and exports the coverage report.
// Note: Golden tests (visual regression) are skipped as they're sensitive to environment differences.
func (m *Dagger) PrintpdfCoverage(ctx context.Context) (*dagger.File, error) {
	container := m.printpdfTestContainer().
		WithLabel("suite", "printpdf-tests").
		WithExec([]string{"gotestsum", "--format", "testname", "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", "./printpdf/pkg/converter", "./printpdf/pkg/fetcher"})

	return container.File("coverage.out"), nil
}

// PrrunCoverage runs the prrun package tests and exports the coverage report.
func (m *Dagger) PrrunCoverage(ctx context.Context) (*dagger.File, error) {
	return m.goTestCoverage(ctx, "prrun", "./prrun/...")
}

// ClaudeTraceCoverage runs the claude-trace package tests and exports the coverage report.
func (m *Dagger) ClaudeTraceCoverage(ctx context.Context) (*dagger.File, error) {
	return m.goTestCoverage(ctx, "claude-trace", "./claude-trace/...")
}

// WantCoverage runs the want package tests and exports the coverage report.
func (m *Dagger) WantCoverage(ctx context.Context) (*dagger.File, error) {
	return m.goTestCoverage(ctx, "want", "./want/...")
}

// MarkdownFormatCoverage runs the markdown-format package tests and exports the coverage report.
func (m *Dagger) MarkdownFormatCoverage(ctx context.Context) (*dagger.File, error) {
	return m.goTestCoverage(ctx, "markdown-format", "./markdown-format/...")
}

// BeadsMergeCoverage runs the beads-merge package tests and exports the coverage report.
func (m *Dagger) BeadsMergeCoverage(ctx context.Context) (*dagger.File, error) {
	return m.goTestCoverage(ctx, "beads-merge", "./beads-merge/...")
}

// GhclientCoverage runs the lib/ghclient package tests and exports the coverage report.
func (m *Dagger) GhclientCoverage(ctx context.Context) (*dagger.File, error) {
	return m.goTestCoverage(ctx, "ghclient", "./lib/ghclient/...")
}

// GhreleaseCoverage runs the lib/ghrelease package tests and exports the coverage report.
func (m *Dagger) GhreleaseCoverage(ctx context.Context) (*dagger.File, error) {
	return m.goTestCoverage(ctx, "ghrelease", "./lib/ghrelease/...")
}

// JjRunCoverage runs the jj-run package tests and exports the coverage report.
func (m *Dagger) JjRunCoverage(ctx context.Context) (*dagger.File, error) {
	container := m.jjRunTestContainer().
		WithLabel("suite", "jj-run-tests").
		WithExec([]string{"gotestsum", "--format", "testname", "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", "./jj-run/..."})

	return container.File("coverage.out"), nil
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

// IngestTests runs the ingest package tests (including integration tests) and returns the standard output.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) IngestTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	// Run both unit tests and integration tests
	return m.goTest(ctx, "ingest", format, "./ingest/...", "./ingest/cmd", "./ingest/pkg/integration")
}

// PrintpdfTests runs the printpdf package tests and returns the standard output.
// This requires poppler-utils, imagemagick, and weasyprint to be installed in the container.
// Note: Golden tests (visual regression) are skipped as they're sensitive to environment differences.
// format: gotestsum output format (e.g., "testname", "pkgname", "dots", "standard-verbose")
func (m *Dagger) PrintpdfTests(ctx context.Context,
	// +optional
	// +default="testname"
	format string) (string, error) {
	// Use the specified format for gotestsum output
	// Skip golden tests (pkg/golden) as they're environment-sensitive visual regression tests
	var args []string
	if format == "testname" {
		args = append([]string{"gotestsum", "--format", format, "--", "-v"}, "./printpdf/pkg/converter", "./printpdf/pkg/fetcher")
	} else {
		args = append([]string{"gotestsum", "--format", format, "--"}, "./printpdf/pkg/converter", "./printpdf/pkg/fetcher")
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
	// Add coverage flags to collect coverage data
	var args []string
	if format == "testname" {
		args = append([]string{"gotestsum", "--format", format, "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic"}, patterns...)
	} else {
		args = append([]string{"gotestsum", "--format", format, "--", "-coverprofile=coverage.out", "-covermode=atomic"}, patterns...)
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

// goTestCoverage runs tests and returns the coverage file.
func (m *Dagger) goTestCoverage(ctx context.Context, label string, patterns ...string) (*dagger.File, error) {
	container := m.testContainer().
		WithLabel("suite", fmt.Sprintf("%s-tests", label)).
		WithExec(append([]string{"gotestsum", "--format", "testname", "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic"}, patterns...))

	return container.File("coverage.out"), nil
}
