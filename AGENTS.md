# Agent Guidelines for This Monorepo

This document contains guidelines for AI agents and automated tools working on projects in this monorepo.

## Dagger Best Practices

When working with Dagger modules (`.dagger/` directory):
- **Always check the official Dagger repository for best practices**: https://github.com/dagger/dagger
- Look at their own Dagger modules for reference implementations (e.g., `modules/go/main.go` for Go project handling)
- Don't reinvent the wheel - see how the Dagger team solves common problems like:
  - Configuration file handling (golangci-lint automatically searches parent directories)
  - Caching strategies (module cache, build cache, tool caches)
  - Parallel execution with limits
  - Directory mounting and workdir patterns

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
- All new Go projects must have CI workflows in `.github/workflows/<project-name>.yml`. Check existing workflow files to see what is expected from you.
- All project tasks are defined in the top-level `mise.toml` file with project-name prefixes (e.g., `project-name:task-name`).
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

## Build and Run Guidelines for Go Projects

**For Go projects, agents should use `go` commands directly. Do not use `mise` for building, testing, or running Go code.**

### Running Go Commands

All Go commands should be run from the repository root:

```bash
# Build a project
go build ./project-name

# Run tests for a project
go test ./project-name/...

# Run a project
go run ./project-name

# Build and install
go install ./project-name
```

### Why Direct Go Commands?

- Go's toolchain is reliable and self-contained
- Simpler for agents to execute directly
- No additional tool installation required
- Each project's AGENTS.md provides specific commands

### Project-Specific Commands

Each project directory has an AGENTS.md file with specific build, test, and run commands for that project. These commands use `go` directly and should be run from the repository root.

For code formatting, use `goimports -w .` from the repository root to format all Go code in the monorepo.

------------------------------------------------------------

## Go Module Download Issues

**If `go test` or `go build` fails with DNS/network errors when downloading modules, try bypassing the Go proxy.**

### Symptom

When running tests or builds, you see errors like:
```
go: downloading modernc.org/sqlite v1.39.1
dial tcp: lookup storage.googleapis.com on [::1]:53: read udp [::1]:53: connection refused
```

Some modules download successfully while others fail - this indicates DNS resolution issues for the Go module proxy (`proxy.golang.org`).

### Solution

Use direct downloads instead of the proxy:

```bash
# Run tests with direct module downloads (no proxy)
GOPROXY=direct go test ./...

# Build with direct module downloads (no proxy)
GOPROXY=direct go build ./...
```

### Why This Works

- Go's default `GOPROXY` is `https://proxy.golang.org,direct`
- The proxy (`proxy.golang.org`) resolves to `storage.googleapis.com` which may have DNS issues in some environments
- Setting `GOPROXY=direct` skips the proxy and fetches modules directly from their source repositories (GitHub, etc.)
- Direct downloads work better in environments with restricted DNS or network configurations

### When to Use

- Claude Code web environments with DNS restrictions
- Containerized environments with network isolation
- CI environments with proxy issues
- Any environment where the Go module proxy is unreachable

### Permanent Configuration

To set this permanently for your shell session:
```bash
export GOPROXY=direct
```

Or configure it in your Go environment:
```bash
go env -w GOPROXY=direct
```

------------------------------------------------------------

## Code Formatting

**All Go code must be formatted before work is considered complete.**

Before submitting any changes to Go projects:
- Run `mise run format` from the repository root
- Applies to both new and modified Go code

The centralized `format` task ensures consistent formatting across all Go code in the monorepo.

------------------------------------------------------------

## Table Display Library

**All Go code that displays ASCII-style tables must use `github.com/jedib0t/go-pretty/v6/table`.**

The standard table configuration should use:
- `table.NewWriter()` to create a table
- `t.SetOutputMirror(os.Stdout)` to output directly to stdout
- `t.SetStyle(table.StyleLight)` for consistent styling
- `t.Style().Options.SeparateRows = true` to separate rows
- `t.Style().Options.DrawBorder = false` to hide outer borders
- `t.Render()` to render the table

Example:
```go
import (
    "os"
    "github.com/jedib0t/go-pretty/v6/table"
)

t := table.NewWriter()
t.SetOutputMirror(os.Stdout)
t.AppendHeader(table.Row{"ID", "Status", "Title"})

t.SetStyle(table.StyleLight)
t.Style().Options.SeparateRows = true
t.Style().Options.DrawBorder = false

t.AppendRow(table.Row{"1", "done", "Example task"})
t.Render()
```

See `tk/display.go` (renderTaskTable function) for a reference implementation.

------------------------------------------------------------

