# Release Workflow for Go Projects

This document explains how the automated release system works for Go projects in this monorepo.

## Overview

The `release.yml` workflow automatically builds and releases Go projects from this monorepo. It creates versioned releases for both the main branch and pull requests, allowing you to easily install and test different versions of tools.

## How It Works

### Automatic Detection

The workflow automatically detects Go projects by:
1. Looking for directories with `go.mod` files at the root level
2. Verifying they contain at least one `main.go` file
3. Currently detected: `dissect`, `markdown-format`

### Version Naming

Releases use the format: `<project>/<branch>.<number>`

Examples:
- `dissect/main.1` - First release of dissect from main branch
- `dissect/main.2` - Second release of dissect from main branch
- `dissect/pr-42.1` - First release of dissect from PR #42
- `markdown-format/main.1` - First release of markdown-format from main branch

The version number automatically increments based on existing tags for that project+branch combination.

### When Releases Are Created

- **Main branch**: On every push to main
- **Pull requests**: On PR open, synchronize (new commits), or reopen

### Supported Platforms

Binaries are built for:
- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

## Installing Releases

### Download and Install

Each release includes installation instructions. General pattern:

```bash
# Download the binary for your platform
wget https://github.com/neongreen/mono/releases/download/<tag>/<binary-name>

# Make it executable (Linux/macOS)
chmod +x <binary-name>

# Move to PATH
sudo mv <binary-name> /usr/local/bin/<tool-name>
```

### Installing from Main

To install the latest version from main:

1. Go to the [Releases page](https://github.com/neongreen/mono/releases)
2. Find the latest release for your project (e.g., `dissect/main.X`)
3. Download the appropriate binary for your platform
4. Follow installation instructions in the release notes

### Installing from a Pull Request

To test a version from a specific PR:

1. Go to the [Releases page](https://github.com/neongreen/mono/releases)
2. Find releases tagged with `<project>/pr-<number>.X`
3. Download the appropriate binary
4. These releases are marked as "Pre-release"

## Examples

### Install dissect from main

```bash
# Find the latest release tag (e.g., dissect/main.5)
# For Linux AMD64:
wget https://github.com/neongreen/mono/releases/download/dissect/main.5/dissect-main.5-linux-amd64
chmod +x dissect-main.5-linux-amd64
sudo mv dissect-main.5-linux-amd64 /usr/local/bin/dissect

# For macOS ARM64 (M1/M2):
wget https://github.com/neongreen/mono/releases/download/dissect/main.5/dissect-main.5-darwin-arm64
chmod +x dissect-main.5-darwin-arm64
sudo mv dissect-main.5-darwin-arm64 /usr/local/bin/dissect
```

### Install dissect from PR #123

```bash
# Find the release tag (e.g., dissect/pr-123.1)
wget https://github.com/neongreen/mono/releases/download/dissect/pr-123.1/dissect-pr-123.1-linux-amd64
chmod +x dissect-pr-123.1-linux-amd64
sudo mv dissect-pr-123.1-linux-amd64 /usr/local/bin/dissect
```

## Technical Details

### Build Process

For each detected Go project:
1. Determines if the project has `cmd/main.go` or root-level `main.go`
2. Builds for all supported platforms using cross-compilation
3. Names binaries as: `<project>-<version>-<os>-<arch>[.exe]`

### Release Creation

- Uses GitHub Releases API via `softprops/action-gh-release@v2`
- Attaches all platform binaries
- Includes installation instructions
- Main branch releases are regular releases
- PR releases are marked as pre-releases

### Workflow Jobs

1. **detect**: Finds all Go projects that should be released
2. **release**: Matrix job that builds and releases each project

## Adding New Go Projects

The workflow automatically detects new Go projects. Just ensure your project:
1. Has a `go.mod` file in its directory
2. Has at least one `main.go` file (in root or `cmd/` subdirectory)
3. Can be built with `go build`

## Troubleshooting

### Release not created

- Check that your project has a `go.mod` file
- Verify there's a `main.go` file somewhere in the project
- Check the workflow run logs in GitHub Actions

### Build failed

- Ensure the project builds locally with `go build`
- Check for platform-specific issues (some packages don't support all platforms)
- Review the workflow logs for specific error messages

## Future Improvements

Potential enhancements:
- Support for custom build flags per project
- Integration with `go install` for direct installation
- Checksums for binaries
- Automated changelog generation
