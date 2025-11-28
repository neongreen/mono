# conf

Smart configuration manager with autocompletion for command-line tools.

## Overview

`conf` is a smart config manager that provides intelligent configuration management with autocomplete for tools like jj (Jujutsu) and mise. It understands tool schemas and provides surgical TOML editing while preserving formatting, comments, and structure.

## Architecture & Workflow

### What's Stored Where

```
┌─────────────────────────────────────────────────────────────────────┐
│ LOCAL MACHINE                                                       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ~/.config/conf/                    Target Config Files             │
│  ├── config.toml ──────────────────► ~/.config/jj/config.toml      │
│  │   (tool metadata)                 (actual jj config)             │
│  │                                                                   │
│  ├── jj.toml ──────────────────────► ~/.config/mise/config.toml    │
│  │   (desired jj state)              (actual mise config)           │
│  │                                                                   │
│  ├── mise.toml ────────────────────► ~/.config/starship.toml       │
│  │   (desired mise state)            (actual starship config)       │
│  │                                                                   │
│  ├── claude.toml ──────────────────► ~/.claude/settings.json       │
│  │   (desired claude state)          (actual claude config)         │
│  │                                                                   │
│  └── .sync-state                                                    │
│      (hashes, timestamps)                                           │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ sync (push/pull)
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ REMOTE (iCloud Drive)                                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ~/Library/Mobile Documents/com~apple~CloudDocs/conf/              │
│  ├── jj.toml       (synced desired state)                          │
│  ├── mise.toml     (synced desired state)                          │
│  ├── claude.toml   (synced desired state)                          │
│  └── starship.toml (synced desired state)                          │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Key Concepts

- **Desired State** (`~/.config/conf/<tool>.toml`): What you want the config to be
- **Actual State** (`~/.config/jj/config.toml` etc.): The actual tool config files
- **Remote State** (iCloud): Synced desired state, shared across machines
- **Sync Metadata** (`.sync-state`): File hashes and timestamps for drift detection

### Command Workflows

#### 1. Setting a Value (`conf jj user.name "Alice"`)

```
conf jj user.name "Alice"
         │
         ├─► Read current desired state from ~/.config/conf/jj.toml
         ├─► Update value in desired state
         ├─► Write back to ~/.config/conf/jj.toml
         └─► Surgical edit to ~/.config/jj/config.toml (preserves comments)
```

#### 2. Importing Config (`conf import jj user.name`)

```
conf import jj user.name
         │
         ├─► Read actual value from ~/.config/jj/config.toml
         ├─► Show preview: desired vs. actual
         ├─► Prompt user for confirmation
         └─► Update ~/.config/conf/jj.toml with actual value
```

#### 3. Checking Status (`conf status`)

```
conf status
         │
         ├─► For each tool:
         │   ├─► Read desired state (~/.config/conf/jj.toml)
         │   ├─► Read actual state (~/.config/jj/config.toml)
         │   ├─► Flatten both to dotted paths
         │   └─► Compare values
         │
         └─► Display table with drift:
             ├─► IN SYNC (green)
             ├─► DRIFT (yellow)
             ├─► MISSING (red)
             └─► UNMANAGED (yellow)
```

#### 4. Applying Changes (`conf apply jj`)

```
conf apply jj
         │
         ├─► Read "before" state from ~/.config/jj/config.toml
         ├─► Read desired state from ~/.config/conf/jj.toml
         ├─► Apply all values to ~/.config/jj/config.toml
         ├─► Read "after" state
         ├─► Show unified diff
         └─► Confirm changes
```

#### 5. Syncing with iCloud (`conf sync`)

```
conf sync
         │
         ├─► For each tool:
         │   │
         │   ├─► Read local desired state
         │   │   (~/.config/conf/jj.toml)
         │   │
         │   ├─► Download remote desired state
         │   │   (~/Library/Mobile Documents/.../conf/jj.toml)
         │   │
         │   ├─► Compare modification times (Last-Write-Wins)
         │   │
         │   ├─► Merge configs (newer values win)
         │   │
         │   ├─► Write merged result to:
         │   │   ├─► Local: ~/.config/conf/jj.toml
         │   │   └─► Remote: iCloud Drive
         │   │
         │   └─► Update .sync-state with hashes
         │
         └─► Display sync summary
