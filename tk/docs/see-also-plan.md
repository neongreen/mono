# TK CLI "See Also" Section - Implementation Plan

## Overview

This document outlines the plan to add "See Also" sections to TK's CLI help pages to improve command discoverability, especially for AI agents who might not guess at less obvious command names.

## Problem Statement

Current issues with command discovery:
1. Agents tend to guess at common command names (create, delete, list) but miss specialized features
2. No guidance on related commands when viewing help
3. Advanced features (relations, containers, metadata) are underutilized due to low discoverability
4. No workflow guidance (e.g., "after creating tasks, you might want to relate them")

## Goals

1. **Improve discoverability**: Help users/agents find related commands they didn't know existed
2. **Guide workflows**: Suggest logical next steps or related operations
3. **Cross-promote features**: Highlight less-obvious but powerful features
4. **Maintain simplicity**: Don't overwhelm help text; keep suggestions relevant

---

## Part 1: Cross-Linking Strategy

### Command Grouping and Cross-References

#### **1. Core Task Management Cluster**

**Primary workflow**: create → view → edit → status → delete

| Command | See Also |
|---------|----------|
| `new` | `show`, `mark`, `edit`, `describe`, `ls` |
| `show` | `edit`, `describe`, `note`, `history`, `relate ls`, `attach` |
| `ls` | `show`, `mark`, `new`, `project ls`, `blocked` |
| `mark` | `show`, `ls`, `blockers` |
| `edit` | `describe`, `show`, `note`, `mv` |
| `describe` | `edit`, `show`, `note` |
| `note` | `show`, `describe`, `history` |
| `attach` | `show`, `note` |
| `mv` | `edit`, `project ls`, `show` |
| `rm` | `show`, `ls` |

**Rationale**:
- `new` users should know about viewing (`show`), organizing (`ls`), and editing (`edit`, `describe`)
- `show` is the hub - point to all task manipulation commands
- `edit` and `describe` are similar, so cross-reference each other
- `note` and `attach` are related ways to add information

#### **2. Relations & Dependencies Cluster**

**Key insight**: This is a powerful feature that agents might not discover

| Command | See Also |
|---------|----------|
| `relate` | `relate add`, `relate ls`, `dup`, `blockers`, `graph` |
| `relate add` | `relate ls`, `relate remove`, `dup`, `graph` |
| `relate remove` | `relate ls`, `relate add` |
| `relate ls` | `relate add`, `graph`, `show`, `blockers` |
| `dup` | `relate add`, `relate ls` |
| `blockers` | `blocked`, `relate add`, `mark`, `graph` |
| `blocked` | `blockers`, `mark`, `relate ls` |
| `graph` | `relate ls`, `blockers`, `show` |

**Rationale**:
- `relate` commands should all cross-reference each other
- `dup` is specialized relation - point to general `relate add`
- `blockers`/`blocked` are high-value for workflow management
- `graph` provides visualization - connect to data sources

#### **3. Project Management Cluster**

| Command | See Also |
|---------|----------|
| `project` | `project create`, `project ls`, `project rename` |
| `project create` | `new`, `mv`, `project ls` |
| `project ls` | `project create`, `ls`, `mv` |
| `project rm` | `project ls`, `mv` |
| `project rename` | `project ls` |
| `mv` | `project ls`, `project create`, `edit` |

**Rationale**:
- After creating projects, users want to create tasks or move existing ones
- `mv` is the bridge between tasks and projects
- `ls` with project filtering connects task and project views

#### **4. Container Management Cluster**

**Key insight**: Queues, stacks, and groups are advanced features that need promotion

