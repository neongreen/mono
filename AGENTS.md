# Agent Guidelines for This Monorepo

This document contains guidelines for AI agents and automated tools working on projects in this monorepo.

## Multi-Agent Environment

**IMPORTANT: Multiple agents may be working on this repository simultaneously.** When making commits, always specify the specific files and subdirectories you're changing to avoid conflicts and provide clear change boundaries.

**Examples:**
- `jj commit conf/ -m "conf: Add schema parsing"` (changes only in conf directory)
- `jj commit lib/ghrelease/ -m "ghrelease: Fix error handling"` (changes only in lib/ghrelease)
- `jj commit .github/workflows/conf.yml -m "conf: Update CI workflow"` (specific file)
- `jj commit README.md AGENTS.md -m "docs: Update project guidelines"` (specific files)

**Do NOT use:**
- `jj commit -m "..."` (commits everything, may include other agents' work)
- `jj commit . -m "..."` (commits current directory, may be too broad)

## Basic rules

- All tools are written in Go unless stated otherwise.
- All new projects are created as top-level folders in the repository unless stated otherwise.
- All projects must contain a `mise.toml`. Check existing `mise.toml` files to see what is expected from you.
- All new Go projects must have CI workflows in `.github/workflows/<project-name>.yml`. Check existing workflow files to see what is expected from you.
- Always create commits with `jj commit -m "commit message"` (use the `-m` flag explicitly).
- In all prose that you write, don't be excited, don't use emojis unless necessary, and don't use pervasive bold text.
- All temporary files (like summaries of fixes you did, one-off scripts you wrote during PR development, etc) must have names prefixed with `ai-temp-`.
- Do not create temporary files in the repository root.
- Keep the list of projects in [README.md](./README.md) up to date.
- Record project status in README.md. If the project is incomplete, has known bugs, exploration, etc, it must be stated.
- Always manage work through bd: create issues when needed, set them to `in_progress` while working, close them as soon as the task is done, and make a Jujutsu commit after closing.
- Release mirroring configuration lives in `release-mirror.toml`. Do not introduce alternative configs for this workflow.
- The release workflow requires the `PUBLIC_RELEASE_TOKEN` secret with permissions to update `neongreen/homebrew-mono`.
- Homebrew tap updates are automated via the release workflow; adjust formula metadata only through `release-mirror.toml`.
- To run releases manually, use the `workflow_dispatch` trigger on `release.yml` and supply a comma/newline separated project list (matches directories like `ingest`).
- Use the provided `mise` tasks instead of `go install`; for example, run `mise run //:lint:actions` when touching workflows to ensure lint and pin checks stay green.
- Never run `npm install -g`; rely on `mise` tasks or `npx` for node tooling.
- Use `yq` for ad-hoc TOML manipulation instead of writing custom scripts or installing global tooling.
- Investigate GitHub Actions failures yourself with available tooling (e.g. `gh run list`, `gh run view --log`); do not defer to the user for log gathering or diagnosis.

------------------------------------------------------------

## Hallucination and bullshitting

Do not try to "sell" the features you are implementing.
Do not describe how convenient or useful the tools are.
If the user hasn't told you why she wants the tool or what she'll be using it for -- don't try to guess; just describe what the tool does.

------------------------------------------------------------

## AGENTS.md

When you are asked to do something "always" or "never", you must also record this rule either in top level AGENTS.md or in project level AGENTS.md, where appropriate.

------------------------------------------------------------

## Pull Request Template

All pull requests must include a code block at the very start showing how to run the PR with the `prrun` tool. This is handled automatically via the pull request template at `.github/pull_request_template.md`. The template should:

- Be at the very start of the PR description
- Show the `prrun` command with placeholders for PR number and project name
- Provide clear examples of how to use it
- Assume the user already has `prrun` installed

------------------------------------------------------------

## Build and Run Guidelines

**Always use `mise` for building and running Go projects. Never use `go build` or `go run` directly.**

### Running Go Projects

Use the mise task syntax from the monorepo root:
```bash
mise run //project-name:task-name
```

Examples:
- `mise run //claude-trace:run` - Run claude-trace with default command (TUI mode)
- `mise run //claude-trace:run list` - Run claude-trace list command
- `mise run //claude-trace:run extract -o output` - Run claude-trace extract command

### Why mise?

- Ensures correct Go version is used
- Manages dependencies consistently
- Provides consistent build environment
- Defined in each project's `mise.toml` file

### Project Tasks

All projects should define standard tasks in their `mise.toml` where applicable:

- **`run`** - Build and run the project (for applications)
- **`test`** - Run all tests

These tasks ensure consistent commands across all projects and make it easy for developers and AI agents to understand how to work with each project.

------------------------------------------------------------

## Code Formatting

**All Go code must be formatted with `go fmt` before work is considered complete.**

Before submitting any changes to Go projects:
- Run `go fmt ./...` in the project directory
- Ensure all Go files are properly formatted
- This applies to both new and modified Go code

The `go fmt` tool ensures consistent formatting across all Go code in the monorepo and is a standard requirement for Go development.

------------------------------------------------------------

## Error Handling Guidelines

**All Go code must follow consistent error handling patterns.**

### Error Wrapping

- **ALWAYS** use `%w` verb when wrapping errors to preserve the error chain
- **NEVER** use `%v` or `%s` for error wrapping as it breaks error unwrapping
- Add context to errors to make debugging easier

**Good examples:**
```go
return fmt.Errorf("failed to fetch release %s/%s tag %s: %w", owner, repo, tag, err)
return fmt.Errorf("failed to create cache directory %s: %w", dirPath, err)
```

**Bad examples:**
```go
return fmt.Errorf("failed to fetch release: %v", err)  // Loses error chain
return fmt.Errorf("error: %s", err)                    // Loses error chain
return fmt.Errorf("GitHub API returned status %d", statusCode)  // No context about what failed
```

### Error Context

Always include relevant context in error messages:
- File paths for file operations
- URLs for HTTP requests
- Resource identifiers (project names, tag names, etc.)
- What operation was being performed

### HTTP Operations

For HTTP operations:
- Use `context.Context` for timeout and cancellation support
- Set reasonable timeouts (30s for API calls, 5min for downloads)
- Include the URL in error messages for debugging
- Include status codes and relevant details

**Example:**
```go
client := &http.Client{Timeout: 30 * time.Second}
resp, err := client.Do(req)
if err != nil {
    return fmt.Errorf("failed to fetch release %s/%s from %s: %w", owner, repo, apiURL, err)
}
if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("GitHub API returned status %d for %s/%s (URL: %s)", resp.StatusCode, owner, repo, apiURL)
}
```

### Context Propagation

- Add context parameters to long-running functions
- Prefer functions with context variants (e.g., `http.NewRequestWithContext`)
- Provide both context and non-context versions for backward compatibility
- Default to `context.Background()` in non-context wrapper functions

------------------------------------------------------------

## Backwards Compatibility Policy

**Unless explicitly stated otherwise, backwards compatibility is NOT important for ANY project in this repository.**

All tools and projects in this monorepo (`diagram-dsl`, `dissect`, `markdown-format`, `want`, and any future projects) are work in progress and do not have users yet. Breaking changes are acceptable and encouraged if they improve the API or implementation.

### What This Means

When making changes to **any project** in this monorepo:
- ❌ **Do NOT** maintain old APIs for compatibility
- ❌ **Do NOT** add migration guides
- ❌ **Do NOT** add "backwards compatible" claims in documentation
- ❌ **Do NOT** worry about breaking changes
- ✅ **DO** focus on making the best possible API and implementation
- ✅ **DO** update documentation to reflect current state only
- ✅ **DO** remove deprecated/old code completely

### Examples from Past PRs

Here are concrete examples from previous pull requests showing what **NOT** to do:

#### ❌ Example 1: Unnecessary Migration Guide (PR #6)

**Project:** diagram-dsl
**File:** `diagram-dsl/IMPROVEMENTS.md`

**What was wrong:**
```markdown
## Migration Guide

No migration needed - all changes are backward compatible.

To use new features:
1. `npm install` to get canvas package
2. Use `renderToSVGWithLayout()` for testing
3. Import `LayoutAssertions` for layout tests
4. Run `npm test` to see new tests in action
```

**Why it's wrong:**
- The library has no users yet
- No one needs migration instructions
- This is just extra documentation that needs to be maintained
- Adds confusion about whether compatibility matters

**What to do instead:**
- Don't add migration sections
- Just document how to use the current version

#### ❌ Example 2: Backwards Compatibility Claims (PR #6)

**Project:** diagram-dsl
**File:** `diagram-dsl/IMPROVEMENTS.md`

**What was wrong:**
```markdown
### Backward Compatibility

All existing APIs remain unchanged:
- `renderToSVG()` still returns a string (uses `renderToSVGWithLayout()` internally)
- All component props remain the same
- All examples work without modifications
```

**Why it's wrong:**
- No users exist to care about API changes
- This adds unnecessary constraints to future development
- Creates false impression that we need to maintain compatibility

**What to do instead:**
- Just document the current API
- If you change APIs, just update the docs

#### ❌ Example 3: Backwards Compatible Claims (PR #8)

**Project:** diagram-dsl
**File:** `diagram-dsl/SEMANTIC_COMPONENTS_SUMMARY.md`

**What was wrong:**
```markdown
## Testing

All existing tests pass (14/14):
- 7 SVG rendering tests
- 7 layout assertion tests

No breaking changes - backwards compatible.
```

**Why it's wrong:**
- Again, no users to worry about
- "backwards compatible" doesn't add value
- Could discourage making better changes in the future

**What to do instead:**
- Just report that tests pass
- Don't make claims about compatibility

#### ❌ Example 4: Defensive Statements About Backwards Compatibility (PR #6)

**Project:** diagram-dsl
**File:** `diagram-dsl/AUDIT.md`

**What was wrong:**
```markdown
### ❌ Not Backward Compatibility Code
The old estimation logic is completely replaced - there's no backward compatibility code lingering.
```

**Why it's wrong:**
- Having to defend that there's no backwards compatibility code suggests we're worried about it
- This shouldn't even be a consideration for work-in-progress projects

**What to do instead:**
- Don't mention backwards compatibility at all
- Just describe what was done

### Owner's Direct Feedback

From the repository owner (PR #8, comment from @neongreen):

> "remove the migration guide and any code left for backwards compatibility, this library is still work in progress and it doesn't have any users right now and no compatibility is needed"

### When Backwards Compatibility DOES Matter

The owner will explicitly request backwards compatibility when needed. Until then, assume it doesn't matter.

### Summary

- **Default assumption:** Backwards compatibility is NOT needed for any project
- **Only consider it when:** The owner explicitly asks for it
- **Focus on:** Making the best possible code, not maintaining old code
- **Applies to:** ALL projects in this monorepo (diagram-dsl, dissect, markdown-format, want, etc.)

------------------------------------------------------------

## Go CI Workflow Configuration

All Go projects in this monorepo must have a `go.sum` file to enable dependency caching in CI workflows.

### Required: go.sum for All Go Projects

**Every Go project must have a `go.sum` file, even if it only has local dependencies.**

- Run `go mod tidy` in each Go project directory to ensure `go.sum` is up-to-date
- If a project has no external dependencies, an empty `go.sum` file is acceptable
- The `go.sum` file must be committed to the repository

### CI Workflow Setup

All CI workflows must specify the `cache-dependency-path` to point to the project's `go.sum` file:

```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.24.7'
    cache-dependency-path: <project-name>/go.sum
```

For example:
- `cache-dependency-path: prrun/go.sum` for the prrun project
- `cache-dependency-path: dissect/go.sum` for the dissect project
- `cache-dependency-path: ${{ matrix.project }}/go.sum` for matrix builds

The action will:
- Look for `go.sum` at the specified path
- Cache Go module dependencies and build cache
- Restore cache on subsequent runs

### Keeping go.sum Up-to-Date

Run `go mod tidy` whenever:
- Adding new dependencies
- Removing dependencies
- Updating Go version
- Changing module requirements

This ensures the `go.sum` file stays synchronized with `go.mod`.

------------------------------------------------------------

## Postmortem Requirements

When a bug or issue is discovered after implementation (especially during code review), agents must create a postmortem analysis documenting:

1. **Timeline**: Chronological sequence of events:
   - What was documented/claimed in the implementation
   - What the reviewer found (the actual bug)
   - What tests were missing

2. **Root Cause**: Why the issue occurred

3. **Prevention Measures**: At least one concrete way this could have been caught earlier

4. **Location**:
   - For project-specific issues: Add to `<project>/AGENTS.md`
   - For cross-cutting concerns: Add to this global `AGENTS.md`

### Example Format

```markdown
### Postmortem: [Brief Title] (YYYY-MM-DD)

**Timeline:**
1. [Initial implementation details]
2. [Review finding]
3. [Missing tests or verification]
4. [Fix applied]

**Root Cause:**
- [Why it happened]

**What Could Have Caught This Earlier:**
1. [Specific action or check]
2. [Another preventive measure]

**Lessons Learned:**
- [Key takeaway 1]
- [Key takeaway 2]
```

### When to Create Postmortems

- When documented functionality doesn't work as claimed
- When tests don't cover documented features
- When assumptions about libraries/APIs are proven wrong
- When edge cases are missed in initial implementation
- When reviewer finds bugs that should have been caught

The goal is continuous improvement: learn from mistakes and build better practices for future work.

------------------------------------------------------------

## bd (Beads) Issue Tracker

This repository uses **bd (beads)** for issue tracking instead of Markdown TODO files or external issue trackers.

### What is bd?

bd is a lightweight, git-based issue tracker designed specifically for AI coding agents. It stores issues in `.beads/issues.jsonl` (committed to git) and maintains a local SQLite database for fast queries.

### Installation

bd is installed globally via:

```bash
# Quick install (recommended)
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/install.sh | bash

# Or via Homebrew
brew tap steveyegge/beads
brew install bd

# Or via go install
go install github.com/steveyegge/beads/cmd/bd@latest
```

### Basic Usage

**Check for ready work:**
```bash
bd ready --json
```

**Create a new issue:**
```bash
bd create "Issue title" -t bug|feature|task -p 0-4 -d "Description" --json
```

**Update issue status:**
```bash
bd update <id> --status in_progress --json
```

**Close an issue:**
```bash
bd close <id> --reason "Done" --json
```

**Show issue details:**
```bash
bd show <id> --json
```

**List all issues:**
```bash
bd list --json
```

**Add dependencies:**
```bash
bd dep add <issue-id> <blocks-id> --type blocks
```

**Show dependency tree:**
```bash
bd dep tree <id>
```

### Issue Types

- `bug` - Something broken that needs fixing
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature composed of multiple issues
- `chore` - Maintenance work (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (nice-to-have features, minor bugs)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Dependency Types

- `blocks` - Hard dependency (issue X blocks issue Y)
- `related` - Soft relationship (issues are connected)
- `parent-child` - Epic/subtask relationship
- `discovered-from` - Track issues discovered during work

Only `blocks` dependencies affect the ready work queue.

### Workflow

1. **At session start**: Run `bd ready` to see what's unblocked
2. **Claim a task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work**: Create issues for bugs/TODOs found during work
5. **Complete**: `bd close <id> --reason "Implemented"`
6. **Auto-sync**: Changes automatically export to `.beads/issues.jsonl` after 5 seconds

### Agent Guidelines

- **Always use `--json` flag** for programmatic use
- **Use bd instead of Markdown** for all new work tracking
- **Link discovered issues** using `discovered-from` dependency type
- **Check `bd ready`** before asking "what should I work on next?"
- **Auto-sync is enabled**: JSONL is automatically updated after CRUD operations
- **Issues are git-versioned**: The `.beads/issues.jsonl` file is the source of truth
- **SQLite DB is local**: The `*.db` files are in `.gitignore` and regenerated from JSONL

### Git Workflow

bd automatically handles git synchronization:

- **Export**: After any CRUD operation, changes are exported to `.beads/issues.jsonl` (5-second debounce)
- **Import**: When JSONL is newer than DB (e.g., after `git pull`), it's automatically imported

```bash
# Make changes
bd create "Fix bug" -p 1
bd update mono-42 --status in_progress

# Wait 5 seconds for auto-export, or run manually
bd export

# Commit
git add .beads/issues.jsonl
git commit -m "Your message"

# After pull, BD auto-imports the updated JSONL
git pull
bd ready  # Fresh data from git
```

### Repository Setup

This repository has been initialized with:
- Database at `.beads/mono.db` (not committed)
- Issue prefix: `mono` (issues are named `mono-1`, `mono-2`, etc.)
- JSONL export at `.beads/issues.jsonl` (committed to git)

### Resources

- [bd GitHub Repository](https://github.com/steveyegge/beads)
- [bd Documentation](https://github.com/steveyegge/beads/blob/main/README.md)
- [bd Workflow Guide](https://github.com/steveyegge/beads/blob/main/WORKFLOW.md)
- [bd for Agents](https://github.com/steveyegge/beads/blob/main/AGENTS.md)

### Merge Tool Configuration

This repository includes `beads-merge`, a custom 3-way merge tool for `.beads/issues.jsonl` files. To use it with jj:

1. Build the tool: `cd beads-merge && go build -o beads-merge .`
2. Add to your jj config (`~/.jjconfig.toml`):

```toml
[merge-tools.beads-merge]
program = "/absolute/path/to/mono/beads-merge/beads-merge"
merge-args = ["$output", "$base", "$left", "$right"]
merge-conflict-exit-codes = [1]

[merge-tools.beads-merge.diff-args]
# Optional: configure for 2-way diff if needed
program = "diff"
```

3. Configure automatic merge for .jsonl files in `.beads/`:

```toml
[merge]
# Use beads-merge for .beads/issues.jsonl
tool-edits = [
  { pattern = ".beads/issues.jsonl", tool = "beads-merge" }
]
```

The `merge-conflict-exit-codes = [1]` setting tells jj that exit code 1 indicates conflict markers are present in the output file, not that the merge should be aborted.

The merge tool will:
- Match issues by id, created_at, and created_by
- Intelligently merge field changes
- Combine dependency arrays
- Write conflict markers to the output file for unresolvable conflicts

See [beads-merge/README.md](./beads-merge/README.md) for details on the merge algorithm.

------------------------------------------------------------

## Go Workspace Management

This monorepo uses a Go workspace (`go.work`) to manage multiple modules. Follow these rules to avoid dependency resolution errors.

### Rules

1. **Never run `go mod tidy` in individual module directories.** Always use `mise run //:go:tidy-all` from the workspace root, which runs `go work sync`.

2. **Use `v0.0.0` for all workspace-local module dependencies.** All modules in this workspace should depend on other workspace modules using exactly `v0.0.0` (not pseudo-versions like `v0.0.0-20251022141859-f6ab99927bb0`).

3. **Replace directives must match actual module versions.** The `go.work` file must contain replace directives that exactly match the versions required by modules:
   ```go
   replace (
       github.com/neongreen/mono/lib/ghclient v0.0.0 => ./lib/ghclient
       github.com/neongreen/mono/lib/ghrelease v0.0.0 => ./lib/ghrelease
       github.com/neongreen/mono/lib/toml v0.0.0 => ./lib/toml
   )
   ```

4. **Use `go work sync` to synchronize versions.** This command updates all workspace modules to use consistent versions from the workspace build list. This is what `mise run //:go:tidy-all` does.

5. **Run the workspace linter before committing.** Use `mise run //:lint:go-workspace` to validate that replace directives match module requirements.

6. **CI validation.** The `go-workspace-lint.yml` workflow runs on every PR to ensure:
   - No individual modules have local replace directives (they should all be in `go.work`)
   - Workspace replace directives match module requirements
   - `go work sync` succeeds without modifications

### Why This Matters

Running `go mod tidy` in individual module directories causes Go to attempt downloading dependencies from the remote repository, even with workspace replace directives in place. Since our local module versions (like `v0.0.0`) don't exist as git tags remotely, this results in errors like:

```
unknown revision lib/ghclient/v0.0.0
```

The solution is to use `go work sync` from the workspace root, which handles all dependency resolution for workspace modules without attempting remote downloads. Individual module tidying is not necessary in a workspace setup.

### Quick Reference

```bash
# Synchronize all workspace modules
mise run //:go:tidy-all

# Or run go work sync directly
export GOPRIVATE=github.com/neongreen/mono
export GONOSUMDB=github.com/neongreen/mono
go work sync

# Validate workspace consistency
mise run //:lint:go-workspace
```

### Adding New Workspace Modules

When adding a new module to the workspace:

1. Add it to the `use` directive in `go.work`
2. Use `v0.0.0` for any dependencies on other workspace modules
3. Add corresponding `replace` directives in `go.work` for the `v0.0.0` versions
4. Run `mise run //:go:tidy-all` to synchronize everything
5. Verify with `mise run //:lint:go-workspace`

------------------------------------------------------------

## Avoid String Rendering for Internal APIs

**Never convert structured data to strings for internal APIs.** String rendering (like dotted paths, JSON, etc.) should only be used for:
- Human-readable output/display
- Accepting human input
- External API contracts
- Serialization for storage/transmission

### The Problem

Converting structured data to strings and then parsing them back is poor engineering:

```go
// ❌ BAD: Flattening, then parsing
values := map[string]interface{}{"aliases": map[string]interface{}{".": ["status"]}}
flatValues := FlattenValues(values)  // Creates "aliases.\".\"" 
for path, value := range flatValues {
    tool.SetConfig(path, value)  // Has to parse "aliases.\".\""
}

// ✅ GOOD: Work with native structures
values := map[string]interface{}{"aliases": map[string]interface{}{".": ["status"]}}
tool.SetAllValues(values)  // No parsing needed
```

### Why This Matters

1. **Parsing complexity**: Special characters (dots, quotes, etc.) require escaping and complex parsing logic
2. **Performance**: Converting to strings and back is wasteful
3. **Error-prone**: Easy to introduce bugs in escaping/unescaping
4. **Fragile**: Changes to rendering format break everything

### Example: conf apply

The `conf apply` command was originally implemented incorrectly:

```go
// ❌ Original (bad) implementation:
func applyTool(conf *config.Config, toolName string) error {
    flatValues := config.FlattenValues(tool.Values)  // Convert to strings
    for path, value := range flatValues {
        ApplyToolValue(toolName, path, value)  // Parse strings back
    }
}
```

This required:
- `FlattenValues` to quote special keys: `aliases."."` 
- `ValidatePath` to parse quoted keys with `parser.ParseKey`
- Extra complexity throughout the codebase

The correct implementation passes structured data directly:

```go
// ✅ Correct implementation:
func applyTool(conf *config.Config, toolName string) error {
    tools.ApplyAllToolValues(toolName, tool.Values)  // Pass native map
}
```

### Guidelines

1. **Keep structured data as structured data** until you need to serialize it
2. **Add bulk/batch methods** (like `SetAllValues`) instead of flattening
3. **Only flatten for display** or when a human needs to provide individual values
4. **Use parser.Key or similar types** internally if you need to represent paths programmatically

### When String Rendering Is Acceptable

- CLI commands that accept user input: `conf jj set user.name "John"`
- Display output: showing config drift, status, etc.
- External APIs: REST endpoints, file formats, etc.
- Configuration file serialization: TOML, JSON, YAML

**The key principle**: If you control both ends (caller and callee), use structured data. Only render to strings at the boundaries.

------------------------------------------------------------

## Database Schema Initialization

**The entire database schema must be created by `InitDB()`.** This includes all tables, indexes, and other database objects required by the application.

### Why This Matters

1. **Single source of truth**: All schema definitions live in one place (`InitDB()`)
2. **Testability**: Tests can create fresh databases without worrying about missing tables
3. **Deployment**: New deployments create complete schemas automatically
4. **Consistency**: No risk of partial schemas from incremental creation

### Implementation

The `InitDB()` function must create ALL tables used by the application, including:
- Event storage tables (`events`, `event_id_map`)
- Metadata tables (`metadata`)
- Counter tables (`task_counter`, `event_counter`)
- Projection tables (`projects`, `project_aliases`, `tasks`, `task_numbers`)
- All indexes on these tables

**Do NOT** create tables lazily in projection functions or other code paths. Tables must exist before any code tries to use them.

### Example

```go
func (d *DB) InitDB() error {
    schema := `
        CREATE TABLE IF NOT EXISTS events (...);
        CREATE TABLE IF NOT EXISTS projects (...);
        CREATE TABLE IF NOT EXISTS tasks (...);
        -- ... all other tables ...
    `
    if _, err := d.db.Exec(schema); err != nil {
        return fmt.Errorf("failed to create schema: %w", err)
    }
    return nil
}
```
