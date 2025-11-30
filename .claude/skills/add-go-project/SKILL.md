---
name: add-go-project
description: Creates a new Go project in the monorepo with all necessary files (main.go, README, AGENTS.md, etc.). Use when asked to create, add, or scaffold a new Go project, CLI tool, or library. This is a single-module monorepo - all projects share the root go.mod.
---

# Add Go Project Skill

## Purpose

Automates the creation of new Go projects in the monorepo, ensuring consistency with established patterns and conventions.

## When to Use

Use this skill when the user requests:
- "Create a new Go project"
- "Add a new CLI tool"
- "Scaffold a new Go library"
- "Set up a new Go project for..."

## Project Types

This monorepo supports three types of Go projects:

1. **CLI Tool** (root-level): Standalone command-line applications
2. **Library** (lib/): Reusable packages without a main entry point
3. **Linter** (linters/): Code analysis tools

## Instructions

### Gather Information

Ask the user for:
1. **Project name** (lowercase, hyphen-separated, e.g., "my-tool")
2. **Project type** (cli, library, or linter)
3. **Brief description** (one sentence describing what it does)
4. **Author name** (for documentation, optional)

### Determine Project Location

Based on project type:
- **CLI Tool**: `./{project-name}/`
- **Library**: `./lib/{project-name}/`
- **Linter**: `./linters/{project-name}/`

### Create Directory Structure

**IMPORTANT**: This is a single-module monorepo. Projects do NOT have individual go.mod files - all projects share the root `go.mod`.

For **CLI Tool**:
```
{project-name}/
├── main.go
├── README.md
├── AGENTS.md
└── .gitignore
```

For **Library**:
```
lib/{project-name}/
├── README.md
└── {project-name}.go
```

For **Linter**:
```
linters/{project-name}/
├── main.go
├── README.md
└── AGENTS.md
```

### Create main.go (CLI Tools and Linters Only)

Use this minimal template:
```go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: {project-name} <command>\n")
		os.Exit(1)
	}

	fmt.Printf("Hello from {project-name}!\n")
}
```

For projects that need a CLI framework, suggest using `github.com/spf13/cobra` and reference the `aihook` project as an example.

### Create README.md

Template structure:
```markdown
# {project-name}

{Brief description from user}

## Overview

{Expanded description of what the project does and why it exists}

## Features

- Feature 1
- Feature 2
- Feature 3

## Installation

### From Source

\`\`\`bash
go install github.com/neongreen/mono/{path-to-project}@latest
\`\`\`

### Local Development

\`\`\`bash
go build -o /tmp/{project-name} ./{project-name}
\`\`\`

## Usage

\`\`\`bash
{project-name} [options] <command>
\`\`\`

### Examples

\`\`\`bash
# Example 1
{project-name} example-command

# Example 2
{project-name} --flag value
\`\`\`

## Development

### Running Tests

\`\`\`bash
go test ./{project-name}/... -v
\`\`\`

### Building

\`\`\`bash
go build -o /tmp/{project-name} ./{project-name}
\`\`\`

## License

See the root LICENSE file in the repository.
```

For **libraries**, omit the "Installation" and "Usage" sections and instead include an "API" section with example code showing how to use the library.

### Create AGENTS.md (CLI Tools and Linters Only)

Template:
```markdown
# Agent Guidelines for {project-name}

## Project Overview

{Brief description of what the project does}

## Project Structure

\`\`\`
{project-name}/
├── main.go                    # CLI entry point
├── README.md                  # User documentation
└── AGENTS.md                  # This file
\`\`\`

## Development Guidelines

### Building and Testing

Always run tests and build from the repository root:

\`\`\`bash
# Run tests
go test ./{project-name}/...

# Build
go build -o /tmp/{project-name} ./{project-name}

# Run
/tmp/{project-name} [args...]
\`\`\`

### Code Structure

- `main.go` - CLI entry point and main logic
- Keep it simple initially - add packages as complexity grows

## Dependencies

All dependencies are managed via the repository's root go.mod file.
```

### Create .gitignore (CLI Tools Only)

Template:
```
# Binaries
{project-name}
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test artifacts
*.test
*.out
coverage.txt

# IDE files
.vscode/
.idea/
```

### Update Dependencies

Run from the repository root to update the shared go.mod/go.sum:
```bash
go mod tidy
```

This updates the root `go.mod` and `go.sum` files with any new dependencies.

### Verify the Project Works

Build and test:
```bash
# Build
go build ./{path-to-project}

# Run (CLI tools only)
./{project-name} --help

# Test
go test ./{path-to-project}/...
```

### Add to go-projects.toml (CRITICAL)

