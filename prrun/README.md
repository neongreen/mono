# prrun - PR Binary Runner

**⚠️ DEPRECATED**: This tool was designed to download and run binaries from GitHub pull request releases. As of 2025, PR releases are no longer being created in this repository, making this tool obsolete for its original purpose. The tool remains in the repository for historical reference.

---

A tool to transparently download and run binaries from GitHub pull request releases.

## What It Does

`prrun` makes it easy to test binaries from pull requests without manual installation:

1. Give it a GitHub PR URL
2. It finds the corresponding PR release
3. Downloads the binary to a cache folder (`~/.cache/prrun/`)
4. Runs the binary with any arguments you provide

Think of it like a version manager (e.g., Mise) but for testing PR binaries.

## Usage

```bash
# Check version and updates
prrun --version

# Auto-detect project (if PR has only one project)
prrun https://github.com/neongreen/mono/pull/123 --help

# Specify project with arguments (no -- separator needed)
prrun https://github.com/neongreen/mono/pull/123 -p dissect --help

# Run with files as arguments
prrun github.com/neongreen/mono/pull/123 -p markdown-format file.md

# Old syntax still works (with -- separator)
prrun https://github.com/neongreen/mono/pull/123 -p dissect -- --help
```

### Syntax

```
prrun <github-pr-url> [args...]
prrun <github-pr-url> --project <name> [args...]
prrun <github-pr-url> -p <name> [args...]
prrun <github-pr-url> --debug [args...]
prrun --version
```

- `github-pr-url`: Full or short GitHub PR URL (e.g., `https://github.com/owner/repo/pull/123` or `github.com/owner/repo/pull/123`)
- `--project, -p`: Specify project name (required only if PR has multiple projects)
- `--debug`: Show detailed debug information (useful for troubleshooting)
- `--version, -v`: Show version information and check for updates
- `[args...]`: Arguments to pass to the binary (no `--` separator needed anymore!)

The tool automatically detects which project the PR modifies. If multiple projects are detected, you'll need to specify one with `--project` or `-p`.

### Debug Mode

When troubleshooting issues, use the `--debug` flag to see detailed information:

```bash
prrun https://github.com/neongreen/mono/pull/89 --debug
```

This shows:
- How many releases are fetched from the GitHub API
- Which releases match the PR number
- Platform detection and binary selection
- All API calls and their responses

## Installation

### From Source

```bash
cd prrun
go build -o prrun .
sudo mv prrun /usr/local/bin/
```

Note: Version information (commit hash and build time) is automatically embedded via Go's VCS stamping.

### Once Released

```bash
# Install from main branch
./install.sh prrun

# Or use the quick install
curl -fsSL https://raw.githubusercontent.com/neongreen/mono/main/install.sh | bash -s prrun
```

## How It Works

1. **Parse PR URL**: Extracts owner, repo, and PR number from the GitHub URL
2. **Find Release**: Uses GitHub API to find releases tagged with `pr-N` for the given PR
3. **Select Binary**: Finds the binary matching your OS and architecture
4. **Cache**: Downloads to `~/.cache/prrun/<release-tag>/<binary-name>`
5. **Execute**: Runs the cached binary with your arguments
6. **Notify**: Warns when a newer release replaces a cached version for the PR

## Authentication

`prrun` transparently supports authentication for accessing private repositories. It tries multiple authentication methods in order:

1. **GITHUB_TOKEN** environment variable
2. **MISE_GITHUB_TOKEN** environment variable
3. **gh CLI** tool (if authenticated via `gh auth login`)

If any of these are available, `prrun` will use them to authenticate GitHub API requests. This allows you to:

- Access releases from private repositories
- Avoid GitHub API rate limits
- Download private release assets

### Examples

```bash
# Using GITHUB_TOKEN
export GITHUB_TOKEN="ghp_your_token_here"
prrun https://github.com/private-org/private-repo/pull/123 tool

# Using MISE_GITHUB_TOKEN
export MISE_GITHUB_TOKEN="ghp_your_token_here"
prrun https://github.com/private-org/private-repo/pull/123 tool

# Using gh CLI (if already authenticated)
gh auth login
prrun https://github.com/private-org/private-repo/pull/123 tool
```

## Caching

Binaries are cached at `~/.cache/prrun/` organized by release tag:

```
~/.cache/prrun/
├── dissect--pr-123.1/
│   └── dissect-pr-123.1-linux-amd64
└── markdown-format--pr-456.2/
    └── markdown-format-pr-456.2-darwin-arm64
```

Once downloaded, subsequent runs use the cached binary instantly.

When a newer `pr-N.X` release appears for a cached PR, prrun prints a notice with the old tag and the new tag so you know the binary changed before it runs.

## Supported Platforms

- Linux: amd64, arm64
- macOS: amd64 (Intel), arm64 (Apple Silicon)

## Examples

### Test dissect from PR #123

```bash
# Auto-detect if PR only modifies dissect
prrun https://github.com/neongreen/mono/pull/123

# Or explicitly specify the project
prrun https://github.com/neongreen/mono/pull/123 -p dissect
```

### Test markdown-format with a file

```bash
prrun https://github.com/neongreen/mono/pull/456 -p markdown-format README.md
```

### Test without specifying project (auto-detect)

```bash
prrun github.com/someowner/somerepo/pull/789 --help
```

### Multiple projects in PR

If a PR modifies multiple projects, prrun will ask you to specify which one:

```bash
$ prrun https://github.com/neongreen/mono/pull/123 --help
Error: multiple projects found for PR #123:
  - dissect
  - markdown-format

Please specify a project with --project or -p flag:
  prrun https://github.com/neongreen/mono/pull/123 --project <project-name> --help
```

## Requirements

- Go 1.16 or later (for building)
- Internet connection (for downloading binaries)
- GitHub releases must follow the naming convention: `project--pr-N.X` or `project/pr-N.X`

## Use Case

This tool is perfect for maintainers and reviewers who want to quickly test changes from pull requests before merging. Instead of checking out the code, building it, and installing it manually, you can:

```bash
# Test the PR immediately
prrun https://github.com/neongreen/mono/pull/123 -p dissect --help

# Test it on your actual files  
prrun https://github.com/neongreen/mono/pull/123 -p dissect myfile.go
```

## Troubleshooting

For detailed troubleshooting, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

### Quick Fixes

**"no releases found for PR #N"**
- Make sure the PR has a release created (this repo auto-creates releases for PRs)
- Check that the release follows the naming convention
- Use `--debug` flag to see how many releases are being fetched from the API
- If the repository has many releases, the fix for pagination (added 2025-01-13) ensures all releases are searched

**"no binary found for <os>/<arch>"**
- prrun will now show all available assets to help diagnose the issue
- Verify that the release includes binaries for your platform
- Check the release page on GitHub to see available binaries

**"download failed with status 404"**
- prrun will now explain possible causes (no assets, wrong names, auth issues)
- The release may exist but have no assets (build may have failed)
- Check the GitHub Actions logs for build errors
- See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for detailed diagnosis steps

**"failed to download binary"**
- Check your internet connection
- Verify the release URL is accessible
- Make sure you have write permissions to `~/.cache/prrun/`