## Go Style Guide

**All Go code must follow the standards defined in [GO_STYLE_GUIDE.md](./GO_STYLE_GUIDE.md).**

The style guide covers:
- CLI framework usage (Cobra)
- Error handling patterns
- Project structure
- Code organization
- Testing patterns
- Documentation conventions

Key standards:
- Use Cobra for all CLI applications
- Use `RunE` for command handlers (return errors, main handles exit codes)
- CLI applications have `main.go` in project root (not in cmd/ subdirectory)
- The cmd/ subdirectory is optional and used for Cobra command logic (not package main)
- Use `fmt.Fprintf(os.Stderr, "Error: %v\n", err)` for error output in main
- See `GO_STYLE_GUIDE.md` for complete details and examples

------------------------------------------------------------

## Testing Guidelines

**Do not write useless tests. See [BULLSHIT_TESTS.md](./BULLSHIT_TESTS.md) for examples of tests to avoid.**

Avoid tests that:
- Verify that constructors return non-nil
- Check that struct fields were assigned the values passed to constructors
- Test that basic language features work (field assignment, struct creation)
- Verify that framework registration happened (Cobra commands, HTTP routes)
- Test library functionality instead of your application logic

Write tests that:
- Verify business logic and algorithms
- Test error conditions and edge cases
- Validate data transformations
- Check integration between components
- Exercise failure modes and recovery

See [BULLSHIT_TESTS.md](./BULLSHIT_TESTS.md) for detailed examples of what was removed and why.

------------------------------------------------------------

## Version Subcommand Pattern

**All Go CLI tools must implement a `version` subcommand using the shared `lib/version` package.**

### Implementation

1. Import the version package in your cmd directory:
   ```go
   import "github.com/neongreen/mono/lib/version"
   ```

2. Add the version command in an `init()` function in a `version.go` file:
   ```go
   func init() {
       rootCmd.AddCommand(version.NewVersionCommand("tool-name"))
   }
   ```

3. The version command automatically provides:
   - Human-readable output: `tool-name version dev`
   - JSON output with `--json` flag
   - Version, git commit, build time, and Go version information

### Build-time Version Information

Version information is automatically embedded by Go's VCS stamping (available since Go 1.18). When building with `go build`, Go automatically includes:
- Git commit hash
- Build time from VCS
- Whether the working tree was modified

This works automatically when:
- Building from a git repository
- Building a package directory (e.g., `.`) rather than a specific file (e.g., `./main.go`)
- Not in a detached HEAD state (or using go build from the module root)

The `lib/version` package reads this information from `debug.ReadBuildInfo()` at runtime.

### Example Usage

```bash
# Human-readable output
$ conf version
conf version main.42
  commit: a1b2c3d
  built:  Jan 15, 2024 10:30 PST
  go:     go1.24.7

# JSON output
$ conf version --json
{
  "build_time": "2024-01-15T18:30:00Z",
  "commit": "a1b2c3d",
  "go_version": "go1.24.7",
  "version": "main.42"
}
```

### Tools with Version Commands

All Go CLI tools in this monorepo have version subcommands:
- tk
- want
- conf
- dissect
- ingest
- printpdf
- claude-trace
- jj-run

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

## Logging Guidelines

**All Go code must follow consistent logging patterns to avoid duplicate and unclear log messages.**

### Core Principle: Log at the Boundaries, Not in Utilities

**Callees (utility functions, library code) should NOT log. Callers (business logic, command handlers) should log.**

### Why This Matters

1. **Avoid duplicates**: If both caller and callee log, you get redundant messages
2. **Context**: Only the caller knows the business context worth logging
3. **Reusability**: Utility functions shouldn't pollute logs when used in different contexts
4. **Separation of concerns**: Low-level functions do work, high-level functions coordinate and log

### The Pattern

```go
// ❌ BAD: Utility function logs
func LoadIndexFile(path string) (*IndexFile, error) {
    slog.Debug("loading index file", "path", path)  // Don't do this
    data, err := os.ReadFile(path)
    // ...
}

// ✅ GOOD: Utility function is silent
func LoadIndexFile(path string) (*IndexFile, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read index: %w", err)  // Just return errors
    }
    // ...
}

// ✅ GOOD: Business logic logs
func Push(...) error {
    log := slog.With("remote", remoteName, "space", space)
    log.Debug("push: loading local index", "path", localIndexPath)
    localIndex, err := LoadIndexFile(localIndexPath)  // Silent utility
    if err != nil {
        log.Error("push: failed to load local index", "error", err)
        return err
    }
    log.Info("push: local index loaded", "segments", len(localIndex.Segments))
}
```

