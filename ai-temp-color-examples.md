# CLI Color Examples

This document shows examples of the colored output from the three CLI tools.

## Color Palette

The unified color palette used across all tools:

| Color | Usage | Function |
|-------|-------|----------|
| 🟢 Green | Success, completion, set values | `cli.Success()`, `cli.Value()` |
| 🟡 Yellow | Warnings, types, in-progress | `cli.Warning()`, `cli.Type()` |
| 🔴 Red | Errors, failures, deletions | `cli.Error()` |
| 🔵 Cyan | Keys, IDs, paths | `cli.Key()`, `cli.Path()` |
| 🔵 Blue | Secondary info, defaults | `cli.Secondary()` |
| ⚫ Gray | Muted, unset values | `cli.Muted()` |
| **Bold** | Headers, emphasis | `cli.Header()` |

## Example 1: want - Package Manager

### Command: `want list`
```
MVP: No requirements tracked yet              [Gray - muted]

This command will show:
  • Tools installed via want
  • Repositories cloned via want
  • Their current status
```

### Command: `want mono printpdf --dry-run`
```
Downloading GitHub release asset from: https://github.com/...
Owner: neongreen                               [Cyan - key]
Repo: mono                                     [Cyan - key]
Tag: main.1                                    [Cyan - key]

✓ Downloaded to: /usr/local/bin/printpdf      [Green ✓, Cyan path]
✓ Binary is available in your PATH            [Green ✓]
```

### Error Example: `want forget nonexistent`
```
MVP: Would forget: nonexistent                 [Gray MVP:, Cyan key]
This command will remove the requirement from tracking without uninstalling.
```

## Example 2: conf - Configuration Manager

### Command: `conf jj --list`
```
Setting         Type     Value                    Description
─────────────────────────────────────────────────────────────
user.name       string   ✓ "John Doe"            [Cyan key, Yellow type, Green value]
user.email      string   ✓ "john@example.com"    [Cyan key, Yellow type, Green value]
core.editor     string   (not set)               [Cyan key, Yellow type, Gray muted]
ui.pager        string   (default: "less")       [Cyan key, Yellow type, Blue default]

Config file: /home/user/.config/jj/config.toml [Cyan path]
```

### Diff Output: `conf apply jj`
```
Applying jj configuration...

--- config.toml (before)
+++ config.toml (after)

-user.name = "Old Name"                        [Red - deletion]
+user.name = "John Doe"                        [Green - addition]

✓ Applied configuration successfully           [Green ✓]
```

## Example 3: jj-run - Jujutsu Batch Operations

### Command: `jj-run -r 'mine()' echo "test"`
```
Current operation: abc123def456. To revert, run: [Cyan key for operation ID]
  jj op restore abc123def456                       [Cyan key]

Processing 3 changes in direct mode...              [Bold header]

Processing change 1/3 xyz789abc123: Fix parser     [Cyan key for change ID]
test

Processing change 2/3 def456ghi789: Add feature    [Cyan key]
test

Processed 3/3 changes.                             [Green success]
```

### Error Example
```
Processing change 1/5 abc123def456: Bad commit     [Cyan key]
Error editing change: exit status 1                 [Red error]
Not all changes were processed successfully.        [Yellow warning]
```

## Code Usage Examples

### Success with Path
```go
fmt.Printf("%s Downloaded to: %s\n", 
    cli.Success("✓"), 
    cli.Path("/usr/local/bin/tool"))
```
Output: `✓ Downloaded to: /usr/local/bin/tool` (Green ✓, Cyan path)

### Configuration Key-Value
```go
fmt.Printf("%s: %s\n", 
    cli.Key("user.name"), 
    cli.Value("John Doe"))
```
Output: `user.name: John Doe` (Cyan bold key, Green value)

### Warning Message
```go
fmt.Printf("%s %s is not in your PATH\n", 
    cli.Warning("Note:"), 
    cli.Path("/usr/local/bin"))
```
Output: `Note: /usr/local/bin is not in your PATH` (Yellow warning, Cyan path)

### Error with Context
```go
fmt.Fprintf(os.Stderr, "%s Failed to process change %s\n", 
    cli.Error("Error:"), 
    cli.Key("abc123"))
```
Output: `Error: Failed to process change abc123` (Red error, Cyan key)

### Status Indicators
```go
if isSet {
    fmt.Printf("%s %s\n", cli.Success("✓"), item)
} else {
    fmt.Printf("%s %s\n", cli.Muted("○"), item)
}
```
Output: `✓ item` (Green) or `○ item` (Gray)

### Table with Colors
```go
import (
    "github.com/jedib0t/go-pretty/v6/table"
    "github.com/neongreen/mono/lib/cli"
)

t := table.NewWriter()
t.SetOutputMirror(os.Stdout)
t.SetStyle(table.StyleLight)
t.Style().Options.SeparateRows = true
t.Style().Options.DrawBorder = false

t.AppendHeader(table.Row{"Setting", "Type", "Value"})
t.AppendRow(table.Row{
    cli.Key("user.name"),
    cli.Type("string"),
    cli.Success("✓ \"John Doe\""),
})
t.AppendRow(table.Row{
    cli.Key("core.editor"),
    cli.Type("string"),
    cli.Muted("(not set)"),
})
t.Render()
```

## NO_COLOR Support

All colored output automatically respects:
- `NO_COLOR=1` environment variable
- Non-TTY output (pipes, redirects)
- Terminal capabilities

Examples:
```bash
# Colored output (in terminal)
$ want list

# Plain output (piped)
$ want list | cat

# Plain output (NO_COLOR)
$ NO_COLOR=1 want list

# Plain output (redirected)
$ want list > output.txt
```

## Benefits

1. **Consistency**: Same colors across all tools
2. **Readability**: Quick visual scanning
3. **Accessibility**: Works in light and dark themes
4. **Automatic**: NO_COLOR support built-in
5. **Maintainable**: Single source of truth
6. **Extensible**: Easy to add to new tools

## Testing

All tools tested with:
- ✅ Interactive terminal output (colors visible)
- ✅ Piped output (colors automatically disabled)
- ✅ NO_COLOR=1 (colors disabled)
- ✅ File redirection (colors disabled)
- ✅ go test passing
- ✅ go fmt applied

## Summary

The unified CLI styling system provides:
- Professional appearance
- Consistent user experience
- Better readability
- Semantic color meanings
- Automatic accessibility support
- Easy maintenance and extension
