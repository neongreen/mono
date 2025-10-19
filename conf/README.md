# conf

Smart configuration manager with autocompletion for command-line tools.

## Overview

`conf` is a smart config manager that provides intelligent configuration management with autocomplete for tools like jj (Jujutsu) and mise. It understands tool schemas and provides surgical TOML editing while preserving formatting, comments, and structure.

## Features

- **Smart autocompletion**: Tab completion for all config options based on tool schemas
- **Surgical editing**: Modifies config files without disturbing formatting or comments
- **Schema-aware**: Uses JSON schemas to understand configuration options
- **Multiple tools**: Supports jj and mise configurations
- **Global configs only**: Focuses on user-level configuration files

## Usage

```bash
# Set jj configuration
conf jj snapshot.max-new-file-size 0
conf jj user.name "Alice"

# Set mise configuration  
conf mise tasks.python.run "python3 {}"

# Generate shell completions
conf --completion bash > /etc/bash_completion.d/conf
```

## Supported Tools

- **jj (Jujutsu)**: `~/.jjconfig.toml`
- **mise**: `~/.config/mise/config.toml`

## Configuration

conf stores its own configuration in `~/.config/conf/config.toml` to track:
- Configured tools and their schema locations
- Tool-specific configuration file paths

## Status

🚧 **Work in Progress** - This tool is currently under development.