| Command | See Also |
|---------|----------|
| `queue` | `queue create`, `queue push`, `queue pop`, `queue list`, `stack`, `group` |
| `queue create` | `queue push`, `queue list`, `new` |
| `queue push` | `queue pop`, `queue show`, `ls` |
| `queue pop` | `queue push`, `queue show`, `mark` |
| `queue list` | `queue show`, `queue create`, `stack list`, `group list` |
| `queue show` | `queue push`, `queue pop`, `ls` |
| `stack` | `stack create`, `stack push`, `stack pop`, `stack list`, `queue`, `group` |
| `stack create` | `stack push`, `stack list`, `new` |
| `stack push` | `stack pop`, `stack show`, `ls` |
| `stack pop` | `stack push`, `stack show`, `mark` |
| `stack list` | `stack show`, `stack create`, `queue list`, `group list` |
| `stack show` | `stack push`, `stack pop`, `ls` |
| `group` | `group create`, `group add`, `group list`, `queue`, `stack` |
| `group create` | `group add`, `group list` |
| `group add` | `group show`, `group remove`, `ls` |
| `group remove` | `group add`, `group show` |
| `group list` | `group show`, `group create`, `queue list`, `stack list` |
| `group show` | `group add`, `ls` |

**Rationale**:
- Cross-promote queue/stack/group as alternative organizational methods
- Connect container operations to task operations (`ls`, `mark`)
- Help users discover all three container types

#### **5. Schema & Metadata Cluster**

| Command | See Also |
|---------|----------|
| `schema` | `schema add`, `schema list`, `schema export`, `meta` |
| `schema add` | `schema list`, `meta set` |
| `schema list` | `schema add`, `schema export`, `meta list` |
| `schema export` | `schema list`, `schema add` |
| `meta` | `meta set`, `meta get`, `meta list`, `schema` |
| `meta set` | `meta get`, `meta list`, `show` |
| `meta get` | `meta set`, `meta list`, `show` |
| `meta list` | `meta get`, `meta claims`, `schema list` |
| `meta claims` | `meta list`, `meta get` |

**Rationale**:
- Schema and metadata work together
- Connect metadata operations to viewing tasks
- Help users understand the full metadata system

#### **6. Sync & Remote Cluster**

| Command | See Also |
|---------|----------|
| `remote` | `remote add`, `remote ls`, `push`, `pull`, `sync` |
| `remote add` | `remote ls`, `sync`, `push` |
| `remote ls` | `remote add`, `remote rm`, `status sync` |
| `remote rm` | `remote ls` |
| `push` | `pull`, `sync`, `remote ls`, `status sync` |
| `pull` | `push`, `sync`, `ingest` |
| `sync` | `push`, `pull`, `status sync`, `remote ls` |
| `ingest` | `pull`, `sync` |
| `status sync` | `sync`, `remote ls`, `push`, `pull` |

**Rationale**:
- Sync workflow: configure remote → sync/push/pull
- Connect status checking to sync operations
- Help users understand the full sync cycle

#### **7. Debugging & Maintenance Cluster**

| Command | See Also |
|---------|----------|
| `debug` | `debug doctor`, `debug repair`, `debug rebuild` |
| `debug doctor` | `debug repair`, `conflicts` |
| `debug repair` | `debug doctor`, `debug rebuild` |
| `debug rebuild` | `debug repair`, `ingest` |
| `debug events` | `debug events list`, `debug events show`, `debug events stats`, `log query` |
| `debug events list` | `debug events show`, `debug events stats` |
| `debug events show` | `debug events list`, `history` |
| `debug node` | `debug node show`, `debug node regen`, `remote ls` |
| `conflicts` | `conflicts numbers`, `debug doctor`, `edit` |
| `history` | `show`, `log query`, `debug events show` |
| `log query` | `log search`, `debug events list`, `history` |
| `log search` | `log query` |

**Rationale**:
- Guide users through debugging workflow
- Connect conflicts to resolution commands
- Link event/log inspection commands

#### **8. Migration & Setup Cluster**

| Command | See Also |
|---------|----------|
| `init` | `new`, `project create`, `remote add` |
| `migrate` | `migrate scan-deprecated`, `debug doctor` |
| `import-beads` | `new`, `ls`, `ingest` |
| `version` | `debug doctor` |

