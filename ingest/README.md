# ingest

A CLI tool to ingest various data sources into an SQLite database.

## Overview

`ingest` is a flexible ingestion tool that can capture metadata and content from multiple sources:
- **Git repositories**: All commits, files, and metadata
- **Filesystems**: Files, directories, and their contents
- **Shell commands**: Command output, exit codes, and execution time

All data is stored in an SQLite database located at `~/.ingest/ingest.db`. Each run is treated as separate, allowing you to track multiple ingestion runs.

## Installation

```bash
cd ingest
go build -o ingest ./cmd
```

## Usage

### Ingest a Git repository

```bash
ingest git /path/to/repository
```

This will:
1. Create a new ingestion run in the database
2. Walk through all commits in the repository (with progress indication)
3. Store commit metadata (hash, author, date, message)
4. Store file metadata for each commit (path, size, mode)

### Ingest a filesystem path

```bash
ingest fs /path/to/directory
```

This will:
1. Create a new ingestion run in the database
2. Recursively walk through the directory (with progress indication)
3. Store file and directory metadata (path, size, mode, modification time)
4. Store file contents (for regular files under 10MB)

### Run a shell command and ingest its output

```bash
ingest cmd "your command here"
```

This will:
1. Create a new ingestion run in the database
2. Execute the command
3. Store the command text, exit code, stdout, stderr, and execution time

### List all ingestion runs

```bash
ingest list-runs
```

This displays:
- Run ID
- Run type (git/fs/cmd)
- Start time
- Path or command
- Status (completed/failed/in_progress)
- Number of commits/files ingested
- Duration (for completed runs)

## Database Schema

The database consists of five tables:

### runs
- `id`: Unique run identifier
- `repo_path`: Path to the repository/directory or command
- `run_type`: Type of ingestion (git/fs/cmd)
- `start_time`: When the ingestion started
- `end_time`: When the ingestion finished (NULL if in progress)
- `commit_count`: Number of commits ingested (git only)
- `file_count`: Number of files/entries ingested
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

### fs_entries
- `id`: Unique entry identifier
- `run_id`: Foreign key to runs table
- `path`: Relative path from ingestion root
- `is_dir`: Whether this is a directory
- `size`: File size in bytes
- `mode`: File mode
- `mod_time`: Last modification time
- `content`: File contents (BLOB, for files under 10MB)

### cmd_runs
- `id`: Unique command run identifier
- `run_id`: Foreign key to runs table
- `command`: The shell command that was executed
- `exit_code`: Exit code of the command
- `stdout`: Standard output
- `stderr`: Standard error
- `duration_ms`: Execution time in milliseconds

## Features

- **Additive database**: Running the tool multiple times adds new runs without deleting old data
- **Multiple repositories**: You can ingest different repositories into the same database
- **Separate runs**: Each ingestion is tracked separately, even for the same repository
- **Statistics**: View aggregated statistics across all runs

## Examples

```bash
# Ingest a Git repository
$ ingest git ~/projects/myrepo
Ingesting repository: /home/user/projects/myrepo
Started ingestion run #1
Looking for commits...
Found 1 commits so far...
Found 150 commits total
Processing commit 150/150...
Processed 150 commits with 1234 files
Ingestion completed successfully!

# Ingest a filesystem directory
$ ingest fs ~/Documents
Ingesting filesystem: /home/user/Documents
Started ingestion run #2
Walking filesystem...
Found 100 entries so far...
Found 523 entries total
Processing entry 523/523...
Processed 523 entries
Ingestion completed successfully!

# Run a command and capture its output
$ ingest cmd "ls -la /tmp | wc -l"
Running command: ls -la /tmp | wc -l
Started ingestion run #3
Command completed with exit code: 0 (took 5ms)
Stdout length: 4 bytes
Ingestion completed successfully!

# View all runs
$ ingest list-runs

ID    Type   Start Time          Path/Command                                  Status   Commits    Files     
-----------------------------------------------------------------------------------------------------------------------------
3     cmd    2025-01-06 14:35:20 ls -la /tmp | wc -l                           completed 0          0          (0.0s)
2     fs     2025-01-06 14:32:15 /home/user/Documents                          completed 0          523        (0.8s)
1     git    2025-01-06 14:25:00 /home/user/projects/myrepo                    completed 150        1234       (2.1s)

Summary: 3 total runs (3 completed), 150 commits, 1757 files
```

## Dependencies

- [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver
- [github.com/go-git/go-git/v5](https://github.com/go-git/go-git) - Git library
- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