### When Lower-Level Code CAN Log

**Exception cases** where lower-level code logs (but still pass logger explicitly):

1. **Long-running background processes** (servers, workers) with no clear caller
2. **Infrastructure components** (database connection pools, caches)
3. **Progress updates** for operations taking >1 second
4. **But**: Pass a logger down explicitly: `func Process(ctx context.Context, log *slog.Logger)`

### Log Message Format

**Make log messages self-descriptive** - include the operation name in the message text:

```go
// ✅ GOOD: Operation is clear from message
log.Debug("push: starting", "remote_path", remote.Path)
log.Debug("push: loading local index", "path", localIndexPath)
log.Info("push: local index loaded", "segments", len(localIndex.Segments))

// ❌ BAD: Have to read attributes to understand what's happening
log.Debug("starting", "operation", "push", "remote_path", remote.Path)
log.Debug("loading local index", "operation", "push", "path", localIndexPath)
```

### Scoped Loggers

Use `slog.With()` to create scoped loggers with default attributes:

```go
func Push(remoteName string, remote config.RemoteConfig, space string, stateDir string) error {
    log := slog.With("remote", remoteName, "space", space)
    log.Debug("push: starting")  // Automatically includes remote and space
    log.Debug("push: loading local index", "path", localIndexPath)
    // ...
}
```

### The Decision Rule

**If a function is called from multiple places, it should NOT log.**

Examples:
- **Utility functions** (`LoadIndexFile`, `SaveIndexFile`, `CalculateSHA256`) - NO logging
- **Business logic** (Push, Pull, Export, command handlers) - YES logging
- **Infrastructure** (segment writer writing files) - YES logging with passed logger

### Go Community Consensus

From the Go community and standard library patterns:

1. **"Errors are values"** - return them, don't log them
2. **Libraries return errors, applications log them** - standard Go proverb
3. **Log at the edges** - handlers, main functions, top-level business logic
4. **Pass loggers explicitly** when needed (not via global state)

### Example: tk Push Operation

Good example from `tk/internal/remote/sync.go`:

- `LoadIndexFile()` - utility function, does NOT log
- `SaveIndexFile()` - utility function, does NOT log  
- `Push()` - business logic, logs all operations with context

This prevents duplicate messages like:
```
[08:34:55] DEBUG  push: loading local index
[08:34:55] DEBUG  loading index file        ← Duplicate!
```

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

## tk Issue Tracker

This repository uses **tk** for issue tracking. tk is an event-sourced task tracker with claims-based status, metadata support, and multi-machine sync.

### Installation

Built from this monorepo:
```bash
want mono tk@local
```

### Project Organization

Tasks are organized into projects:

- **Tool projects** (want, conf, tk, print, mdbook, dissect, etc.) - Tasks specific to each tool in the repo
- **mono** - Repository-wide concerns (CI, formatting, build systems, monorepo tooling)
- **Personal projects** (life, infra, ema) - Personal/work tasks

### Common Commands

**List tasks:**
```bash
tk ls                        # All tasks
tk ls -p want                # Want tool tasks only
tk ls -p mono                # Repo-wide tasks
tk ls --json                 # JSON output with full metadata
```

**Create task:**
```bash
tk new "Fix bug" --project want
tk new "Update CI workflow" --project mono
```

**Update status:**
```bash
tk mark want-1 wip           # Mark as in progress
tk mark want-1 done          # Mark as done
tk mark want-1 wip --role agent  # Agent can claim status
```

**View task:**
```bash
tk show want-1               # Show full details
tk show want-1 --json        # JSON output
```

**Add notes:**
```bash
tk note want-1 "Implemented feature X"
```

**Metadata (priority, labels, custom fields):**
```bash
tk meta set want-1 priority 1
tk meta set want-1 labels '["bug","urgent"]'
tk meta get want-1 priority
tk meta list want-1          # Show all metadata
```

**Relationships:**
```bash
tk relate want-1 blocks want-2
tk blockers want-2           # See what's blocking a task
tk blocked                   # List all blocked tasks
```

### For Agents

**Workflow:**
1. Check tasks: `tk ls -p <project>` 
2. Start work: `tk mark <id> wip --role agent`
3. Add progress notes: `tk note <id> "Status update"`
4. Complete: `tk mark <id> done --role agent`
5. Set metadata: `tk meta set <id> priority N --role agent`

**Claims and Authority:**
- Agents use `--role agent` when making status changes
- Human claims override agent claims (authority lattice: human > qa > rel > agent > bot)
- Conflicting claims are preserved as "tentative"