**Rationale**:
- After `init`, guide users to first steps
- Connect migration to health checking
- Help with data import workflows

---

## Part 2: Implementation Approach

### Option A: Extend Cobra's Long Field (Recommended)

**Approach**: Append "See Also:" section to each command's `Long` field.

**Example**:
```go
var showCmd = &cobra.Command{
    Use:   "show <task>",
    Short: "Show task details",
    Long: `Display detailed information about a task.

Examples:
  tk show 42
  tk show #42

The command shows the task's title, status, notes, relations, and history.

See Also:
  tk edit      - Edit task fields
  tk describe  - Update task title
  tk note      - Add notes to a task
  tk history   - View task history
  tk relate ls - List task relations
  tk attach    - Manage task attachments`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // implementation
    },
}
```

**Pros**:
- ✅ Simple to implement
- ✅ Works with existing Cobra infrastructure
- ✅ No custom templates needed
- ✅ Appears in `--help` output automatically
- ✅ Easy to maintain

**Cons**:
- ❌ Manual formatting required
- ❌ Can't automatically verify command names exist
- ❌ Duplicates command name strings

### Option B: Custom Cobra Template

**Approach**: Create a custom help template that adds a "See Also" section.

**Implementation**:
```go
const customHelpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
{{if .SeeAlso}}

See Also:
{{range .SeeAlso}}  {{.}}
{{end}}{{end}}
`

// Add SeeAlso field to commands via Annotations
var showCmd = &cobra.Command{
    Use:   "show <task>",
    Short: "Show task details",
    Long:  `...`,
    Annotations: map[string]string{
        "seeAlso": "edit, describe, note, history, relate ls, attach",
    },
    RunE: func(cmd *cobra.Command, args []string) error {
        // implementation
    },
}

func init() {
    // Set custom template on root
    rootCmd.SetHelpTemplate(customHelpTemplate)
}
```

**Pros**:
- ✅ Consistent formatting
- ✅ Structured data (Annotations)
- ✅ Easier to parse/validate
- ✅ Can add features later (clickable links in terminal?)

**Cons**:
- ❌ More complex implementation
- ❌ Custom template maintenance
- ❌ Need to understand Cobra templates

### Option C: Helper Function + Long Field

**Approach**: Create a helper function to format "See Also" sections consistently.

**Implementation**:
```go
// In cmd/help.go
package cmd

import "strings"

// SeeAlso formats a "See Also" section for command help text
func SeeAlso(commands ...string) string {
    if len(commands) == 0 {
        return ""
    }

    var b strings.Builder
    b.WriteString("\n\nSee Also:\n")
    for _, cmd := range commands {
        b.WriteString("  tk ")
        b.WriteString(cmd)
        b.WriteString("\n")
    }
    return b.String()
}

// Usage in commands
var showCmd = &cobra.Command{
    Use:   "show <task>",
    Short: "Show task details",
    Long: `Display detailed information about a task.

Examples:
  tk show 42
  tk show #42

The command shows the task's title, status, notes, relations, and history.` +
    SeeAlso("edit", "describe", "note", "history", "relate ls", "attach"),
    RunE: func(cmd *cobra.Command, args []string) error {
        // implementation
    },
}
```

**Pros**:
- ✅ Consistent formatting via helper
- ✅ Simple implementation
- ✅ Easy to add descriptions: `SeeAlso("edit - Edit task", "note - Add notes")`
- ✅ Type-safe command names (could add validation)
- ✅ No custom templates

**Cons**:
- ❌ Still manual work per command
- ❌ String concatenation

### Option D: Centralized Registry + Helper

**Approach**: Define all "See Also" relationships in one place, apply via helper.

**Implementation**:
```go
// In cmd/help.go
package cmd

var seeAlsoRegistry = map[string][]string{
    "new":    {"show", "mark", "edit", "describe", "ls"},
    "show":   {"edit", "describe", "note", "history", "relate ls", "attach"},
    "ls":     {"show", "mark", "new", "project ls", "blocked"},
    // ... all commands
}

