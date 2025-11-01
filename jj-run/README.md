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
conf jj 'aliases.x' '["util", "exec", "--", "jj-run"]'
```

### Other Installation Methods

#### Quick Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/neongreen/mono/main/install.sh | bash -s jj-run
```

#### Via Homebrew

```bash
brew tap neongreen/mono
brew install jj-run
```

#### Manual Install

1. Go to the [Releases](https://github.com/neongreen/mono/releases) page
2. Find the release you want (e.g., `jj-run--main.1`)
3. Download the binary for your platform
4. Make it executable and move to your PATH:

```bash
chmod +x jj-run
sudo mv jj-run /usr/local/bin/
```

## Usage

Simplest form:

```sh
jj-run <command>    # run a command on all mutable&reachable changes
```

Full form:

```sh
jj-run -r <revset> [-e <error_strategy>] [-d] <command>
```

- `-r`, `--revset`: The revset of changes to process. If not provided, defaults to `reachable(@, mutable())` (same as `jj fix`).
- `-e`, `--err-strategy`: How to handle errors. One of:
  - `continue` (default): Log errors and continue to next change.
  - `stop`: Stop on the first error, but finish already started changes.
  - `fatal`: Abort immediately on any error.
- `-d`, `--direct`: Enable direct mode (see below).
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
jj-run --direct -r 'mutable()' 'jj describe -m "$(jj log -r @ --no-graph -T description) [updated]"'

# Add a co-author to recent commits
jj-run --direct -r '::@' 'jj describe -m "$(jj log -r @ --no-graph -T description)\n\nCo-authored-by: Name <email>"'
```

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
