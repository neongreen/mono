# CLI Style Guide

This document defines the unified design strategy for CLI output in this monorepo. All CLI tools should follow these guidelines to provide a consistent, beautiful, and readable user experience.

## Design Principles

1. **Consistency**: All tools use the same color scheme and formatting patterns
2. **Clarity**: Colors enhance readability and help users quickly identify important information
3. **Accessibility**: Colors are chosen to work well in both dark and light terminal themes
4. **Restraint**: Use colors purposefully, not excessively

## Color Palette

### Semantic Colors

Use these semantic color definitions consistently across all CLI tools:

- **Success/Positive**: Green (`color.FgGreen`)
  - Checkmarks (✓)
  - "done" status
  - Success messages
  - Set/configured values

- **Warning/In-Progress**: Yellow (`color.FgYellow`)
  - "wip" status
  - Warnings
  - Prompts for user action

- **Error/Danger**: Red (`color.FgRed`)
  - Error messages
  - Failed operations
  - Critical warnings

- **Info/Primary**: Cyan (`color.FgCyan`)
  - Paths
  - Primary identifiers
  - Key names in key-value pairs

- **Secondary/Meta**: Blue (`color.FgBlue`)
  - Default values
  - Secondary information
  - Types

- **Muted/Disabled**: Bright Black (`color.FgHiBlack`)
  - Unset values
  - Placeholder text
  - "(not set)" indicators

- **Emphasis**: Bold (`color.Bold`)
  - Headers
  - Important items
  - Combined with colors for extra emphasis

## Libraries

### Required Dependencies

All CLI tools must use these libraries for consistent output:

```go
import (
    "github.com/fatih/color"
    "github.com/jedib0t/go-pretty/v6/table"
    "github.com/jedib0t/go-pretty/v6/text"
)
```

### Shared Package

Use the `lib/cli` package for pre-configured color formatters and utility functions:

```go
import "github.com/neongreen/mono/lib/cli"

// Print success message
cli.Success("✓ Operation completed")

// Print error message  
cli.Error("Failed to process file")

// Colorize a path
fmt.Printf("Config file: %s\n", cli.Path("/path/to/config"))
```

## Common Patterns

### Success Messages

```go
cli.Success("✓ Created repository")
cli.Success("✓ Configuration saved")
```

### Error Messages

```go
cli.Error("Error: Failed to connect to server")
fmt.Fprintf(os.Stderr, "Error: %v\n", err)  // Standard error output
```

### Status Indicators

Use colored symbols for visual status:

```go
cli.Success("✓ Installed")    // Green checkmark
cli.Warning("⚠ Not found")    // Yellow warning
cli.Error("✗ Failed")         // Red X
```

### Key-Value Display

For configuration and info display:

```go
fmt.Printf("%s: %s\n", cli.Key("Setting"), cli.Value("value"))
fmt.Printf("%s: %s\n", cli.Key("Path"), cli.Path("/path/to/file"))
fmt.Printf("%s: %s\n", cli.Key("Status"), cli.Muted("(not set)"))
```

### Tables

Standard table configuration:

```go
t := table.NewWriter()
t.SetOutputMirror(os.Stdout)
t.SetStyle(table.StyleLight)
t.Style().Options.SeparateRows = true
t.Style().Options.DrawBorder = false
```

Color table cells using the clistyle formatters:

```go
t.AppendRow(table.Row{
    cli.Key("setting.path"),
    cli.Type("string"),
    cli.Value(currentValue),
})
```

### Lists

For bullet point lists:

```go
fmt.Printf("  • %s\n", item)         // Regular item
fmt.Printf("  ✓ %s\n", completed)    // Completed item (green)
fmt.Printf("  ⚠ %s\n", warning)      // Warning item (yellow)
```

### Section Headers

Use bold text for section headers:

```go
cli.Header("Configuration Settings")
fmt.Println()  // Add blank line after header
```

### Progress/Status Updates

```go
fmt.Fprintf(os.Stderr, "Processing %d/%d...\n", current, total)
cli.Success(fmt.Sprintf("Processed %d items", count))
```

## Specific Tool Guidelines

### want

- Use success colors (green) for installed items
- Use muted colors for items not yet installed
- Use cyan for repository names and paths
- Display command suggestions in muted colors

### conf

- Use cyan for configuration paths
- Use green for set values with ✓ prefix
- Use blue for default values
- Use bright black for unset values
- Maintain existing table styling

### jj-run

- Use cyan for change IDs (first 12 chars)
- Use yellow for processing status
- Use green for successful operations
- Use red for errors with appropriate exit codes
- Keep stderr messages distinct from stdout

## Implementation Examples

### Example 1: Simple Status Output

```go
package main

import (
    "fmt"
    "github.com/neongreen/mono/lib/cli"
)

func printStatus(installed bool, name string) {
    if installed {
        fmt.Printf("  %s %s\n", cli.Success("✓"), name)
    } else {
        fmt.Printf("  %s %s\n", cli.Muted("○"), name)
    }
}
```

### Example 2: Configuration Display

```go
package main

import (
    "fmt"
    "github.com/neongreen/mono/lib/cli"
)

func showConfig(key, value string, isSet bool) {
    if isSet {
        fmt.Printf("%s = %s\n", 
            cli.Key(key), 
            cli.Value(value))
    } else {
        fmt.Printf("%s = %s\n", 
            cli.Key(key), 
            cli.Muted("(not set)"))
    }
}
```

### Example 3: Error with Context

```go
package main

import (
    "fmt"
    "os"
    "github.com/neongreen/mono/lib/cli"
)

func handleError(operation string, err error) {
    cli.Error(fmt.Sprintf("✗ Failed to %s", operation))
    fmt.Fprintf(os.Stderr, "  %v\n", err)
}
```

## Testing CLI Output

When testing CLI tools:

1. Run the tool in a terminal with colors enabled
2. Check that colors enhance readability
3. Verify colors work in both dark and light terminal themes
4. Ensure output is still readable when colors are disabled (NO_COLOR env var)

The `fatih/color` library automatically respects the `NO_COLOR` environment variable and terminal capabilities.

## Guidelines for Future Tools

When creating new CLI tools:

1. Import `lib/cli` package
2. Use the semantic color functions consistently
3. Follow the table formatting standards
4. Use appropriate symbols (✓, ✗, ⚠, •) for status
5. Keep colors purposeful and not overwhelming
6. Test output with and without colors
