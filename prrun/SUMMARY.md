# prrun - Implementation Summary

## What Was Built

A Go tool called **prrun** (PR Runner) that transparently downloads and runs binaries from GitHub pull request releases. This makes it easy to test changes from pull requests without manual installation.

## How It Works

1. **Input**: User provides a GitHub PR URL and optional project name
   ```bash
   prrun https://github.com/neongreen/mono/pull/123 dissect
   ```

2. **Parse**: Tool extracts owner, repo, and PR number from the URL

3. **Find Release**: Uses GitHub API to find releases matching `pr-N` pattern
   - Looks for tags like `dissect--pr-123.1` or `dissect/pr-123.1`
   - Selects the latest release for that PR

4. **Download**: Downloads the appropriate binary for the current platform
   - Detects OS (linux/darwin) and architecture (amd64/arm64)
   - Caches in `~/.cache/prrun/<release-tag>/<binary-name>`

5. **Execute**: Runs the cached binary with user-provided arguments
   - Forwards all arguments after `--` to the binary
   - Forwards stdin/stdout/stderr for transparent execution

## Key Features

- **Smart Caching**: Downloads only once, subsequent runs use cached binary
- **Platform Detection**: Automatically selects the right binary for your OS/arch
- **Transparent Execution**: Behaves exactly like running the binary directly
- **Multiple Projects**: Supports monorepos with multiple projects
- **Flexible URLs**: Accepts full or short GitHub PR URLs
- **Release Alerts**: Warns when a newer PR release replaces a cached binary

## File Structure

```
prrun/
├── main.go        # Core implementation
├── main_test.go   # Unit tests
├── go.mod         # Go module definition
├── README.md      # User documentation
└── SUMMARY.md     # This file
```

## Implementation Details

### Core Functions

- `parsePRURL()`: Extracts owner, repo, and PR number from GitHub URLs
- `findPRRelease()`: Queries GitHub API to find PR releases
- `getPlatformBinaryName()`: Selects appropriate binary for current platform
- `downloadBinary()`: Downloads and caches binary with proper permissions
- `runBinary()`: Executes cached binary with argument forwarding

### Error Handling

- Invalid URLs return clear error messages
- Missing releases provide helpful feedback
- Platform mismatches are detected early
- Download failures include retry suggestions

### Testing

- Unit tests for URL parsing (multiple formats)
- Cache directory verification
- Platform binary selection
- All tests passing

## Usage Examples

### Basic Usage
```bash
# Test dissect from PR #123
prrun https://github.com/neongreen/mono/pull/123 dissect
```

### With Arguments
```bash
# Run with help flag
prrun https://github.com/neongreen/mono/pull/123 dissect -- --help

# Process a file
prrun https://github.com/neongreen/mono/pull/123 dissect -- myfile.go
```

### Multiple Projects
```bash
# Test different projects from the same PR
prrun github.com/neongreen/mono/pull/123 dissect
prrun github.com/neongreen/mono/pull/123 markdown-format
```

## Cache Structure

Binaries are organized by release tag:

```
~/.cache/prrun/
├── dissect--pr-123.1/
│   └── dissect-pr-123.1-linux-amd64
├── dissect--pr-123.2/
│   └── dissect-pr-123.2-linux-amd64
└── markdown-format--pr-456.1/
    └── markdown-format-pr-456.1-darwin-arm64
```

When a higher `pr-N.X` release is published for a PR you have cached, prrun prints a notice with the previous tag and the new tag before running the binary.

## Integration with Existing Workflow

This tool complements the existing release system:

1. **PR created** → GitHub Actions creates release (`project--pr-N.1`)
2. **prrun used** → Downloads and runs the PR binary
3. **Testing** → User tests the changes locally
4. **PR updated** → New release created (`project--pr-N.2`)
5. **prrun used** → Downloads new version automatically

## Benefits

1. **Fast Testing**: No need to clone, build, or install
2. **Isolation**: Changes are cached separately per release
3. **Reproducibility**: Always uses the exact binary from the PR
4. **Convenience**: Works like Mise or similar version managers
5. **Safety**: Read-only execution from cache

## Future Enhancements (Not Implemented)

Potential improvements that could be added:

- Cache cleanup command (`prrun clean`)
- List cached binaries (`prrun list`)
- Specify cache directory via env variable
- Support for non-PR releases (main branch, tags)
- Verbose mode for debugging

## Supported Platforms

- Linux: amd64, arm64
- macOS: amd64 (Intel), arm64 (Apple Silicon)

Windows support could be added by extending the platform detection.

## Dependencies

- Standard Go library only
- No external dependencies required
- Uses GitHub REST API (no authentication needed for public repos)

## Build and Install

```bash
# Build from source
cd prrun
go build -o prrun .

# Install globally
sudo mv prrun /usr/local/bin/

# Or wait for automatic release
./install.sh prrun  # Once released
```
