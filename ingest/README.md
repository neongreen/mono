# ingest

A CLI tool to ingest various data sources into an SQLite database.

## Overview

`ingest` is a flexible ingestion tool that can capture metadata and content from multiple sources:
- **Git repositories**: All commits, files, and metadata
- **Filesystems**: Files, directories, and their contents
- **Shell commands**: Command output, exit codes, and execution time
- **GitHub repositories**: Issues, pull requests, comments, and all metadata

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

### Ingest GitHub issues and pull requests

```bash
ingest github owner/repo
```

This will:
1. Create a new ingestion run in the database
2. Fetch all issues from the repository (with progress indication)
3. Fetch all pull requests from the repository (with progress indication)
4. Store issue/PR metadata (number, title, body, state, author, dates, labels, assignees, etc.)
5. Store all comments for each issue and pull request

Authentication is handled automatically via:
- `GITHUB_TOKEN` environment variable (recommended for CI/CD)
- `gh` CLI authentication (run `gh auth login` first)

### List all ingestion runs

```bash
ingest list-runs
```

This displays:
- Run ID
- Run type (git/fs/cmd/github)
- Start time
- Path or command
- Status (completed/failed/in_progress)
- Number of items ingested (context-dependent: commits for git, entries for fs, commands for cmd, issues+PRs for github)
- Duration (for completed runs)

### Query the database with SQL

```bash
ingest query "SELECT * FROM runs"
```

This allows you to run arbitrary SQL queries against the ingest database and outputs results as JSON. Examples:

```bash
# Get all runs
ingest query "SELECT id, run_type, status FROM runs"

# Get commits with file counts
ingest query "SELECT c.hash, c.author, COUNT(f.id) as file_count FROM commits c LEFT JOIN files f ON c.id = f.commit_id GROUP BY c.id"

# Find large files
ingest query "SELECT path, size FROM files WHERE size > 1000000 ORDER BY size DESC LIMIT 10"

# Get filesystem entries
ingest query "SELECT path, size FROM fs_entries WHERE is_dir = 0 ORDER BY size DESC LIMIT 5"

# Get all GitHub issues with their comment counts
ingest query "SELECT i.number, i.title, i.state, COUNT(c.id) as comment_count FROM github_issues i LEFT JOIN github_comments c ON c.item_type = 'issue' AND c.item_number = i.number GROUP BY i.id"

# Get all GitHub pull requests with labels
ingest query "SELECT number, title, state, merged, labels FROM github_prs WHERE labels != ''"

# Find most active GitHub issue/PR participants
ingest query "SELECT author, COUNT(*) as comment_count FROM github_comments GROUP BY author ORDER BY comment_count DESC LIMIT 10"
```

## Database Schema

The database consists of eleven tables with efficient blob storage and deduplication:

### runs
- `id`: Unique run identifier
- `repo_path`: Path to the repository/directory, command, or GitHub repository
- `run_type`: Type of ingestion (git/fs/cmd/github)
- `start_time`: When the ingestion started
- `end_time`: When the ingestion finished (NULL if in progress)
- `item_count`: Number of items ingested (commits/entries/commands depending on type)
- `status`: Run status (in_progress/completed/failed)

### blobs
- `id`: Unique blob identifier
- `sha256`: SHA256 hash of the content (unique index for deduplication)
- `content`: The actual file content (BLOB)
- `size`: Size of the content in bytes

**Note**: The `blobs` table provides efficient storage by deduplicating identical file contents across all ingestion types. Multiple files/entries can reference the same blob.

### commits
- `id`: Unique commit identifier
- `run_id`: Foreign key to runs table
- `hash`: Git commit hash
- `author`: Author name
- `author_email`: Author email
- `committer`: Committer name
- `committer_email`: Committer email
- `date`: Commit date
- `message`: Commit message

### commit_parents
- `commit_id`: Foreign key to commits table
- `parent_hash`: Hash of the parent commit

**Note**: This table tracks the commit graph structure, allowing reconstruction of the full commit history.

### files
- `id`: Unique file identifier
- `commit_id`: Foreign key to commits table
- `path`: File path
- `size`: File size in bytes
- `mode`: File mode (permissions)
- `blob_id`: Foreign key to blobs table (NULL if content not available)

### git_refs
- `id`: Unique ref identifier
- `run_id`: Foreign key to runs table
- `ref_type`: Type of reference (branch/tag/other)
- `name`: Reference name (e.g., "main", "v1.0.0")
- `target_hash`: Hash the reference points to