```

### Data Flow Summary

```
                    ┌──────────────────┐
                    │   conf command   │
                    └────────┬─────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
  ┌──────────┐        ┌───────────┐       ┌──────────┐
  │  import  │        │    set    │       │  apply   │
  └────┬─────┘        └─────┬─────┘       └────┬─────┘
       │                    │                   │
       │ reads actual       │ writes            │ writes
       │                    │                   │
       ▼                    ▼                   ▼
┌──────────────────────────────────────────────────────┐
│           Target Config Files                        │
│  ~/.config/jj/config.toml                            │
│  ~/.config/mise/config.toml                          │
│  ~/.claude/settings.json                             │
└──────────────────────────────────────────────────────┘
       │                    │                   ▲
       │                    │ reads/writes      │
       │                    │                   │
       ▼                    ▼                   │
┌──────────────────────────────────────────────────────┐
│           Desired State (Local)                      │
│  ~/.config/conf/jj.toml                              │
│  ~/.config/conf/mise.toml                            │
│  ~/.config/conf/claude.toml                          │
└────────────────────┬─────────────────────────────────┘
                     │
                     │ sync (push/pull)
                     │
                     ▼
┌──────────────────────────────────────────────────────┐
│           Desired State (Remote)                     │
│  ~/Library/Mobile Documents/.../conf/jj.toml         │
│  ~/Library/Mobile Documents/.../conf/mise.toml       │
└──────────────────────────────────────────────────────┘
```

### Key Design Principles

1. **Separation of Concerns**: Desired state lives in `~/.config/conf/`, actual configs remain in their original locations
2. **Surgical Editing**: Changes preserve comments, formatting, and structure of target files
3. **Drift Detection**: Always know when desired state differs from actual state
4. **Last-Write-Wins**: Sync uses modification time for conflict resolution
5. **Schema-Aware**: Uses JSON schemas for validation and autocompletion

## Features

- **Smart autocompletion**: Tab completion for all config options based on tool schemas
- **Surgical editing**: Modifies config files without disturbing formatting or comments
- **Multiple format support**: TOML and JSON target config files
- **Schema-aware**: Uses JSON schemas to understand configuration options
- **Import existing configs**: Import configurations from existing tool config files
- **Declarative state management**: Track desired configuration state separately from actual files
- **Drift detection**: Compare desired state with actual config files
- **Multiple tools**: Supports jj, mise, starship, and Claude Code configurations
- **Folder tracking**: Track entire folders and sync them across machines
- **iCloud sync**: Sync both tool configs and folders via iCloud Drive
- **Global configs only**: Focuses on user-level configuration files

## Usage

### Basic Configuration Management

```bash
# Set jj configuration
conf jj user.name "Alice"
conf jj user.email "alice@example.com"

# Set mise configuration  
conf mise settings.experimental true
conf mise settings.jobs 4

# Set Claude Code configuration (JSON target)
conf claude model "claude-3-5-sonnet-20241022"
conf claude api.key "sk-ant-..."

# Get current values
conf jj user.name
conf mise settings.experimental
conf claude model
```

### Configuring Claude Code Hooks

Claude Code supports hooks that execute commands at specific points in its lifecycle. Since hooks require complex nested structures (arrays of objects), they should be configured by editing `~/.config/conf/claude.toml` directly, then applying with `conf apply`.

**Example: Configure a Stop hook**

Edit `~/.config/conf/claude.toml`:

```toml
# Top-level settings
model = "claude-3-5-sonnet-20241022"
alwaysThinkingEnabled = true

