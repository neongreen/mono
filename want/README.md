# want

Interactive task fulfillment tool for macOS.

## Overview

`want` helps you get things you need on your system through CLI commands. It's an interactive assistant that respects your preferences.

```bash
want jujutsu                # Get a tool
want github.com/user/repo   # Clone a repository
want list                   # Show what you have
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

Get a tool or resource:
```bash
want jujutsu                    # Get a tool
want github.com/user/repo       # Clone a repository
```

View tracked requirements:
```bash
want list
```

Check status of requirements:
```bash
want check
```

Remove from tracking:
```bash
want forget <requirement>
```

## MVP Status

This is currently a minimal viable product (MVP) implementation. The tool provides the basic command structure but does not yet implement full functionality.

Current MVP features:
- Basic CLI interface
- Command structure (`want`, `list`, `check`, `forget`)
- Help and version commands

Planned features:
- Install tools via mise, Homebrew, or other providers
- Clone git repositories
- Interactive option selection when multiple options exist
- Preference learning and storage
- Configuration persistence at `~/.config/want/`

## How it works

- CLI-first interaction (not YAML editing)
- Presents multiple options when available
- Learns your preferences over time
- Doesn't reset existing work

Configuration will be stored at `~/.config/want/`.

See [DESIGN.md](DESIGN.md) for complete design.
