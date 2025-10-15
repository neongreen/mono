# Agent Guidelines for This Monorepo

This document contains guidelines for AI agents and automated tools working on projects in this monorepo.

## Basic rules

- Don't use Markdown files for tracking issues or TODOs. Use Beads. Run `bd quickstart` to see how.
- All tools are written in Go unless stated otherwise. 
- All new projects are created as top-level folders in the repository unless stated otherwise.
- All projects must contain a `mise.toml`. Check existing `mise.toml` files to see what is expected from you.
- All new Go projects must have CI workflows in `.github/workflows/<project-name>.yml`. Check existing workflow files to see what is expected from you.
- In all prose that you write, don't be excited, don't use emojis unless necessary, and don't use pervasive bold text.
- All temporary files (like summaries of fixes you did, one-off scripts you wrote during PR development, etc) must have names prefixed with `ai-temp-`.
- Do not create temporary files in the repository root.
- Keep the list of projects in [README.md](./README.md) up to date.
- Record project status in README.md. If the project is incomplete, it must be stated.
  - Use `bd` to keep track of known bugs, issues, and future work.

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
