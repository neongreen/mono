# ghrelease Library Summary

## Purpose

Shared library for downloading GitHub release assets with platform detection and authentication support. Extracted from `prrun` to be reusable across projects.

## Features

- Platform detection (OS and architecture normalization)
- GitHub authentication via environment variables or gh CLI
- Release fetching by tag name via GitHub API
- Asset selection based on platform
- Binary download with authentication support
- Support for both public and private repositories

## Projects Using This Library

### prrun
Uses ghrelease for:
- Downloading PR release binaries
- Authentication with GitHub API
- Platform-specific asset selection

### want
Uses ghrelease for:
- Downloading GitHub release assets by tag or URL
- Auto-detecting platform for downloads
- Installing binaries to `~/.local/bin/`

## When to Use This Library

Use `ghrelease` when you need to:
- Download assets from GitHub releases using the GitHub API
- Automatically detect the user's platform (OS/arch)
- Support authentication for private repositories
- Find assets by tag name without knowing the exact asset URL

## When NOT to Use This Library

Don't use `ghrelease` if you:
- Already have direct download URLs and don't need API access
- Need to extract archives (tar.gz, zip, etc.) - use a generic downloader
- Are downloading from non-GitHub sources
- Don't need platform detection or authentication

For example, `printpdf` doesn't use this library because it:
- Constructs direct download URLs without needing the API
- Downloads from public releases only
- Needs archive extraction capabilities
- Has its own caching and extraction logic

## API Overview

```go
// Platform detection
platform := ghrelease.GetCurrentPlatform()

// Authentication
token := ghrelease.GetGitHubToken()
req, err := ghrelease.CreateAuthenticatedRequest("GET", apiURL)

// Release and asset fetching
release, err := ghrelease.GetReleaseByTag(owner, repo, tag)
asset, err := ghrelease.FindPlatformAsset(release, projectName)

// Download
err := ghrelease.DownloadAsset(asset, destPath)

// Convenience function
err := ghrelease.DownloadReleaseAsset(owner, repo, tag, projectName, destPath)
```

## Testing

Run tests with:
```bash
cd ghrelease
go test -v
```

All functions are unit tested with mock data.
