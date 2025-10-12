# prrun - PR Binary Runner

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
# Basic usage - run a PR binary
prrun https://github.com/neongreen/mono/pull/123 dissect

# Pass arguments to the binary
prrun https://github.com/neongreen/mono/pull/123 dissect -- --help

# Run with files as arguments
prrun github.com/neongreen/mono/pull/123 markdown-format -- file.md
```

### Syntax

```
prrun <github-pr-url> [project-name] [-- args...]
```

- `github-pr-url`: Full or short GitHub PR URL (e.g., `https://github.com/owner/repo/pull/123` or `github.com/owner/repo/pull/123`)
- `project-name`: Optional. Name of the project if the repo contains multiple projects (e.g., `dissect`, `markdown-format`)
- `-- args...`: Optional. Arguments to pass to the binary (everything after `--` is forwarded)

## Installation

### From Source

```bash
cd prrun
go build -o prrun .
sudo mv prrun /usr/local/bin/
```

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

## Supported Platforms

- Linux: amd64, arm64
- macOS: amd64 (Intel), arm64 (Apple Silicon)

## Examples

### Test dissect from PR #123

```bash
prrun https://github.com/neongreen/mono/pull/123 dissect
```

### Test markdown-format with a file

```bash
prrun https://github.com/neongreen/mono/pull/456 markdown-format -- README.md
```

### Test without specifying project (if repo has single project)

```bash
prrun github.com/someowner/somerepo/pull/789
```

## Requirements

- Go 1.16 or later (for building)
- Internet connection (for downloading binaries)
- GitHub releases must follow the naming convention: `project--pr-N.X` or `project/pr-N.X`

## Use Case

This tool is perfect for maintainers and reviewers who want to quickly test changes from pull requests before merging. Instead of checking out the code, building it, and installing it manually, you can:

```bash
# Test the PR immediately
prrun https://github.com/neongreen/mono/pull/123 dissect -- --help

# Test it on your actual files
prrun https://github.com/neongreen/mono/pull/123 dissect -- myfile.go
```

## Troubleshooting

### "no releases found for PR #N"

- Make sure the PR has a release created (this repo auto-creates releases for PRs)
- Check that the release follows the naming convention

### "no binary found for <os>/<arch>"

- Verify that the release includes binaries for your platform
- Check the release page on GitHub to see available binaries

### "failed to download binary"

- Check your internet connection
- Verify the release URL is accessible
- Make sure you have write permissions to `~/.cache/prrun/`
