# want

Interactive task fulfillment tool for macOS.

## Overview

`want` helps you get things you need on your system through CLI commands. It's an interactive assistant that respects your preferences.

```bash
want jujutsu                # Get a tool
want github.com/user/repo   # Clone a repository
want list                   # Show what you have
```

## How it works

- CLI-first interaction (not YAML editing)
- Presents multiple options when available
- Learns your preferences over time
- Doesn't reset existing work

Configuration is stored at `~/.config/want/`.

See [DESIGN.md](DESIGN.md) for complete design.