# Hooks configuration
[hooks]

# Hook that runs when agents finish responding
[[hooks.Stop]]
hooks = [
  { type = "command", command = "echo 'Agent stopped'" }
]

# Hook that runs after Edit or Write tool calls
[[hooks.PostToolUse]]
matcher = "Edit|Write"
hooks = [
  { type = "command", command = "prettier --write", timeout = 5 }
]
```

Then apply the configuration:

```bash
# Apply the hooks to Claude's settings.json
conf apply claude

# Check that hooks are in sync
conf status
```

Available hook types:
- `hooks.Stop` - Runs when agents finish responding
- `hooks.SubagentStop` - Runs when subagents finish responding
- `hooks.PreToolUse` - Runs before tool calls
- `hooks.PostToolUse` - Runs after tool completion
- `hooks.Notification` - Triggers on notifications
- `hooks.UserPromptSubmit` - Runs when a user submits a prompt
- `hooks.SessionStart` - Runs when a new session starts
- `hooks.SessionEnd` - Runs when a session ends
- `hooks.PreCompact` - Runs before the context is compacted

See `conf claude --list` for all available Claude Code configuration options.

### Importing Existing Configurations

Import your existing configurations into conf's state management:

```bash
# Import all tool configurations
conf import

# Import specific tool only
conf import jj

# Import a single setting from a tool
conf import claude model
conf import jj user.name

# Preview what would be imported
conf import --dry-run
conf import jj user.email --dry-run
```

This is useful for:
- Migrating existing configurations to conf management
- Capturing manual changes made to config files
- Setting up conf on a new machine with existing configs
- Selectively importing individual settings without affecting others

### State Management

```bash
# View drift between desired and actual state
conf status

# Apply desired state to target files
conf apply

# Apply specific tool only
conf apply jj
```

### Folder Tracking and Sync

Track entire folders and sync them across machines via iCloud:

```bash
# Track a folder
conf track ~/Documents/my-docs --name my-docs

# Track with exclusion patterns
conf track ~/scripts --name scripts --exclude "*.tmp" --exclude ".DS_Store"

# Check status of tracked folders
conf status

# Sync folders with iCloud
conf sync              # Sync all tools and folders
conf sync my-docs      # Sync specific folder only

# Stop tracking a folder
conf untrack my-docs              # Remove tracking and delete copy
conf untrack my-docs --keep-copy  # Remove tracking but keep copy
```

Folder sync workflow:
1. **Track**: `conf track` copies the folder to `~/.config/conf/<name>/` and creates a manifest
2. **Status**: `conf status` detects drift between source folder and conf copy
3. **Import**: `conf import` pulls changes from source folder to conf copy
4. **Sync**: `conf sync` syncs conf copy with iCloud Drive using Last-Write-Wins
5. **Apply**: `conf apply` pushes changes from conf copy back to source folder

### Shell Completions

```bash
# Generate shell completions
conf --completion bash > /etc/bash_completion.d/conf
conf --completion zsh > ~/.oh-my-zsh/completions/_conf
conf --completion fish > ~/.config/fish/completions/conf.fish
```

## Supported Tools

- **jj (Jujutsu)**: `~/.config/jj/config.toml` (TOML format)
- **mise**: `~/.config/mise/config.toml` (TOML format)
- **starship**: `~/.config/starship.toml` (TOML format)
- **claude**: `~/.claude/settings.json` (JSON format)

## Configuration

conf stores its configuration state in `~/.config/conf/`:

- **Main config**: `~/.config/conf/config.toml` - Tool metadata and settings
- **Per-tool state**: `~/.config/conf/<tool>.toml` - Desired configuration values for each tool (e.g., `jj.toml`, `mise.toml`)

This separation allows conf to:
- Track your desired configuration state
- Detect drift between desired and actual configurations
- Apply configurations across multiple machines
- Import and export configurations easily

## Status

🚧 **Work in Progress** - This tool is currently under development.