### Database

- **Default location**: `~/.tk/tk.db`
- **Custom location**: Set `TK_DB_PATH` environment variable
  ```bash
  export TK_DB_PATH=/tmp/test.db
  tk ls  # Uses /tmp/test.db instead of ~/.tk/tk.db
  ```
- **Use cases**: Testing, multiple instances, custom locations
- **Diagnostics**: Run `tk debug doctor` to check database health

### Sync (Multi-Machine)

tk supports syncing across machines:
```bash
tk remote add icloud folder ~/Library/Mobile\ Documents/...
tk sync icloud
```

Events are stored as immutable segments that can be synced via iCloud, git, or other remotes.

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

## CLI Output Styling

**All CLI tools must follow the unified styling guidelines defined in [CLI_STYLE_GUIDE.md](./CLI_STYLE_GUIDE.md).**

### Required Library

Use `lib/cli` package for all CLI color formatting:

```go
import "github.com/neongreen/mono/lib/cli"

// Success messages
cli.PrintSuccess("✓ Operation completed")

// Error messages  
cli.PrintError("Error: Failed to process")

// Colored elements
fmt.Printf("Config: %s\n", cli.Path("/path/to/file"))
fmt.Printf("%s: %s\n", cli.Key("setting"), cli.Value("value"))
```

### Semantic Colors

- **Success** (green): `cli.Success()` - checkmarks, success messages
- **Warning** (yellow): `cli.Warning()` - warnings, in-progress status
- **Error** (red): `cli.Error()` - errors, failures
- **Key** (cyan bold): `cli.Key()` - config keys, identifiers
- **Path** (cyan): `cli.Path()` - file paths
- **Value** (green): `cli.Value()` - configured values
- **Type** (yellow): `cli.Type()` - type information
- **Muted** (bright black): `cli.Muted()` - unset values, placeholders
- **Header** (bold): `cli.Header()` - section headers

### Guidelines

- Use colors purposefully, not excessively
- Maintain consistency across all tools (want, conf, jj-run, etc.)
- Colors should enhance readability, not distract
- The `fatih/color` library automatically respects NO_COLOR environment variable

See [CLI_STYLE_GUIDE.md](./CLI_STYLE_GUIDE.md) for complete documentation and examples.

------------------------------------------------------------

## GitHub Copilot Agent

This section contains specific guidance for the GitHub Copilot coding agent.

### Accessing GitHub Pull Request Review Comments

**Problem**: When asked to address PR review comments, the agent may claim it doesn't have access to them, when in fact the GitHub MCP tools are available.

**Solution**: Use the GitHub MCP server tools to fetch PR review comments. Here's the correct workflow:

1. **Find the PR number** by searching for PRs on the branch:
   ```
   Tool: github-mcp-server-search_pull_requests
   Parameters:
     - owner: "neongreen"
     - repo: "mono"
     - query: "head:BRANCH_NAME repo:neongreen/mono"
   ```

2. **Get review comments** (inline code review comments):
   ```
   Tool: github-mcp-server-pull_request_read
   Parameters:
     - method: "get_review_comments"
     - owner: "neongreen"
     - repo: "mono"
     - pullNumber: <PR_NUMBER>
   ```

3. **Get reviews** (overall review summaries):
   ```
   Tool: github-mcp-server-pull_request_read
   Parameters:
     - method: "get_reviews"
     - owner: "neongreen"
     - repo: "mono"
     - pullNumber: <PR_NUMBER>
   ```

4. **Get general comments** (issue comments on the PR):
   ```
   Tool: github-mcp-server-pull_request_read
   Parameters:
     - method: "get_comments"
     - owner: "neongreen"
     - repo: "mono"
     - pullNumber: <PR_NUMBER>
   ```

5. **Get PR details**:
   ```
   Tool: github-mcp-server-pull_request_read
   Parameters:
     - method: "get"
     - owner: "neongreen"
     - repo: "mono"
     - pullNumber: <PR_NUMBER>
   ```

**Important notes**:
- The tool name is `github-mcp-server-pull_request_read`, NOT `pull_request_read`
- Review comments contain inline code feedback with file paths, line numbers, and diffs
- Reviews contain overall review summaries and approval states
- Comments contain general discussion on the PR
- Always check all three to get complete feedback

**Example from PR #259**:
- Review comments included P0 (critical) issues about missing imports
- Review comments included P1 (important) issues about silent error handling
- Each comment had specific file paths, line numbers, and detailed explanations

**Common mistake**: Saying "I don't have access to review comments" without first trying the GitHub MCP tools. Always try fetching them first.
