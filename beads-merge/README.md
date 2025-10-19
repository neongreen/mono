# beads-merge

A 3-way merge tool for beads `.jsonl` issue tracker files, designed to work with jj (Jujutsu version control).

## Overview

`beads-merge` intelligently merges beads issue tracker files by:

- Matching issues by their unique key (`.id`, `.created_at`, `.created_by`)
- Applying smart merge rules for each field
- Combining dependency arrays and removing duplicates
- Outputting conflict markers for unresolvable conflicts

## Usage

```bash
beads-merge <output-file> <base-file> <left-file> <right-file>
```

The tool reads three versions of a `.jsonl` file, performs a 3-way merge, and writes the result to the output file. If there are conflicts, they are written as conflict markers in the output file and the tool exits with code 1.

### As a jj Merge Tool

Configure in your jj config (e.g., `~/.jjconfig.toml`):

```toml
[merge-tools.beads-merge]
program = "beads-merge"
merge-args = ["$output", "$base", "$left", "$right"]
merge-conflict-exit-codes = [1]
```

The `merge-conflict-exit-codes = [1]` setting tells jj that exit code 1 indicates conflict markers are present in the output file, not that the merge should be aborted.

Then use it with:

```bash
jj resolve --tool=beads-merge
```

## Merge Algorithm

### Issue Matching

Issues are matched by their composite key:
- `.id` - Issue identifier
- `.created_at` - Creation timestamp
- `.created_by` - Creator identifier

### Field Merging Rules

For matched issues, fields are merged as follows:

**String fields** (title, description, notes, status, issue_type):
- If base == left and base != right: take right
- If base == right and base != left: take left
- Otherwise: take left (including when both changed to same value)

**Priority** (integer):
- Same logic as string fields

**Timestamps** (updated_at, closed_at):
- Take the maximum (latest) value
- If one is null, take the non-null value

**Dependencies** (array):
- Combine arrays from both sides
- Remove duplicates based on (issue_id, depends_on_id, type)

### Conflict Handling

Conflicts are output in standard merge conflict format:

```
<<<<<<< left
{"id":"bd-1","title":"Left version",...}
=======
{"id":"bd-1","title":"Right version",...}
>>>>>>> right
```

Conflicts occur when:
- An issue is modified in both branches with incompatible changes to the same field
- An issue is added in both branches with different content
- An issue is modified in one branch and deleted in the other

## Building

```bash
# Run tests
mise run //beads-merge:test

# Build binary
mise run //beads-merge:build

# Format code
mise run //beads-merge:fmt
```

## Examples

### Simple merge

Base:
```jsonl
{"id":"bd-1","title":"Original","status":"open","created_at":"2025-10-16T20:51:29+02:00","created_by":"user1"}
```

Left (updated title):
```jsonl
{"id":"bd-1","title":"Updated","status":"open","created_at":"2025-10-16T20:51:29+02:00","created_by":"user1"}
```

Right (changed status):
```jsonl
{"id":"bd-1","title":"Original","status":"closed","created_at":"2025-10-16T20:51:29+02:00","created_by":"user1"}
```

Result (both changes merged):
```jsonl
{"id":"bd-1","title":"Updated","status":"closed","created_at":"2025-10-16T20:51:29+02:00","created_by":"user1"}
```

### Dependency merge

Left adds dependency on bd-2, right adds dependency on bd-3:

Result combines both dependencies without duplicates.

## Exit Codes

- `0` - Successful merge with no conflicts
- `1` - Conflicts present (conflict markers written to output file)
- Non-zero (other than 1) - Error occurred during processing
