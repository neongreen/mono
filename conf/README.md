# conf

Smart configuration manager with autocompletion for command-line tools.

## Overview

`conf` is a smart config manager that provides intelligent configuration management with autocomplete for tools like jj (Jujutsu) and mise. It understands tool schemas and provides surgical TOML editing while preserving formatting, comments, and structure.

## Features

- **Smart autocompletion**: Tab completion for all config options based on tool schemas
- **Surgical editing**: Modifies config files without disturbing formatting or comments
- **Schema-aware**: Uses JSON schemas to understand configuration options
- **Import existing configs**: Import configurations from existing tool config files
- **Declarative state management**: Track desired configuration state separately from actual files
- **Drift detection**: Compare desired state with actual config files
- **Multiple tools**: Supports jj, mise, and starship configurations
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

# Get current values
conf jj user.name
conf mise settings.experimental
```

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

- **jj (Jujutsu)**: `~/.config/jj/config.toml`
- **mise**: `~/.config/mise/config.toml`
- **starship**: `~/.config/starship.toml`

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