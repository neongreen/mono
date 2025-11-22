# conf

Smart configuration manager with autocompletion for command-line tools.

## Overview

`conf` is a smart config manager that provides intelligent configuration management with autocomplete for tools like jj (Jujutsu) and mise. It understands tool schemas and provides surgical TOML editing while preserving formatting, comments, and structure.

## Features

- **Smart autocompletion**: Tab completion for all config options based on tool schemas
- **Surgical editing**: Modifies config files without disturbing formatting or comments
- **Multiple format support**: TOML and JSON target config files
- **Schema-aware**: Uses JSON schemas to understand configuration options
- **Import existing configs**: Import configurations from existing tool config files
- **Declarative state management**: Track desired configuration state separately from actual files
- **Drift detection**: Compare desired state with actual config files
- **Multiple tools**: Supports jj, mise, starship, and Claude Code configurations
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
# Apply the hooks to Claude's config.json
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

# Preview what would be imported
conf import --dry-run
```

This is useful for:
- Migrating existing configurations to conf management
- Capturing manual changes made to config files
- Setting up conf on a new machine with existing configs

### State Management

```bash
# View drift between desired and actual state
conf status

# Apply desired state to target files
conf apply

# Apply specific tool only
conf apply jj
```

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
- **claude**: `~/.config/claude/config.json` (JSON format)

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