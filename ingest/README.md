# ingest

A CLI tool to ingest Git repository metadata into an SQLite database.

## Overview

`ingest` walks through all commits in a Git repository and stores metadata about commits and files into an SQLite database located at `~/.ingest/ingest.db`. Each run is treated as separate, allowing you to track multiple ingestion runs for the same or different repositories.

## Installation

```bash
cd ingest
go build -o ingest ./cmd
```

## Usage

### Ingest a repository

```bash
ingest ingest /path/to/repository
```

This will:
1. Create a new ingestion run in the database
2. Walk through all commits in the repository
3. Store commit metadata (hash, author, date, message)
4. Store file metadata for each commit (path, size, mode)

### List all ingestion runs

```bash
ingest list-runs
```

This displays:
- Run ID
- Start time
- Repository path
- Status (completed/failed/in_progress)
- Number of commits ingested
- Number of files ingested
- Duration (for completed runs)

## Database Schema

The database consists of three tables:

### runs
- `id`: Unique run identifier
- `repo_path`: Path to the repository
- `start_time`: When the ingestion started
- `end_time`: When the ingestion finished (NULL if in progress)
- `commit_count`: Number of commits ingested
- `file_count`: Number of files ingested
- `status`: Run status (in_progress/completed/failed)

### commits
- `id`: Unique commit identifier
- `run_id`: Foreign key to runs table
- `hash`: Git commit hash
- `author`: Author name
- `author_email`: Author email
- `date`: Commit date
- `message`: Commit message

### files
- `id`: Unique file identifier
- `commit_id`: Foreign key to commits table
- `path`: File path
- `size`: File size in bytes
- `mode`: File mode (permissions)

## Features

- **Additive database**: Running the tool multiple times adds new runs without deleting old data
- **Multiple repositories**: You can ingest different repositories into the same database
- **Separate runs**: Each ingestion is tracked separately, even for the same repository
- **Statistics**: View aggregated statistics across all runs

## Examples

```bash
# Ingest a repository
$ ingest ingest ~/projects/myrepo
Ingesting repository: /home/user/projects/myrepo
Started ingestion run #1
Found 150 commits
Processing commit 150/150...
Processed 150 commits with 1234 files
Ingestion completed successfully!

# Ingest the same repository again (creates a separate run)
$ ingest ingest ~/projects/myrepo
Ingesting repository: /home/user/projects/myrepo
Started ingestion run #2
Found 150 commits
Processing commit 150/150...
Processed 150 commits with 1234 files
Ingestion completed successfully!

# View all runs
$ ingest list-runs

ID    Start Time           Repository                                         Status   Commits    Files     
-----------------------------------------------------------------------------------------------------------------------------
2     2025-01-06 14:30:15  /home/user/projects/myrepo                        completed 150        1234       (2.3s)
1     2025-01-06 14:25:00  /home/user/projects/myrepo                        completed 150        1234       (2.1s)

Summary: 2 total runs (2 completed), 300 commits, 2468 files
```

## Dependencies

- [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver
- [github.com/go-git/go-git/v5](https://github.com/go-git/go-git) - Git library
- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
