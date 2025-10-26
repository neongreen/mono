# want

Interactive task fulfillment tool for macOS.

## ⚠️ Implementation Status

### ✅ Currently Implemented
- [x] **Fulfillment plans** - Shows execution plan before taking action
  - Structured plan with steps showing what will be done
  - Interactive confirmation before executing (unless dry-run or plan-json mode)
  - JSON output option with `--plan-json` flag
  - Highlights manual steps that require user action
  - Skips confirmation for safe/read-only commands
- [x] **Configuration directory** - Creates `~/.config/want/` if missing
- [x] **Download GitHub release assets** - `want https://github.com/owner/repo/releases/tag/v1.0.0`
  - Auto-detects platform (OS and architecture)
  - Works with both public and private repositories (with GITHUB_TOKEN)
  - Downloads to `~/.local/bin/` and makes executable
  - Shows plan and asks for confirmation
- [x] **Install from mono repository** - `want mono <project>@<version>`
  - List all releases for a project: `want mono printpdf --list`
  - Install specific version: `want mono printpdf@main.1`
  - Supports main branch releases
- [x] **Install tools via mise** - Request and install tools using `want <tool>`
  - Shows plan including mise installation if needed
  - Asks for confirmation before installation
- [x] **Install mise itself** - `want mise` installs mise and detects shell config needs
- [x] **Check if tools are already available** - Detects existing installations
- [x] **Shell configuration detection** - Detects if mise activation is configured
- [x] **Dry-run mode** - Preview actions with `--dry-run` flag (no confirmation prompt)
- [x] **JSON plan output** - Output fulfillment plan as JSON with `--plan-json` flag
- [x] **Compound commands with parameters** - Transform and execute commands
  - [x] **`want json <command>`** - Convert command output to JSON using `jc`
  - [x] **`want md <url>`** - Convert URL to markdown using `markitdown` or `pure.md`
- [x] **Basic CLI interface** - Help, version, and command structure
- [x] **Command stubs** - `list`, `check`, `forget` commands (not functional yet)

### ❌ Not Yet Implemented
- [ ] **Clone git repositories** - `want github.com/user/repo` not working yet
- [ ] **Track requirements** - No persistence of what you've requested
- [ ] **List tracked items** - `want list` shows placeholder message
- [ ] **Check status** - `want check` shows placeholder message
- [ ] **Forget requirements** - `want forget` shows placeholder message
- [ ] **Multiple provider support** - Only mise is supported; no Homebrew, GitHub releases, etc.
- [ ] **Interactive prompts** - No interactive selection when multiple options exist
- [ ] **Preference learning** - No storage of user preferences
- [ ] **Configuration persistence** - Configuration data is not stored yet

## Overview

`want` helps you get things you need on your system through CLI commands. It's an interactive assistant that respects your preferences.

```bash
want jujutsu                # Install a tool via mise (asks for confirmation)
want --dry-run jujutsu      # Preview what would be done (no confirmation)
want --plan-json jujutsu    # Output fulfillment plan as JSON
want json ps                # Get process list as JSON (installs jc if needed)
want md https://example.com # Convert URL to markdown
want mono printpdf --list   # List all releases of printpdf from mono
want mono printpdf@main.1   # Install printpdf version main.1 from mono
want https://github.com/org/repo/releases/tag/v1.0.0  # Download GitHub release
want github.com/user/repo   # Clone a repository (NOT YET IMPLEMENTED)
want list                   # Show what you have (NOT YET IMPLEMENTED)
```

## Installation

### With Go

```bash
cd want
go build
```

### With mise

