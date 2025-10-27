# Monorepo

This repository contains multiple independent projects.

## Projects

| Project | Status | Notes |
| --- | --- | --- |
| [dissect](dissect/) | beta (actively used internally) | Go tool for structural code refactoring; feature set continues to grow. |
| [markdown-format](markdown-format/) | alpha | Markdown formatter; command surface and formatting rules are still evolving. |
| [prrun](prrun/) | deprecated | Was designed to run binaries from PR releases; PR releases are no longer created. |
| [printpdf](printpdf/) | alpha | Markdown/web-to-PDF tool; rendering pipeline has known gaps documented in project issues. |
| [beads-merge](beads-merge/) | alpha | 3-way merge tool for beads `.jsonl` issue files; designed for jj version control. |
| [ingest](ingest/) | pre-alpha | Data ingestion orchestrator; schema and connectors change frequently. |
| [diagram-dsl](diagram-dsl/) | experimental | TypeScript DSL for diagrams; layout system under active refactor. |
| [mdbook-comments](mdbook-comments/) | alpha | mdbook preprocessor for paragraph-level commenting with Supabase backend. |
| [want](want/) | pre-alpha | Planning/fulfilment assistant; core design still in flux. |
| [claude-trace](claude-trace/) | alpha | TUI for reviewing Claude Code conversations; storage format being stabilized. |
| [conf](conf/) | pre-alpha | Smart configuration manager; command coverage incomplete. |
| [tk](tk/) | pre-alpha | System-wide event-sourced task tracker; v0 implements basic claims and authority lattice. |
| [fslog](lib/fslog/) | pre-alpha | ACID-ish filesystem operations library with immutable operation log and rollback support. |
| [ghrelease](lib/ghrelease/) | internal library | Shared helper for fetching release assets; API may change without notice. |

## Installing Tools

### Quick Install

Use the install script to easily download and install the latest version:

```bash
# Install latest version from main
curl -fsSL https://raw.githubusercontent.com/neongreen/mono/main/install.sh | bash -s dissect

# Install specific version
curl -fsSL https://raw.githubusercontent.com/neongreen/mono/main/install.sh | bash -s dissect main.5
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

Main channel releases are considered unstable snapshots unless explicitly tagged. Stable channels are being defined (see bd-313).

Homebrew formulas for the Go CLIs (`ingest`, `want`, `printpdf`, `conf`, `dissect`, `markdown-format`, `prrun`, `claude-trace`, `tk`) are published to [neongreen/homebrew-mono](https://github.com/neongreen/homebrew-mono). Tap it with `brew tap neongreen/mono` and install what you need:

```bash
brew tap neongreen/mono
brew install ingest
brew install want
brew install printpdf
brew install conf
brew install dissect
brew install markdown-format
brew install prrun
brew install claude-trace
brew install tk
```

See [Release Workflow Documentation](.github/workflows/RELEASE_WORKFLOW.md) for more details.

## Testing Changes

To test changes before they're merged:

1. Check out the branch locally
2. Build and test the tool manually

For testing released versions, see the [Releases](#releases) section above.

## Development

Each project has its own development workflow and documentation. See the individual project directories for details.

### Issue Tracking with bd

This repository uses **bd (beads)** for issue tracking. bd is a lightweight, git-based issue tracker designed for AI coding agents.

**Quick Reference:**

```bash
# Check for ready work
bd ready

# Create an issue
bd create "Issue title" -t bug|feature|task -p 0-4

# Update status
bd update <id> --status in_progress

# Close an issue
bd close <id> --reason "Done"
```

Issues are stored in `.beads/issues.jsonl` and tracked in a local SQLite database. See [AGENTS.md](AGENTS.md#bd-beads-issue-tracker) for complete documentation.

### CI/CD

- Each project has its own test workflow
- All Go projects are automatically released via `.github/workflows/release.yml`
- GitHub Copilot coding agent environment is configured via `.github/workflows/copilot-setup-steps.yml`
- See [CI Guidelines](.github/CI_GUIDELINES.md) for workflow structure

## Contributing

See individual project READMEs for contribution guidelines.
