package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"dagger/internal/dagger"
)

type Dagger struct{}

// Platform represents a target platform for cross-compilation
type Platform struct {
	OS   string
	Arch string
}

// Build binaries for all platforms and return them as a directory
// project: the project directory to build (e.g., "tk", "ingest")
// version: the version string (currently unused, kept for API compatibility)
// gitCommit: the git commit hash (currently unused, kept for API compatibility)
func (m *Dagger) BuildRelease(ctx context.Context, project string, version string, gitCommit string) (*dagger.Directory, error) {
	repo := dag.CurrentModule().Source().Directory("..")

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
	outputContainer := dag.Container().From("alpine:latest").WithWorkdir("/output")

	for _, platform := range platforms {
		outputName := fmt.Sprintf("%s-%s-%s-%s", project, version, platform.OS, platform.Arch)

		buildContainer := baseContainer.
			WithEnvVariable("GOOS", platform.OS).
			WithEnvVariable("GOARCH", platform.Arch).
			WithWorkdir(fmt.Sprintf("/src/%s", project)).
			WithExec([]string{
				"go", "build",
				"-o", fmt.Sprintf("/output/%s", outputName),
				buildTarget,
			})

		// Copy the built binary to our output container
		binary := buildContainer.File(fmt.Sprintf("/output/%s", outputName))
		outputContainer = outputContainer.WithFile(fmt.Sprintf("/output/%s", outputName), binary)
	}

	return outputContainer.Directory("/output"), nil
}

// Package binaries as tar.gz archives for Homebrew
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

// Generate a Homebrew formula file for a project
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
	var className strings.Builder
	for part := range strings.SplitSeq(project, "-") {
		if len(part) > 0 {
			runes := []rune(part)
			runes[0] = unicode.ToUpper(runes[0])
			className.WriteString(string(runes))
		}
	}

	// Build the formula
	var formulaBuilder strings.Builder
	formulaBuilder.WriteString(fmt.Sprintf("class %s < Formula\n", className.String()))
	formulaBuilder.WriteString(fmt.Sprintf("  desc \"%s\"\n", strings.ReplaceAll(desc, "\"", "\\\"")))
	formulaBuilder.WriteString(fmt.Sprintf("  homepage \"%s\"\n", strings.ReplaceAll(homepage, "\"", "\\\"")))
	formulaBuilder.WriteString(fmt.Sprintf("  version \"%s\"\n", version))

	// Group archives by OS
	macosArchives := make(map[string]ArchiveInfo)
	linuxArchives := make(map[string]ArchiveInfo)

	for _, archive := range archives {
		if after, ok := strings.CutPrefix(archive.Platform, "darwin-"); ok {
			arch := after
			macosArchives[arch] = archive
		} else if after, ok := strings.CutPrefix(archive.Platform, "linux-"); ok {
			arch := after
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
	var testArgsRuby strings.Builder
	for i, arg := range testArgsList {
		arg = strings.TrimSpace(arg)
		if i > 0 {
			testArgsRuby.WriteString(", ")
		}
		testArgsRuby.WriteString(fmt.Sprintf("\"%s\"", arg))
	}

	formulaBuilder.WriteString(fmt.Sprintf("    system \"#{bin}/%s\", %s\n", binaryName, testArgsRuby.String()))
	formulaBuilder.WriteString("  end\n")
	formulaBuilder.WriteString("end\n")

	// Return as a file
	return dag.Container().
		From("alpine:latest").
		WithNewFile(fmt.Sprintf("/Formula/%s.rb", project), formulaBuilder.String()).
		File(fmt.Sprintf("/Formula/%s.rb", project)), nil
}
