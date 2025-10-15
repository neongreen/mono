# Devcontainer Configuration

This devcontainer provides a complete development environment for the monorepo with all required tools pre-installed.

## What's Included

- **Go 1.24.7** - For Go projects (dissect, markdown-format, prrun, printpdf, claude-trace, want)
- **Node.js (LTS)** - For diagram-dsl
- **Rust** - For mdbook-comments
- **mise** - Task runner and tool version manager
- **beads** - Issue tracking tool (installed via mise)
- **GitHub CLI** - For PR workflows

## How to Use

### With VS Code

1. Install the "Dev Containers" extension
2. Open the repository in VS Code
3. Click "Reopen in Container" when prompted (or use Command Palette: "Dev Containers: Reopen in Container")
4. Wait for the container to build and setup to complete

### With GitHub Copilot

The devcontainer is automatically detected by GitHub Copilot agents. When a Copilot agent works on this repository, it will use this container environment.

## Running Projects

All projects use mise for task management. From the repository root:

```bash
# Run a Go project
mise run //prrun:run

# Run tests
mise run //dissect:test

# Build diagram-dsl
cd diagram-dsl && npm run build

# Build mdbook-comments
cd mdbook-comments && cargo build
```

## Manual Setup

If you need to rebuild or update dependencies:

```bash
# Reinstall mise tools
mise install

# Update Go dependencies
cd <project> && go mod download

# Update Node.js dependencies
cd diagram-dsl && npm install

# Rebuild Rust project
cd mdbook-comments && cargo build
```