Install using [mise](https://mise.jdx.dev/) with the Go backend:

```bash
mise use -g go:github.com/neongreen/mono/want@main
```

Or add to your `.mise.toml`:

```toml
[tools]
"go:github.com/neongreen/mono/want" = "main"
```

## Usage

### Install from mono repository (✅ WORKING)

The `want mono` command now supports the same fulfillment plan machinery as other `want` commands, including `--dry-run` and `--plan-json` flags.

```bash
# List all available releases and open PRs for a project
want mono printpdf --list

# Install a specific version (asks for confirmation)
want mono printpdf@main.1

# Build from the latest commit on main branch
want mono printpdf@main

# Build from a specific branch
want mono dissect@feature-refactor

# Build from a specific commit SHA
want mono want@abc1234

# Preview what would be done without confirmation
want mono --dry-run printpdf@main.1

# Get the fulfillment plan as JSON
want mono --plan-json printpdf@main.1

# Preview building from a branch (builds locally if no release)
want mono --dry-run dissect@main

# Install from a branch with plan output as JSON
want mono --plan-json dissect@main
```

The tool will:
- Build a fulfillment plan showing all steps before execution
- Ask for confirmation before proceeding (unless using --dry-run or --plan-json)
- Fetch all releases from neongreen/mono repository
- Filter releases for the specified project
- For version strings:
  - If it matches a release tag (e.g., `main.1`): download the pre-built binary
  - If it's a branch name (e.g., `main`, `feature-xyz`): build from the latest commit on that branch
  - If it's a commit SHA (7-40 hex chars): build from that specific commit
- Download the binary matching your platform (OS and architecture)
- Install to `~/.local/bin/`
- Make it executable
- Notify you if `~/.local/bin/` is not in your PATH

Example output for listing releases:

```bash
$ want mono printpdf --list
Fetching releases for printpdf from neongreen/mono...

Available releases for printpdf:
  main.5
  main.4
  main.3

To install a specific version:
  want mono printpdf@<version>

Examples:
  want mono printpdf@main.5
```

Example using dry-run mode:

```bash
$ want mono --dry-run dissect@main
Building dissect from main branch...

Fulfillment plan:

Step 1: Fetch main branch information from GitHub
  $ GET https://api.github.com/repos/neongreen/mono/branches/main

Step 2: Clone neongreen/mono repository (main branch)
  $ git clone --depth=1 --branch main https://github.com/neongreen/mono.git <tmpdir>

Step 3: Build dissect from source
  $ go build -o /home/user/.local/bin/dissect .

Step 4: Make binary executable
  $ chmod +x /home/user/.local/bin/dissect
```

When installing from a branch without a release:

```bash
$ want mono dissect@main
No release found for main (would be tagged as dissect--main)
Building from main branch instead...

Fulfillment plan:

Step 1: Fetch main branch information from GitHub
  $ GET https://api.github.com/repos/neongreen/mono/branches/main

Step 2: Clone neongreen/mono repository (PR branch)
  $ git clone --depth=1 --branch <pr-branch> https://github.com/neongreen/mono.git <tmpdir>

Step 3: Build dissect from source
  $ go build -o /home/user/.local/bin/dissect .

Step 4: Make binary executable
  $ chmod +x /home/user/.local/bin/dissect

Proceed with this plan? [Y/n]: y

PR #100: Add support for new syntax
Branch: feature-branch

Cloning neongreen/mono (branch: feature-branch)...
[clone output]

Building dissect...
[build output]

✓ Built and installed dissect from PR #100 to: ~/.local/bin/dissect

✓ Binary is available in your PATH as: dissect
```

When building from a branch or commit:

```bash
$ want mono printpdf@main
No release found for main (would be tagged as printpdf--main)
Building from latest commit on main branch instead...

Fulfillment plan:

Step 1: Clone neongreen/mono repository (latest commit on main branch)
  $ git clone --depth=1 --branch main https://github.com/neongreen/mono.git <tmpdir>

Step 2: Build printpdf from source
  $ go build -o /home/user/.local/bin/printpdf .

Step 3: Make binary executable
  $ chmod +x /home/user/.local/bin/printpdf

Proceed with this plan? [Y/n]: y

Cloning neongreen/mono (latest commit on main branch)...
[clone output]

Building printpdf...
[build output]

✓ Built and installed printpdf from latest commit on main branch to: ~/.local/bin/printpdf

✓ Binary is available in your PATH as: printpdf
```

### Download GitHub release assets (✅ WORKING)

```bash
# Download from a release tag (auto-detects platform)
want https://github.com/neongreen/mono/releases/tag/want--main.3

# Download specific asset
want https://github.com/neongreen/mono/releases/download/want--main.3/want-main.3-darwin-arm64

# Preview what would be downloaded
want --dry-run https://github.com/neongreen/mono/releases/tag/want--main.3

# Show the fulfillment plan as JSON
want --plan-json https://github.com/neongreen/mono/releases/tag/want--main.3
```

The tool will:
- Show a fulfillment plan with all steps
- Ask for confirmation before proceeding
- Fetch the release information from GitHub
- Find the asset matching your platform (OS and architecture)
- Download it to `~/.local/bin/`
- Make it executable
- Notify you if `~/.local/bin/` is not in your PATH

For private repositories, set `GITHUB_TOKEN` environment variable.

### Install a tool (✅ WORKING)

```bash
want jujutsu                    # Install jujutsu via mise (asks for confirmation)
want mise                       # Install mise itself (special case)
want --plan-json jujutsu        # Show installation plan as JSON
```

The tool will:
- Build a fulfillment plan showing all required steps
- Ask for confirmation before proceeding
- Execute the installation steps

If the tool is already available (installed via brew, apt, or any other method), `want` will detect it and skip installation:

```bash
$ want jq
✓ jq is already available
  Location: /usr/bin/jq
```

Example of interactive installation:

```bash
$ want jujutsu
Fulfillment plan:

Step 1 (MANUAL): Install mise (required for tool installation)
  $ want mise

Step 2: Install jujutsu via mise
  $ mise use -g jujutsu

Proceed with this plan? [Y/n]: y
```

When installing `mise` itself, `want` will:
- Show a fulfillment plan with all steps
- Download and install mise using the official installation script
- Detect if mise activation is needed in your shell config (.bashrc, .zshrc, etc.)
- Provide instructions for adding the activation line

Example JSON output:

```bash
$ want --plan-json mise
{
  "requirement": "mise",
  "steps": [
    {
      "type": "install",
      "description": "Download and run mise installation script",
      "command": "curl https://mise.run | sh",
      "automatic": true
    },
    {
      "type": "configure",
      "description": "Add mise activation to /home/user/.bashrc",
      "command": "echo 'eval \"$(mise activate bash)\"' >> /home/user/.bashrc",
      "automatic": true
    },
    {
      "type": "configure",
      "description": "Activate mise in current shell (or restart shell)",
      "command": "eval \"$(mise activate bash)\"",
      "automatic": false
    }
  ]
}
```

### Fulfillment plans (✅ WORKING)

Every operation now uses a fulfillment plan that shows what will be done. You can:

1. **View the plan interactively** - Default mode shows the plan and asks for confirmation
2. **Preview with dry-run** - Use `--dry-run` to see the plan without executing or prompting
3. **Get JSON output** - Use `--plan-json` to get the plan as structured JSON

```bash
# Interactive mode (shows plan and asks for confirmation)
want jujutsu

# Dry-run mode (shows plan, no execution, no prompt)
want --dry-run jujutsu

# JSON output mode (outputs plan as JSON)
want --plan-json jujutsu
```

Plans highlight manual steps that require user action:
- Manual steps are marked with **(MANUAL)**
- Automatic steps have no label (executed by want)

For safe/read-only commands (like `want json ps`), confirmation is skipped automatically.

### Compound commands with parameters (✅ WORKING)

Transform command output or URLs using specialized tools:

```bash
# Convert command output to JSON
want json ps                    # Get running processes as JSON
want json ls -la                # Get directory listing as JSON
want json df -h                 # Get disk usage as JSON

# Convert URLs to markdown
want md https://example.com     # Convert webpage to markdown
want md https://news.ycombinator.com
```

**How it works:**
- `want json <command>` shows a plan to install `jc` via mise if needed, asks for confirmation, then runs `jc <command>`
- `want md <url>` builds a plan using the best available method to run `markitdown`:
  - Prefers `uvx` (runs in isolated environment without installation)
  - Falls back to `mise` for installation
  - If neither available, suggests installing `uv`
  - Final fallback to `pure.md` web service
- Both commands support `--dry-run` and `--plan-json` flags

### Commands not yet functional (❌ NOT IMPLEMENTED)

```bash
want github.com/user/repo       # Clone a repository - NOT WORKING YET
want list                       # View tracked requirements - NOT WORKING YET
want check                      # Check status of requirements - NOT WORKING YET
want forget <requirement>       # Remove from tracking - NOT WORKING YET
```

## Fulfillment Plan Structure

Every `want` operation builds a fulfillment plan consisting of steps. Each step has:

- **type** - The type of operation: `install`, `download`, `execute`, or `configure`
- **description** - Human-readable description of what the step does
- **command** - The command that will be executed (or a description for complex operations)
- **automatic** - Boolean indicating if the step is automatic (true) or requires manual action (false)

The plan is presented to the user for confirmation before execution (unless using `--dry-run` or `--plan-json`).

Example JSON plan:

```json
{
  "requirement": "jujutsu",
  "steps": [
    {
      "type": "install",
      "description": "Install mise (required for tool installation)",
      "command": "want mise",
      "automatic": false
    },
    {
      "type": "install",
      "description": "Install jujutsu via mise",
      "command": "mise use -g jujutsu",
      "automatic": true
    }
  ]
}
```

## How it works

- CLI-first interaction (not YAML editing)
- Shows fulfillment plans before taking action
- Asks for confirmation before executing (unless dry-run or plan-json mode)
- Presents multiple options when available (future feature)
- Learns your preferences over time (future feature)
- Doesn't reset existing work

Configuration will be stored at `~/.config/want/`.

See [DESIGN.md](DESIGN.md) for complete design.
