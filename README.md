# Monorepo

This repository contains multiple independent projects.

## Projects

- **[dissect](dissect/)** - Go tool for code refactoring
- **[markdown-format](markdown-format/)** - Go tool for markdown formatting
- **[diagram-dsl](diagram-dsl/)** - TypeScript DSL for creating diagrams
- **[want](want/)** - Work in progress

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

## Development

Each project has its own development workflow and documentation. See the individual project directories for details.

### CI/CD

- Each project has its own test workflow
- All Go projects are automatically released via `.github/workflows/release.yml`
- See [CI Guidelines](.github/CI_GUIDELINES.md) for workflow structure

## Contributing

See individual project READMEs for contribution guidelines.
