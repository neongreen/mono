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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

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
func (m *Dagger) PrintpdfCoverage(ctx context.Context) (*dagger.File, error) {
	container := m.printpdfTestContainer().
		WithLabel("suite", "printpdf-tests").
		WithExec([]string{"gotestsum", "--format", "testname", "--", "-v", "-coverprofile=coverage.out", "-covermode=atomic", "./printpdf/..."})

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

// Platform represents a target platform for cross-compilation
type Platform struct {
	OS   string
	Arch string
}

// BuildRelease builds binaries for all platforms and returns them as a directory.
// project: the project directory to build (e.g., "tk", "ingest")
// version: the version string to embed in the binary
// gitCommit: the git commit hash to embed in the binary
func (m *Dagger) BuildRelease(ctx context.Context, project string, version string, gitCommit string) (*dagger.Directory, error) {
	repo := dag.CurrentModule().Source().Directory("..")
	buildTime := time.Now().UTC().Format(time.RFC3339)

	platforms := []Platform{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
	}

	// Create a base container for building
	baseContainer := dag.Container().
		From("golang:1.24.7").
		WithMountedDirectory("/src", repo).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithWorkdir("/src").
		WithEnvVariable("GOPRIVATE", "github.com/neongreen/mono").
		WithEnvVariable("GONOSUMDB", "github.com/neongreen/mono").
		WithEnvVariable("GOWORK", "off")

	// Determine build target
	buildTarget := "."
	cmdDir := filepath.Join(project, "cmd")
	if entries, err := repo.Directory(cmdDir).Entries(ctx); err == nil && len(entries) > 0 {
		buildTarget = "./cmd"
	}

	// Build for each platform and collect binaries
	distContainer := dag.Container().From("alpine:latest").WithWorkdir("/dist")

	for _, platform := range platforms {
		outputName := fmt.Sprintf("%s-%s-%s-%s", project, version, platform.OS, platform.Arch)

		ldflags := fmt.Sprintf("-X main.Version=%s -X main.GitCommit=%s -X main.BuildTime=%s",
			version, gitCommit, buildTime)

		buildContainer := baseContainer.
			WithEnvVariable("GOOS", platform.OS).
			WithEnvVariable("GOARCH", platform.Arch).
			WithWorkdir(fmt.Sprintf("/src/%s", project)).
			WithExec([]string{
				"go", "build",
				"-ldflags", ldflags,
				"-o", fmt.Sprintf("/dist/%s", outputName),
				buildTarget,
			})

		// Copy the built binary to our dist container
		binary := buildContainer.File(fmt.Sprintf("/dist/%s", outputName))
		distContainer = distContainer.WithFile(fmt.Sprintf("/dist/%s", outputName), binary)
	}

	return distContainer.Directory("/dist"), nil
}

// PackageHomebrew takes a directory of binaries and packages them as tar.gz archives for Homebrew.
// binariesDir: directory containing the built binaries
// project: the project name (e.g., "tk", "ingest")
func (m *Dagger) PackageHomebrew(ctx context.Context, binariesDir *dagger.Directory, project string) (*dagger.Directory, error) {
	// Use a container to package the binaries
	container := dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/binaries", binariesDir).
		WithWorkdir("/output")

	// Get list of binaries
	entries, err := binariesDir.Entries(ctx)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Skip already packaged files
		if strings.HasSuffix(entry, ".tar.gz") {
			continue
		}

		// Create a tarball for each binary
		container = container.WithExec([]string{
			"sh", "-c",
			fmt.Sprintf("mkdir -p /tmp/%s && cp /binaries/%s /tmp/%s/%s && tar -C /tmp/%s -czf /output/%s.tar.gz %s && rm -rf /tmp/%s",
				project, entry, project, project, project, entry, project, project),
		})
	}

	return container.Directory("/output"), nil
}

