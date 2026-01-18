# jj-run

A tool to execute shell commands across multiple repository changes in isolated workspaces using [jj](https://github.com/jj-vcs/jj).

- Runs arbitrary shell commands for each change in a revset, in isolation.
- Uses a temporary workspace for each run, so your main repo doesn't change while the script is running.

## Installation

### Standard Installation

Install using Go:

```bash
go install github.com/neongreen/mono/jj-run/cmd@main
```

This installs the `jj-run` binary. To use it as `jj x`, add the following to your Jujutsu config (`~/.config/jj/config.toml`):

```toml
[aliases]
x = ["util", "exec", "--", "jj-run"]
```

### Installation with `want` and `conf`

If you have [`want`](https://github.com/neongreen/mono/tree/main/want) and [`conf`](https://github.com/neongreen/mono/tree/main/conf) installed:

```bash
# Install jj-run
want mono jj-run@main

# Configure the jj alias
conf jj aliases.x '["util", "exec", "--", "jj-run"]'
```

### Installation with Homebrew

(Nov 1, 2025: Not sure if this still works, if it doesn't please yell)

```bash
brew tap neongreen/mono
brew install jj-run
```

## Usage

Simplest form:

```sh
jj x <command>    # run a command on all mutable&reachable changes
```

Full form:

```sh
jj x [-r <revset>] [-e <error_strategy>] [-d] [--repo-untested <url>] <command>
```

- `-r`, `--revset`: The revset of changes to process. If not provided, defaults to `reachable(@, mutable())` (same as `jj fix`).
- `-e`, `--err-strategy`: How to handle errors. One of:
  - `continue` (default): Log errors and continue to next change.
  - `stop`: Stop on the first error, but finish already started changes.
  - `fatal`: Abort immediately on any error.
- `-d`, `--direct`: Enable direct mode (see below).
- `--repo-untested <url>`: Clone and work on a remote repository. The repository will be cloned to a temporary directory, commands will be executed there, and you'll be prompted to push changes when done. **Note: This feature is AI-written and untested.**
- `<command>`: **Required positional argument.** The shell command to execute for each change (runs in the temp workspace, or in main repo in direct mode).

## Direct Mode

Direct mode (`--direct` or `-d`) provides a simpler execution model that doesn't use temporary workspaces:

- For each change in the revset, `jj edit` is run to make that change the working copy
- The command is executed in the main repository's working directory
- No temporary workspaces are created or cleaned up
- No parent rewriting or change squashing is performed

**Use cases for direct mode:**
- Changing commit metadata (descriptions, authors, etc.)
- Running operations that don't require file isolation
- Batch operations that need to work directly with the repository

**Example:**
```sh
# Change the description of all mutable commits
jj x --direct -r 'mutable()' 'jj describe -m "$(jj log -r @ --no-graph -T description) [updated]"'

# Add a co-author to recent commits
jj x --direct -r '::@' 'jj describe -m "$(jj log -r @ --no-graph -T description)\n\nCo-authored-by: Name <email>"'
```

## Readonly Mode

Readonly mode (`--readonly`) runs commands in isolated workspaces without modifying any commits:

- For each change in the revset, a temporary workspace is created
- `jj edit` is used to check out the revision (not `jj new`)
- The command is executed in the workspace
- After execution, the workspace is dropped using `--ignore-working-copy` to prevent snapshotting
- No changes are squashed back to the original commits

**Use cases for readonly mode:**
- Running tests across multiple commits
- Analyzing code at different points in history
- Generating reports or metrics for each commit
- Any read-only operation that shouldn't modify the repository

**Example:**
```sh
# Run tests on all mutable commits
jj x --readonly -r 'mutable()' 'cargo test'

# Check which commits compile successfully
jj x --readonly -r '::@' 'cargo build'

# Generate coverage reports for recent changes
jj x --readonly -r 'recent(5)' 'cargo llvm-cov --html'
```

## Environment Variables

When running commands, jj-run sets the following environment variables with metadata about the current change being processed:

| Variable | Description | Example |
|----------|-------------|---------|
| `JJ_CHANGE_ID` | The change ID (jj's unique identifier) | `plupmpozpuunnxwmxvsussklkvqlsmul` |
| `JJ_COMMIT_ID` | The full commit hash | `162b3b24f7088209c33630ec56deefebc70c2128` |
| `JJ_COMMIT_TIMESTAMP` | ISO 8601 timestamp of the commit | `2026-01-18T02:55:02+01:00` |

These variables are available in all modes (workspace, direct, and readonly).

**Example usage:**
```sh
# Print change info for each commit
jj x --readonly -r 'mutable()' 'echo "$JJ_COMMIT_TIMESTAMP $JJ_CHANGE_ID"'

# Run tests and log results with timestamps (useful when results arrive out of order)
jj x --readonly -j4 -r 'ancestors(@, 20)' '
  echo "=== $JJ_COMMIT_TIMESTAMP $JJ_CHANGE_ID ==="
  cargo test 2>&1 | grep -E "(Passed|Failed|Total)"
'
```

## Working with Remote Repositories

**⚠️ EXPERIMENTAL: This feature is AI-written and untested. Use at your own risk.**

The `--repo-untested` flag allows you to work on remote repositories without manually cloning them first:

- The repository is cloned to a temporary directory
- All commands are executed in that temporary directory
- After processing, you'll be prompted whether to push changes
- The temporary directory is cleaned up automatically

**Use cases:**
- Fix formatting issues in pull requests
- Run automated fixes across repositories
- Test changes on a clean clone
- Apply batch updates to remote branches

**Examples:**
```sh
# Fix formatting in a remote repository
jj-run --repo-untested https://github.com/user/repo 'go fmt ./...'

# Run go mod tidy on all mutable changes
jj-run --repo-untested https://github.com/user/repo 'go mod tidy'

# Work on a specific branch by checking it out first
jj-run --repo-untested https://github.com/user/repo --direct 'jj new main && npm run format'
```

**Note:** When working with private repositories, make sure you have appropriate SSH keys or credentials configured for git operations.

## Limitations

- jj-run can't encapsulate its changes into a single operation, so to undo the changes you will have to use `jj op restore`.
- Doesn't support `--ignore-immutable` yet, so it will fail if the revset contains immutable changes.
- In workspace mode (default): Can't change descriptions of existing commits (it's "for-each-run-and-squash", not "for-each-run").
- In direct mode: Commands that modify the working copy will affect your main repository directly.

## How it works

### Workspace Mode (default)

- For each run, a unique temporary directory is created and a new `jj` workspace is added there.
- The script finds the set of changes matching the revset (excluding the workspace's own change and root).
- For each change:
  1. `jj new <change>` is run in the temp workspace to create a mutable copy.
  2. The provided command is run in the temp workspace.
  3. Output and errors are printed.
- After all changes are processed:
  - The script attempts to rewrite parent snapshots for the new changes.
  - The temp workspace is forgotten and all created changes are abandoned.

### Direct Mode (`--direct`)

- The script finds the set of changes matching the revset.
- For each change:
  1. `jj edit <change>` is run in the main repository to make it the working copy.
  2. The provided command is run in the main repository's working directory.
  3. Output and errors are printed.
- No workspace creation, cleanup, or parent rewriting is performed.

## Error Handling

- If a command fails, the script follows the selected error strategy:
  - `continue`: Logs the error and moves to the next change.
  - `stop`: Stops processing new changes after the first error, but completes any already started ones.
  - `fatal`: Exits immediately on the first error.
- All changes are isolated in the temp workspace. If the script crashes, cleanup is handled per session. The original repository is never modified by failed runs.

## License

MIT

## Development

Build the project:

```bash
mise run //jj-run:build
```

Run tests:

```bash
mise run //jj-run:test
```