func ApplySeeAlso(cmd *cobra.Command) {
    if related, ok := seeAlsoRegistry[cmd.Name()]; ok {
        cmd.Long = cmd.Long + SeeAlso(related...)
    }
}

// Usage in init() functions
func init() {
    RootCmd.AddCommand(showCmd)
    ApplySeeAlso(showCmd)
}
```

**Pros**:
- ✅ Centralized relationship definitions
- ✅ Easy to maintain and audit
- ✅ Can validate all command names exist
- ✅ Can generate documentation from registry
- ✅ DRY principle

**Cons**:
- ❌ Extra init step per command
- ❌ Separation between command definition and relationships

---

## Recommended Implementation: Hybrid Approach

**Combination of Options C + D**:

1. **Helper function** for formatting (Option C)
2. **Centralized registry** for relationships (Option D)
3. **Automatic application** via root initialization

**Why this approach**:
- ✅ Consistency: All formatting in one place
- ✅ Maintainability: All relationships in one registry
- ✅ Validation: Can check all commands exist
- ✅ Flexibility: Can override per-command if needed
- ✅ Documentation: Registry doubles as documentation
- ✅ Simplicity: No custom Cobra templates

### Implementation Steps

1. **Create `cmd/help.go`** with:
   - `seeAlsoRegistry` map
   - `SeeAlso()` formatter function
   - `ApplySeeAlso()` applicator function
   - `ValidateSeeAlso()` function (for tests only)

2. **Create `cmd/help_test.go`** with:
   - `TestSeeAlsoValidation()` - ensures all references are valid
   - `TestSeeAlsoRegistry()` - checks registry invariants

3. **Update `cmd/root.go`**:
   - Import help package
   - Call `ApplySeeAlso()` for all commands after registration

4. **Populate registry** following the cross-linking strategy above

5. **Test**:
   - Run `go test ./cmd/...` to verify validation passes
   - Manually check help output for formatting
   - Verify all commands show appropriate cross-references

### File Structure
```
tk/cmd/
├── help.go              # NEW: See Also registry and helpers
├── help_test.go         # NEW: Tests for See Also validation
├── root.go              # MODIFIED: Apply See Also to all commands
└── [other commands]     # UNCHANGED
```

### Code Skeleton

```go
// cmd/help.go
package cmd

import (
    "fmt"
    "strings"
    "github.com/spf13/cobra"
)

// seeAlsoRegistry maps command names to related commands
var seeAlsoRegistry = map[string][]string{
    // Core task management
    "new":      {"show", "mark", "edit", "describe", "ls"},
    "show":     {"edit", "describe", "note", "history", "relate ls", "attach"},
    "ls":       {"show", "mark", "new", "project ls", "blocked"},
    // ... (full registry from Part 1)
}

// SeeAlso formats a "See Also" section for command help
func SeeAlso(commands ...string) string {
    if len(commands) == 0 {
        return ""
    }

    var b strings.Builder
    b.WriteString("\n\nSee Also:")
    for _, cmd := range commands {
        b.WriteString("\n  tk ")
        b.WriteString(cmd)
    }
    return b.String()
}

// ApplySeeAlso adds "See Also" section to a command from the registry
func ApplySeeAlso(cmd *cobra.Command) {
    cmdPath := cmd.CommandPath()[3:] // Remove "tk " prefix
    if related, ok := seeAlsoRegistry[cmdPath]; ok {
        cmd.Long = strings.TrimSpace(cmd.Long) + SeeAlso(related...)
    }

    // Recursively apply to subcommands
    for _, subcmd := range cmd.Commands() {
        ApplySeeAlso(subcmd)
    }
}