// GenerateHomebrewFormula generates a Homebrew formula file for a project.
// project: the project name (e.g., "tk", "ingest")
// version: the version string
// tag: the git tag for the release
// sourceRepo: the GitHub repository (e.g., "neongreen/mono")
// archivesDir: directory containing the packaged tar.gz archives
// desc: project description for the formula
// homepage: project homepage URL
// binaryName: the name of the binary to install
// testArgs: arguments to pass to the binary for testing (comma-separated)
func (m *Dagger) GenerateHomebrewFormula(
	ctx context.Context,
	project string,
	version string,
	tag string,
	sourceRepo string,
	archivesDir *dagger.Directory,
	// +optional
	desc string,
	// +optional
	homepage string,
	// +optional
	binaryName string,
	// +optional
	testArgs string,
) (*dagger.File, error) {
	if desc == "" {
		desc = fmt.Sprintf("%s CLI", project)
	}
	if homepage == "" {
		homepage = fmt.Sprintf("https://github.com/%s", sourceRepo)
	}
	if binaryName == "" {
		binaryName = project
	}
	if testArgs == "" {
		testArgs = "--help"
	}

	// Calculate SHA256 for each archive
	container := dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/archives", archivesDir).
		WithWorkdir("/work")

	// Get list of archives
	entries, err := archivesDir.Entries(ctx)
	if err != nil {
		return nil, err
	}

	type ArchiveInfo struct {
		Platform string
		Filename string
		SHA256   string
	}

	var archives []ArchiveInfo

	for _, entry := range entries {
		if !strings.HasSuffix(entry, ".tar.gz") {
			continue
		}

		// Extract platform from filename
		// Format: project-version-os-arch.tar.gz
		parts := strings.Split(entry, "-")
		if len(parts) < 4 {
			continue
		}

		osName := parts[len(parts)-2]
		archWithExt := parts[len(parts)-1]
		arch := strings.TrimSuffix(archWithExt, ".tar.gz")
		platform := fmt.Sprintf("%s-%s", osName, arch)

		// Calculate SHA256
		content, err := archivesDir.File(entry).Contents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read archive %s: %w", entry, err)
		}
		hash := sha256.Sum256([]byte(content))
		sha256sum := hex.EncodeToString(hash[:])

		archives = append(archives, ArchiveInfo{
			Platform: platform,
			Filename: entry,
			SHA256:   sha256sum,
		})
	}

	// Generate the formula class name
	className := ""
	for _, part := range strings.Split(project, "-") {
		if len(part) > 0 {
			runes := []rune(part)
			runes[0] = unicode.ToUpper(runes[0])
			className += string(runes)
		}
	}

	// Build the formula
	var formulaBuilder strings.Builder
	formulaBuilder.WriteString(fmt.Sprintf("class %s < Formula\n", className))
	formulaBuilder.WriteString(fmt.Sprintf("  desc \"%s\"\n", strings.ReplaceAll(desc, "\"", "\\\"")))
	formulaBuilder.WriteString(fmt.Sprintf("  homepage \"%s\"\n", strings.ReplaceAll(homepage, "\"", "\\\"")))
	formulaBuilder.WriteString(fmt.Sprintf("  version \"%s\"\n", version))

	// Group archives by OS
	macosArchives := make(map[string]ArchiveInfo)
	linuxArchives := make(map[string]ArchiveInfo)

	for _, archive := range archives {
		if strings.HasPrefix(archive.Platform, "darwin-") {
			arch := strings.TrimPrefix(archive.Platform, "darwin-")
			macosArchives[arch] = archive
		} else if strings.HasPrefix(archive.Platform, "linux-") {
			arch := strings.TrimPrefix(archive.Platform, "linux-")
			linuxArchives[arch] = archive
		}
	}

	// Add macOS section
	if len(macosArchives) > 0 {
		formulaBuilder.WriteString("  on_macos do\n")
		for arch, archive := range macosArchives {
			cpuMethod := "arm?"
			if arch == "amd64" {
				cpuMethod = "intel?"
			}
			url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", sourceRepo, tag, archive.Filename)
			formulaBuilder.WriteString(fmt.Sprintf("    if Hardware::CPU.%s\n", cpuMethod))
			formulaBuilder.WriteString(fmt.Sprintf("      url \"%s\"\n", url))
			formulaBuilder.WriteString(fmt.Sprintf("      sha256 \"%s\"\n", archive.SHA256))
			formulaBuilder.WriteString("    end\n")
		}
		formulaBuilder.WriteString("  end\n")
	}

	// Add Linux section
	if len(linuxArchives) > 0 {
		formulaBuilder.WriteString("  on_linux do\n")
		for arch, archive := range linuxArchives {
			cpuMethod := "arm?"
			if arch == "amd64" {
				cpuMethod = "intel?"
			}
			url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", sourceRepo, tag, archive.Filename)
			formulaBuilder.WriteString(fmt.Sprintf("    if Hardware::CPU.%s\n", cpuMethod))
			formulaBuilder.WriteString(fmt.Sprintf("      url \"%s\"\n", url))
			formulaBuilder.WriteString(fmt.Sprintf("      sha256 \"%s\"\n", archive.SHA256))
			formulaBuilder.WriteString("    end\n")
		}
		formulaBuilder.WriteString("  end\n")
	}

	// Add install and test sections
	formulaBuilder.WriteString("  def install\n")
	formulaBuilder.WriteString(fmt.Sprintf("    bin.install \"%s\"\n", binaryName))
	formulaBuilder.WriteString("  end\n")
	formulaBuilder.WriteString("\n")
	formulaBuilder.WriteString("  test do\n")

	// Parse test args
	testArgsList := strings.Split(testArgs, ",")
	testArgsRuby := ""
	for i, arg := range testArgsList {
		arg = strings.TrimSpace(arg)
		if i > 0 {
			testArgsRuby += ", "
		}
		testArgsRuby += fmt.Sprintf("\"%s\"", arg)
	}

	formulaBuilder.WriteString(fmt.Sprintf("    system \"#{bin}/%s\", %s\n", binaryName, testArgsRuby))
	formulaBuilder.WriteString("  end\n")
	formulaBuilder.WriteString("end\n")

	// Return as a file
	return dag.Container().
		From("alpine:latest").
		WithNewFile(fmt.Sprintf("/Formula/%s.rb", project), formulaBuilder.String()).
		File(fmt.Sprintf("/Formula/%s.rb", project)), nil
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
