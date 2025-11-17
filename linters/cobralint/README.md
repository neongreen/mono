# Cobra Command Linter

A linter for Cobra commands that can enforce project-specific conventions.

## Features

- Extracts structured information about Cobra commands from Go source code
- Provides a framework for custom checkers
- Built on the standard Go analysis framework

## Installation

```bash
go install github.com/neongreen/mono/linters/cobralint/cmd/cobralint@latest
```

Or build from source:

```bash
cd linters/cobralint/cmd/cobralint
go build -o cobralint
```

## Usage

Run the linter on your Go packages:

```bash
cobralint ./...
```

Or on specific files:

```bash
cobralint ./cmd/...
```

## Built-in Checkers

### RequireJSONFlagChecker

Enforces that all Cobra commands have a `--json` flag. This is useful for ensuring consistent output formatting across all commands.

Example violation:

```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Does something",
    RunE: func(cmd *cobra.Command, args []string) error {
        return nil
    },
}
// Missing: myCmd.Flags().Bool("json", false, "Output as JSON")
```

The linter will report:

```
command "myCmd" (use: "mycommand") missing required --json flag
```

#### Exempting Commands from JSON Flag Requirement

The JSON flag requirement applies to **read-only commands** that query and display data. Commands that modify state don't need JSON output. You can exempt such commands using the `cobralint:exemptjson` directive:

```go
// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var createCmd = &cobra.Command{
    Use:   "create",
    Short: "Create a new item",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Creates a new item
        return nil
    },
}

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var deleteCmd = &cobra.Command{
    Use:   "delete",
    Short: "Delete an item",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Deletes an item
        return nil
    },
}
```

The directive must:
- Be placed in a comment directly above the command variable declaration
- Include a reason explaining why the command doesn't need JSON output
- Follow the format: `// cobralint:exemptjson reason: <your explanation>`

**Common exemption reasons:**
- `Modifies state; JSON only required for read-only commands` - for create/update/delete commands
- `Interactive command; JSON only required for read-only commands` - for commands requiring user input

Commands with valid exemption directives will not trigger the "missing required --json flag" error.

## Writing Custom Checkers

You can create custom checkers by implementing the `Checker` interface:

```go
type MyCustomChecker struct{}

func (c *MyCustomChecker) Name() string {
    return "my-custom-checker"
}

func (c *MyCustomChecker) Check(pass *analysis.Pass, cmd *CommandInfo) {
    // Your custom logic here
    // Example: check if command names follow a convention
    if !strings.HasSuffix(cmd.Name, "Cmd") {
        pass.Reportf(cmd.Pos, "command variable name %q should end with 'Cmd'", cmd.Name)
    }
}
```

Then add it to the `EnabledCheckers` list in `cobralint.go`:

```go
var EnabledCheckers = []Checker{
    &RequireJSONFlagChecker{},
    &MyCustomChecker{},
}
```

## Extracted Command Information

The linter extracts the following information about each command:

- `Name`: Variable name (e.g., "lsCmd")
- `Use`: Command use string (e.g., "ls")
- `Flags`: List of flags with their names, types, and short names
- `IsRoot`: Whether this is the root command

## Example

See `testdata/src/a/a.go` for a simple example of commands that the linter can analyze.

## Testing

Run the tests:

```bash
go test ./...
```

## Integration with CI

You can integrate this linter into your CI pipeline:

```yaml
- name: Run Cobra linter
  run: |
    go install github.com/neongreen/mono/linters/cobralint/cmd/cobralint@latest
    cobralint ./...
```

## Architecture

The linter works in two passes:

1. **Command extraction**: Scans the AST to find all `cobra.Command` variable declarations and extracts basic information (name, Use field)
2. **Flag extraction**: Scans for `cmdName.Flags().FlagType(...)` calls to extract flag information

Then it runs all enabled checkers on each command found.

## Supported Flag Patterns

The linter detects flags added using both regular and `*Var` methods:

```go
// Regular methods (flag name is first argument)
cmd.Flags().Bool("json", false, "Output as JSON")
cmd.Flags().BoolP("json", "j", false, "Output as JSON")
cmd.Flags().String("output", "", "Output file")
cmd.Flags().StringP("output", "o", "", "Output file")

// *Var methods (flag name is second argument)
var jsonFlag bool
cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
cmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output as JSON")

var outputFile string
cmd.Flags().StringVar(&outputFile, "output", "", "Output file")
cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file")
```

## Limitations

- Only detects flags added with direct `cmdName.Flags()` method calls. Flags added through other mechanisms may not be detected.
- Requires that commands are declared as package-level variables.
- Does not follow subcommand relationships (yet).
