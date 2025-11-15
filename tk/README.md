# tk

tk is my experimental task / personal info tracker for humans and agents.

I want to track thousands of tasks across multiple repositories, chats, emails, feature requests, and hobbies.
I want agents to have access to my data and be able to answer questions like "what promises have I made to people", etc.

*tk* is inspired by [beads](https://github.com/steveyegge/beads), but while beads focuses on helping the agents work better, *tk* focuses on helping my ADHD.
I suppose it helps the agents because they, too, have ADHD.

## This is a personal tool

tk is built for my brain and my usecase of "I want to do everything".
If your are similar to me, you're welcome to use it. If you suggest any ideas, I might try them out.

I try to make *tk* fairly general so that I can try out different things and see what works best.

## Technical Overview

tk is a command-line tool that tracks tasks system-wide using:

- **Event sourcing**: All changes recorded as immutable events in SQLite
- **Offline-first sync**: Sync between machines using immutable segment files (iCloud Drive support)
- **Task relations**: Model dependencies (blocks, subtasks) and hierarchies
- **Multi-valued registers**: Conflicting claims resolved based on role authority
- **No CGO required**: Pure Go SQLite implementation for easy deployment

## Quick Start

```bash
# Create a project
tk project create work "Work tasks"

# Create a task
tk new "Fix authentication bug" --project work

# Mark it in progress
tk mark work-1 in_progress

# List tasks
tk ls
```

## Documentation

- **[Usage Guide](docs/usage.md)** - How to use tk (projects, tasks, relations, sync)
- **[Concepts](docs/concepts.md)** - Architecture and core concepts (events, claims, authority)
- **[Development](docs/development.md)** - Contributing and development guide
- **[Relations](RELATIONS.md)** - Task relations documentation
- **[Specifications](specs/)** - Technical specifications

## Installation

tk can be installed via:
- Homebrew: `brew install tk` (if available)
- Want: `want mono tk@latest`
- From source: `cd tk && go build -o tk .`

### Optional: DuckDB for Query Features

The `tk log query` and `tk log search` commands require DuckDB CLI to be installed:

```bash
# macOS
brew install duckdb

# Linux
wget https://github.com/duckdb/duckdb/releases/latest/download/duckdb_cli-linux-amd64.zip
unzip duckdb_cli-linux-amd64.zip
sudo mv duckdb /usr/local/bin/

# For more installation options: https://duckdb.org/docs/installation/
```

These commands will show a helpful error message if DuckDB is not installed.

## Testing

```bash
go test ./...
```
