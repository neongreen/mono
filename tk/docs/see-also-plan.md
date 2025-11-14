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
   - Validation function to check command names

2. **Update `cmd/root.go`**:
   - Import help package
   - Call `ApplySeeAlso()` for all commands after registration
   - Add validation in debug mode

3. **Populate registry** following the cross-linking strategy above

4. **Test**:
   - Verify help output for all commands
   - Check formatting consistency
   - Validate all referenced commands exist

### File Structure
```
tk/cmd/
├── help.go              # NEW: See Also registry and helpers
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

// ValidateSeeAlso checks all referenced commands exist (for testing)
func ValidateSeeAlso(root *cobra.Command) error {
    allCommands := make(map[string]bool)

    var collect func(*cobra.Command, string)
    collect = func(cmd *cobra.Command, prefix string) {
        path := prefix + cmd.Name()
        allCommands[path] = true
        for _, subcmd := range cmd.Commands() {
            collect(subcmd, path + " ")
        }
    }
    collect(root, "")

    for cmd, related := range seeAlsoRegistry {
        for _, rel := range related {
            if !allCommands[rel] {
                return fmt.Errorf("command %q references unknown command %q", cmd, rel)
            }
        }
    }
    return nil
}
```

```go
// cmd/root.go (additions)
func init() {
    // ... existing init code ...

    // Apply "See Also" sections to all commands
    ApplySeeAlso(RootCmd)

    // In debug mode, validate all references
    if debug {
        if err := ValidateSeeAlso(RootCmd); err != nil {
            fmt.Fprintf(os.Stderr, "Warning: See Also validation failed: %v\n", err)
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

## Alternative: Description-Rich See Also

Instead of just command names, include brief descriptions:

```go
func SeeAlsoWithDesc(items ...string) string {
    // items format: "command - description", "command - description"
    var b strings.Builder
    b.WriteString("\n\nSee Also:\n")
    for _, item := range items {
        parts := strings.SplitN(item, " - ", 2)
        cmd := parts[0]
        desc := ""
        if len(parts) > 1 {
            desc = " - " + parts[1]
        }
        fmt.Fprintf(&b, "  tk %-15s%s\n", cmd, desc)
    }
    return b.String()
}

// Usage:
SeeAlsoWithDesc(
    "edit - Edit task fields",
    "note - Add notes to task",
    "history - View task history",
)
```

**Output**:
```
See Also:
  tk edit            - Edit task fields
  tk note            - Add notes to task
  tk history         - View task history
```

This could be a future enhancement if plain command names aren't descriptive enough.

---

## Summary

### What to Implement
1. **Centralized registry** mapping commands to related commands (Part 1 table)
2. **Helper functions** for formatting "See Also" sections
3. **Automatic application** via root command initialization
4. **Validation** to ensure referenced commands exist

### Key Benefits
- 🎯 Improves command discoverability
- 🤖 Helps AI agents find less-obvious features
- 📚 Creates workflow guidance
- 🔗 Cross-promotes powerful features (relations, containers, metadata)
- ✅ Maintains clean, consistent help output
- 🛠️ Easy to maintain and extend

### Next Steps
1. Review and approve this plan
2. Implement cmd/help.go with registry and helpers
3. Update cmd/root.go to apply See Also sections
4. Test with representative commands
5. Roll out to all commands
6. Add validation to CI/tests