### git_remotes
- `id`: Unique remote identifier
- `run_id`: Foreign key to runs table
- `name`: Remote name (e.g., "origin")
- `url`: Remote URL

### fs_entries
- `id`: Unique entry identifier
- `run_id`: Foreign key to runs table
- `path`: Relative path from ingestion root
- `is_dir`: Whether this is a directory
- `size`: File size in bytes
- `mode`: File mode
- `mod_time`: Last modification time
- `blob_id`: Foreign key to blobs table (NULL for directories or if content not available)

### cmd_runs
- `id`: Unique command run identifier
- `run_id`: Foreign key to runs table
- `command`: The shell command that was executed
- `exit_code`: Exit code of the command
- `stdout`: Standard output
- `stderr`: Standard error
- `duration_ms`: Execution time in milliseconds

### github_issues
- `id`: Unique issue identifier
- `run_id`: Foreign key to runs table
- `number`: Issue number
- `title`: Issue title
- `body`: Issue body/description
- `state`: Issue state (open/closed)
- `author`: Issue author username
- `created_at`: When the issue was created
- `updated_at`: When the issue was last updated
- `closed_at`: When the issue was closed (NULL if open)
- `labels`: Comma-separated list of label names
- `assignees`: Comma-separated list of assignee usernames
- `milestone`: Milestone title (NULL if no milestone)

### github_prs
- `id`: Unique pull request identifier
- `run_id`: Foreign key to runs table
- `number`: Pull request number
- `title`: Pull request title
- `body`: Pull request body/description
- `state`: Pull request state (open/closed)
- `author`: Pull request author username
- `created_at`: When the PR was created
- `updated_at`: When the PR was last updated
- `closed_at`: When the PR was closed (NULL if open)
- `merged_at`: When the PR was merged (NULL if not merged)
- `merged`: Whether the PR was merged (0/1)
- `draft`: Whether the PR is a draft (0/1)
- `base_branch`: Base branch name
- `head_branch`: Head branch name
- `labels`: Comma-separated list of label names
- `assignees`: Comma-separated list of assignee usernames
- `reviewers`: Comma-separated list of requested reviewer usernames
- `milestone`: Milestone title (NULL if no milestone)

### github_comments
- `id`: Unique comment identifier
- `run_id`: Foreign key to runs table
- `item_type`: Type of item (issue/pr)
- `item_number`: Issue or PR number
- `comment_id`: GitHub comment ID
- `author`: Comment author username
- `body`: Comment body
- `created_at`: When the comment was created
- `updated_at`: When the comment was last updated

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
Collecting repository metadata...
Found 2 remotes and 5 refs
Looking for commits...
Found 1 commits so far...
Found 150 commits total
Processing commit 150/150...
Processed 150 commits with 1234 files and 856 blobs
Ingestion completed successfully!

# Ingest a filesystem directory
$ ingest fs ~/Documents
Ingesting filesystem: /home/user/Documents
Started ingestion run #2
Walking filesystem...
Found 100 entries so far...
Found 523 entries total
Processing entry 523/523...
Processed 523 entries with 400 blobs
Ingestion completed successfully!

# Run a command and capture its output
$ ingest cmd "ls -la /tmp | wc -l"
Running command: ls -la /tmp | wc -l
Started ingestion run #3
Command completed with exit code: 0 (took 5ms)
Stdout length: 4 bytes
Ingestion completed successfully!

# Ingest GitHub issues and pull requests
$ ingest github neongreen/mono
Ingesting GitHub repository: neongreen/mono
Started ingestion run #4
Fetching issues...
Found 5 issues so far...
Found 5 issues total
Processing issue 5/5...
Processed 5 issues with their comments
Fetching pull requests...
Found 10 pull requests so far...
Found 42 pull requests total
Processing PR 42/42...
Processed 42 pull requests with their comments
Ingestion completed successfully!

# View all runs
$ ingest list-runs

ID    Type   Start Time          Path/Command                                       Status   Items     
-----------------------------------------------------------------------------------------------------------------------------
4     github 2025-01-06 14:40:30 neongreen/mono                                     completed 47         (15.2s)
3     cmd    2025-01-06 14:35:20 ls -la /tmp | wc -l                                completed 0          (0.0s)
2     fs     2025-01-06 14:32:15 /home/user/Documents                               completed 523        (0.8s)
1     git    2025-01-06 14:25:00 /home/user/projects/myrepo                         completed 150        (2.1s)

Summary: 4 total runs (4 completed), 720 items
```

## Dependencies

- [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver
- [github.com/go-git/go-git/v5](https://github.com/go-git/go-git) - Git library
- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
