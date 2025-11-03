# Monorepo

This repository contains multiple independent projects.

## Projects

| Project | Status | Notes |
| --- | --- | --- |
| [dissect](dissect/) | beta | Go tool for structural code refactoring; feature set continues to grow. |
| [markdown-format](markdown-format/) | alpha | Markdown formatter; command surface and formatting rules are still evolving. |
| [prrun](prrun/) | deprecated | Was designed to run binaries from PR releases; PR releases are no longer created. |
| [printpdf](printpdf/) | beta | Markdown/web-to-PDF tool; rendering pipeline has known gaps documented in project issues. |
| [beads-merge](beads-merge/) | alpha | 3-way merge tool for beads `.jsonl` issue files; designed for jj version control. |
| [ingest](ingest/) | pre-alpha | Data ingestion orchestrator; schema and connectors change frequently. |
| [diagram-dsl](diagram-dsl/) | pre-alpha | TypeScript DSL for diagrams; layout system under active refactor. |
| [mdbook-comments](mdbook-comments/) | alpha | mdbook preprocessor for paragraph-level commenting with Supabase backend. |
| [want](want/) | alpha | Planning/fulfilment assistant; core design still in flux. |
| [claude-trace](claude-trace/) | pre-alpha | TUI for reviewing Claude Code conversations; storage format being stabilized. |
| [conf](conf/) | alpha | Smart configuration manager; command coverage incomplete. |
| [tk](tk/) | beta | System-wide event-sourced task tracker; v0 implements basic claims and authority lattice. |
| [jj-run](jj-run/) | alpha | Jujutsu subcommand to execute shell commands against multiple revisions. |
| [jj-run-py](jj-run-py/) | deprecated | Old version of jj-run written in Python. |
| [tk-vscode](tk-vscode/) | alpha | VS Code extension that lists tk tasks by running `tk ls --json`. |
| [ghrelease](lib/ghrelease/) | beta | Shared helper for fetching release assets; API may change without notice. |
| [.dagger](.dagger/) | alpha | Dagger module for running tk and dissect test suites. |

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

### Issue Tracking with tk

This repository uses **tk** for issue tracking. tk is an event-sourced task tracker with claims-based status, metadata support, and multi-machine sync.

**Quick Reference:**

```bash
# List tasks
tk ls                        # All tasks
tk ls -p want                # Want tool tasks
tk ls -p mono                # Repo-wide tasks

# Create task
tk new "Fix bug" --project want
tk new "Update CI" --project mono

# Update status
tk mark want-1 wip           # Mark as in progress
tk mark want-1 done          # Complete

# View details
tk show want-1

# Add metadata
tk meta set want-1 priority 1
tk meta set want-1 labels '["bug"]'
```

Tasks are tracked in `~/.tk/tk.db` (event-sourced SQLite database). See [AGENTS.md](AGENTS.md#tk-issue-tracker) for complete documentation.

### CI/CD

- Each project has its own test workflow
- All Go projects are automatically released via `.github/workflows/release.yml`
- See [CI Guidelines](.github/CI_GUIDELINES.md) for workflow structure

## Contributing

See individual project READMEs for contribution guidelines.
