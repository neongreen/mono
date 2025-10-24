# ghrelease

Library for downloading GitHub release assets with platform detection and authentication support.

## Features

- Platform detection (OS and architecture)
- GitHub authentication via environment variables or gh CLI
- Download release assets by tag name
- Support for both public and private repositories
- Automatic pagination to fetch all releases (up to 100 per page)

## Usage

```go
import "github.com/neongreen/mono/lib/ghrelease"

// Download a release asset for the current platform
err := ghrelease.DownloadReleaseAsset(
    "neongreen",           // owner
    "mono",                // repo
    "want--main.3",        // tag
    "want",                // project name (can be empty)
    "/path/to/destination", // destination path
)
```

## API

### Platform Detection

```go
platform := ghrelease.GetCurrentPlatform()
// Returns Platform{OS: "linux", Arch: "amd64"} or similar
```

### Authentication

```go
token := ghrelease.GetGitHubToken()
// Checks GITHUB_TOKEN, MISE_GITHUB_TOKEN, or gh CLI
```

### Finding Assets

```go
release, err := ghrelease.GetReleaseByTag("owner", "repo", "tag")
asset, err := ghrelease.FindPlatformAsset(release, "project-name")
```

### Downloading

```go
err := ghrelease.DownloadAsset(asset, "/path/to/destination")
```

### Listing Releases

```go
releases, err := ghrelease.ListReleases("owner", "repo")
// Returns all releases (automatically paginated)
```

## Authentication

The library checks for GitHub tokens in this order:
1. `GITHUB_TOKEN` environment variable
2. `MISE_GITHUB_TOKEN` environment variable
3. `gh` CLI (via `gh auth token`)

For private repositories, authentication is required.

## Platform Support

Supported platforms:
- Linux: amd64, arm64
- macOS (darwin): amd64, arm64