// ValidateSeeAlso checks all referenced commands exist using Cobra's command lookup
func ValidateSeeAlso(root *cobra.Command) error {
    var errors []string

    for cmdPath, relatedCmds := range seeAlsoRegistry {
        // Verify the source command exists
        sourceCmd, _, err := root.Find(strings.Fields(cmdPath))
        if err != nil || sourceCmd == nil {
            errors = append(errors, fmt.Sprintf("source command %q not found in command tree", cmdPath))
            continue
        }

        // Verify each related command exists
        for _, relPath := range relatedCmds {
            targetCmd, _, err := root.Find(strings.Fields(relPath))
            if err != nil || targetCmd == nil {
                errors = append(errors,
                    fmt.Sprintf("command %q references non-existent command %q", cmdPath, relPath))
            }
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("See Also validation failed:\n  - %s", strings.Join(errors, "\n  - "))
    }
    return nil
}

// findCommand is a helper that uses Cobra's Find to locate a command by path
func findCommand(root *cobra.Command, path string) (*cobra.Command, error) {
    args := strings.Fields(path)
    cmd, _, err := root.Find(args)
    if err != nil {
        return nil, fmt.Errorf("command %q not found: %w", path, err)
    }
    if cmd == nil {
        return nil, fmt.Errorf("command %q not found", path)
    }
    return cmd, nil
}
```

```go
// cmd/root.go (additions)
func init() {
    // ... existing init code ...

    // Apply "See Also" sections to all commands
    ApplySeeAlso(RootCmd)

    // NOTE: Validation is done in tests (see cmd/help_test.go)
    // not at runtime, since users can't fix registry issues anyway
}
```

### Benefits of Test-Time Validation

Validation runs **only in tests**, not at runtime. This makes sense because:

1. **Users can't fix registry bugs** - broken references are code bugs, not user errors
2. **Caught during development** - Tests fail in CI before code is merged
3. **No runtime overhead** - Production code doesn't waste time validating
4. **Clear feedback** - Test failures show exactly what's wrong

Using Cobra's `Find()` method in tests provides:

1. **Accurate validation**: Uses the same lookup mechanism Cobra uses internally
2. **Handles subcommands**: Automatically works with nested command paths like "relate add"
3. **Catches typos**: Any misspelled command name in the registry will fail tests
4. **Maintains consistency**: References are validated against actual command structure
5. **Refactor-safe**: If you rename a command, validation tests will fail immediately

### Example Test Failure

When validation fails in tests, you'll see clear error messages:
```
--- FAIL: TestSeeAlsoValidation (0.00s)
    help_test.go:10: See Also validation failed:
          - command "new" references non-existent command "shw" (typo for "show")
          - command "relate ls" references non-existent command "graff" (typo for "graph")
          - source command "deleted-command" not found in command tree
FAIL
```

### Adding Tests

Create a test to ensure validation always passes:

```go
// cmd/help_test.go
package cmd

import (
    "testing"
)

func TestSeeAlsoValidation(t *testing.T) {
    // This test ensures all "See Also" references are valid
    if err := ValidateSeeAlso(RootCmd); err != nil {
        t.Fatalf("See Also validation failed:\n%v", err)
    }
}

func TestSeeAlsoRegistry(t *testing.T) {
    // Ensure registry is not empty
    if len(seeAlsoRegistry) == 0 {
        t.Error("seeAlsoRegistry is empty - did you populate it?")
    }

    // Ensure no command references itself
    for cmd, related := range seeAlsoRegistry {
        for _, rel := range related {
            if cmd == rel {
                t.Errorf("command %q references itself in See Also", cmd)
            }
        }
    }
}
```

---

## Testing Plan

1. **Manual testing**:
   ```bash
   tk new --help        # Should show "See Also: show, mark, edit..."
   tk show --help       # Should show related commands
   tk queue push --help # Should show queue/stack related commands
   ```

2. **Validation testing**:
   ```bash
   tk-dev debug doctor  # Should pass validation
   # Add test that calls ValidateSeeAlso()
   ```

3. **Formatting testing**:
   - Verify consistent spacing
   - Check alignment
   - Ensure no double newlines

4. **Coverage testing**:
   - Ensure all commands in registry have "See Also"
   - Identify commands missing from registry

---

## Migration Path

1. **Phase 1**: Implement infrastructure (help.go, helpers)
2. **Phase 2**: Add core command relationships (new, show, ls, mark)
3. **Phase 3**: Add advanced features (relate, containers, meta)
4. **Phase 4**: Add maintenance commands (debug, sync, migrate)
5. **Phase 5**: Validation and cleanup

---

## Future Enhancement: Adding Descriptions

**Start simple**: Begin with just command names (no descriptions). This keeps the initial implementation clean and focused.

**Add descriptions later** if needed, without changing the registry structure:

### Option 1: Separate Description Registry

```go
// Registry stays simple - just command names
var seeAlsoRegistry = map[string][]string{
    "show": {"edit", "describe", "note", "history"},
}

// Optional: Add descriptions in a separate registry
var commandDescriptions = map[string]string{
    "edit":     "Edit task fields",
    "describe": "Update task title",
    "note":     "Add notes to task",
    "history":  "View task history",
}

// Enhanced formatter that uses descriptions if available
func SeeAlso(commands ...string) string {
    var b strings.Builder
    b.WriteString("\n\nSee Also:")
    for _, cmd := range commands {
        b.WriteString("\n  tk ")
        b.WriteString(cmd)

        // Add description if available
        if desc, ok := commandDescriptions[cmd]; ok {
            // Pad command name to 15 chars for alignment
            padding := 15 - len(cmd)
            if padding > 0 {
                b.WriteString(strings.Repeat(" ", padding))
            }
            b.WriteString(" - ")
            b.WriteString(desc)
        }
    }
    return b.String()
}
```

### Option 2: Extract from Command.Short

Even simpler - just pull descriptions from the command's existing `Short` field:

```go
func SeeAlsoWithDescriptions(root *cobra.Command, commands ...string) string {
    var b strings.Builder
    b.WriteString("\n\nSee Also:")

    for _, cmdPath := range commands {
        cmd, err := findCommand(root, cmdPath)

        b.WriteString("\n  tk ")
        b.WriteString(cmdPath)

        // Use the command's Short description if available
        if err == nil && cmd != nil && cmd.Short != "" {
            padding := 15 - len(cmdPath)
            if padding > 0 {
                b.WriteString(strings.Repeat(" ", padding))
            }
            b.WriteString(" - ")
            b.WriteString(cmd.Short)
        }
    }
    return b.String()
}
```

### Comparison

**Without descriptions** (recommended for initial implementation):
```
See Also:
  tk edit
  tk describe
  tk note
  tk history
```

**With descriptions** (future enhancement):
```
See Also:
  tk edit        - Edit task fields
  tk describe    - Update task title
  tk note        - Add notes to task
  tk history     - View task history
```

**Recommendation**: Start without descriptions. They can be added later by:
1. Modifying just the `SeeAlso()` formatter function
2. No changes needed to the registry
3. No changes needed to validation
4. Backward compatible enhancement

---

## Summary

### What to Implement
1. **Centralized registry** mapping commands to related commands (Part 1 table)
2. **Helper functions** for formatting "See Also" sections
3. **Automatic application** via root command initialization
4. **Test-time validation** to ensure referenced commands exist (no runtime overhead)

### Key Benefits
- 🎯 Improves command discoverability
- 🤖 Helps AI agents find less-obvious features
- 📚 Creates workflow guidance
- 🔗 Cross-promotes powerful features (relations, containers, metadata)
- ✅ Maintains clean, consistent help output
- 🛠️ Easy to maintain and extend
- ⚡ Zero runtime overhead (validation only in tests)
- 🔒 Refactor-safe (tests catch broken references immediately)

### Next Steps
1. Implement cmd/help.go with registry and helpers
2. Implement cmd/help_test.go with validation tests
3. Update cmd/root.go to apply See Also sections
4. Populate registry following cross-linking strategy
5. Run tests to verify all references are valid
6. Manually test help output formatting
