# Monorepo

This repository contains multiple independent projects.

## Projects

- **[dissect](dissect/)** - Go tool for code refactoring
- **[markdown-format](markdown-format/)** - Go tool for markdown formatting
- **[prrun](prrun/)** - Go tool to run binaries from GitHub PR releases (for easy testing)
- **[printpdf](printpdf/)** - Go tool for converting Markdown and web content to PDF
- **[diagram-dsl](diagram-dsl/)** - TypeScript DSL for creating diagrams
- **[reactflow-layout-system](reactflow-layout-system/)** - Constraint-based layout system for ReactFlow documents
- **[want](want/)** - Work in progress
- **[claude-trace](claude-trace/)** - Terminal UI for reviewing Claude Code conversation logs
- **[ghrelease](lib/ghrelease/)** - Go library for downloading GitHub release assets

## Installing Tools

### Quick Install

Use the install script to easily download and install the latest version:

```bash
# Install latest version from main
curl -fsSL https://raw.githubusercontent.com/neongreen/mono/main/install.sh | bash -s dissect

# Install specific version
curl -fsSL https://raw.githubusercontent.com/neongreen/mono/main/install.sh | bash -s dissect main.5

# Install from a pull request
curl -fsSL https://raw.githubusercontent.com/neongreen/mono/main/install.sh | bash -s dissect pr-42.1
```

Or download the script and run it locally:

```bash
wget https://raw.githubusercontent.com/neongreen/mono/main/install.sh
chmod +x install.sh
./install.sh dissect main.1
```

### Manual Install

1. Go to the [Releases](https://github.com/neongreen/mono/releases) page
2. Find the release you want (e.g., `dissect--main.1`)
3. Download the binary for your platform
4. Make it executable and move to your PATH:

```bash
# Linux/macOS
chmod +x <binary-name>
sudo mv <binary-name> /usr/local/bin/<tool-name>
```

### Supported Platforms

- Linux: amd64, arm64
- macOS: amd64 (Intel), arm64 (Apple Silicon)

## Releases

Go projects in this repository are automatically released:

- **Main branch releases**: Created on every push to main (e.g., `dissect--main.1`, `dissect--main.2`)
- **PR releases**: Created for pull requests (e.g., `dissect--pr-42.1`) - useful for testing changes before merge

See [Release Workflow Documentation](.github/workflows/RELEASE_WORKFLOW.md) for more details.

## Testing PR Changes

Use **[prrun](prrun/)** to easily test binaries from pull requests without manual installation:

```bash
# Build prrun first (or wait for it to be released)
cd prrun && go build -o prrun .

# Test a PR binary directly (auto-detects project if only one modified)
./prrun https://github.com/neongreen/mono/pull/123 --help

# Or specify the project explicitly
./prrun https://github.com/neongreen/mono/pull/123 -p dissect --help

# Run it on your files
./prrun https://github.com/neongreen/mono/pull/123 -p dissect myfile.go
```

The tool automatically:
1. Detects which project the PR modifies (or lets you specify with `-p`)
2. Finds the PR release
3. Downloads the binary to `~/.cache/prrun/`
4. Runs it with your arguments
5. Warns if the release workflow is pending approval

## Development

Each project has its own development workflow and documentation. See the individual project directories for details.

### CI/CD

- Each project has its own test workflow
- All Go projects are automatically released via `.github/workflows/release.yml`
- See [CI Guidelines](.github/CI_GUIDELINES.md) for workflow structure

## Contributing

See individual project READMEs for contribution guidelines.