**This step is REQUIRED for the project to be properly integrated into the monorepo.**

Add an entry to `go-projects.toml`:

```toml
[[project]]
dir = "{project-dir}"           # e.g., "my-tool" or "lib/my-lib"
module = "{full-module-path}"   # REQUIRED: Always use the full module path, e.g., "github.com/neongreen/mono/my-tool"
```

**Important:** Always use the full module path for the `module` field (e.g., `github.com/neongreen/mono/my-tool`). Some existing projects use short names (e.g., `aihook`, `conf`) for historical reasons, but new projects should always use the full path for consistency.

Examples:
- CLI Tool: `dir = "my-tool"`, `module = "github.com/neongreen/mono/my-tool"`
- Library: `dir = "lib/my-lib"`, `module = "github.com/neongreen/mono/lib/my-lib"`
- Linter: `dir = "linters/my-linter"`, `module = "github.com/neongreen/mono/linters/my-linter"`

This file is used by various tools including the release workflow to detect Go projects.

### Dagger Integration

**See @DAGGER.md for complete Dagger setup instructions.**

Summary:
- Add project to `dagger.json` include list
- Create Dagger project file at `.dagger/project_{name}.go`
- Verify with `dagger call project {project-name} build`

### Add to release-mirror.toml (Optional - For Homebrew)

**Only if the project should be published to Homebrew**, add an entry to `release-mirror.toml`:

```toml
[[mirror.projects]]
name = "{project-name}"
desc = "{Brief description for Homebrew}"
homepage = "https://github.com/neongreen/mono/tree/main/{project-dir}"
binary = "{binary-name}"
test_args = ["--help"]
```

This is optional and only needed for projects that should have a Homebrew formula.

### Verify Integration

Run these commands to verify everything is set up correctly:

```bash
# Verify go.mod is valid
go mod tidy

# Verify the project appears in go-projects.toml
grep -A1 "{project-name}" go-projects.toml
```

For Dagger verification, see @DAGGER.md.

### Create Task (Optional but Recommended)

Use `tk` to create a task tracking this new project:
```bash
tk create "Add {project-name} project" --description "Created new Go {type} project: {brief description}"
```

## Output Format

After creating the project, provide a summary:

```markdown
✅ Created new {type} project: {project-name}

Location: {full-path}

Files created:
- main.go (or {project-name}.go for libraries)
- README.md
- AGENTS.md (for CLI/linter)
- .gitignore (for CLI)

Integration files updated:
- go.mod (root - updated dependencies)
- go-projects.toml (added project entry)
- dagger.json (added to include list)
- .dagger/project_{name}.go (created Dagger build configuration)
- release-mirror.toml (if Homebrew release requested)

Next steps:
1. Implement the core functionality
2. Add tests
3. Update README.md with actual usage examples
4. Test the Dagger build: `dagger call project {project-name} build`
5. Consider adding:
   - GitHub workflow (.github/workflows/{project-name}.yml) for project-specific CI
   - More comprehensive documentation
   - Integration tests
```

## Common Dependencies

Suggest these dependencies based on project needs:

- **CLI framework**: `github.com/spf13/cobra`
- **Version command**: `github.com/neongreen/mono/lib/version`
- **Configuration**: `github.com/neongreen/mono/lib/configschema`
- **Path language**: `github.com/neongreen/mono/lib/pathlang`
- **GitHub API**: `github.com/neongreen/mono/lib/ghclient`
- **Testing**: Standard `testing` package (no external framework needed)

## Examples from Existing Projects

Reference these as templates:

- **Simple CLI**: `aihook`, `jj-run`, `printpdf`
- **Complex CLI**: `tk`, `conf`, `want`
- **Library**: `lib/pathlang`, `lib/setlang`, `lib/version`
- **Linter**: `linters/cobralint`, `linters/uselesswrapper`

## Best Practices

1. **Start simple**: Begin with minimal structure, add complexity as needed
2. **Follow conventions**: Use existing projects as references
3. **Test early**: Add tests from the start
4. **Document as you go**: Keep README.md and AGENTS.md up to date
5. **Use local libraries**: Leverage shared libraries in `lib/` for common functionality
6. **Consistent style**: Match the coding style of existing projects
7. **Build verification**: Always verify the project builds before considering it complete

## Troubleshooting

### Issue: "package not found"
**Solution**: Run `go mod tidy` from the repository root

### Issue: "import cycle"
**Solution**: Check if you're creating circular dependencies between packages

### Issue: "go.sum mismatch"
**Solution**: Delete go.sum and run `go mod tidy` again

### Issue: Library not found in root go.mod
**Solution**: Add the replace directive and require statement, then run `go mod tidy`
